package mssql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/bunnier/sqlmer"

	_ "github.com/denisenkom/go-mssqldb"
)

// DriverName 是 SqlServer 驱动名称。
const DriverName = "sqlserver"

var _ sqlmer.DbClient = (*MsSqlDbClient)(nil)

// MsSqlDbClient 是针对 SqlServer 的 DbClient 实现。
type MsSqlDbClient struct {
	*sqlmer.AbstractDbClient
}

// NewMsSqlDbClient 用于创建一个 MsSqlDbClient。
func NewMsSqlDbClient(dsn string, options ...sqlmer.DbClientOption) (*MsSqlDbClient, error) {
	fixedOptions := []sqlmer.DbClientOption{
		sqlmer.WithDsn(DriverName, dsn),
		sqlmer.WithUnifyDataTypeFunc(unifyDataType),
		sqlmer.WithBindArgsFunc(bindArgs), // SqlServer 要支持命名参数，需要定制一个参数解析函数。
	}
	options = append(fixedOptions, options...) // 用户自定义选项放后面，以覆盖默认。

	config, err := sqlmer.NewDbClientConfig(options...)
	if err != nil {
		return nil, err
	}

	internalDbClient, err := sqlmer.NewAbstractDbClient(config)
	if err != nil {
		return nil, err
	}

	return &MsSqlDbClient{internalDbClient}, nil
}

// unifyDataType 用于统一数据类型。
func unifyDataType(columnType *sql.ColumnType, dest *any) {
	switch columnType.DatabaseTypeName() {
	case "DECIMAL", "SMALLMONEY", "MONEY":
		switch v := (*dest).(type) {
		case []byte:
			if v == nil {
				*dest = nil
				break
			}
			*dest = string(v)
		case *string:
			*dest = v
		}

	case "VARBINARY", "BINARY":
		switch v := (*dest).(type) {
		case []byte:
			if v == nil { // 将 nil 的切片转为无类型 nil。
				*dest = nil
				break
			}
		}
	}
}

// bindArgs 用于对 sql 语句和参数进行预处理。
// 第一个参数如果是 map，且仅且只有一个参数的情况下，做命名参数处理；其余情况做位置参数处理。
func bindArgs(sqlText string, args ...any) (string, []any, error) {
	var params map[string]any

	if len(args) == 0 {
		return sqlText, args, nil
	}

	// 多个参数的情况，必然是索引参数，反之则为命名参数。
	if len(args) > 1 || reflect.TypeOf(args[0]).Kind() != reflect.Map {
		// 下面循环用于判断参数中是否有需要展开的参数。
		var hasSliceParam bool
		for i := 0; i < len(args); i++ {
			paramValue := reflect.ValueOf(args[i])
			if (paramValue.Kind() == reflect.Slice || paramValue.Kind() == reflect.Array) && !paramValue.Type().ConvertibleTo(reflect.TypeOf([]byte{})) {
				hasSliceParam = true
				break
			}
		}

		// 如果没有需要展开的参数，索引参数可以直接交给驱动执行。
		if !hasSliceParam {
			return sqlText, args, nil
		}

		// 有需要展开的参数，将所有参数转为 map 类型，走下面命名参数的逻辑。
		params = make(map[string]any, len(args))
		for i := 0; i < len(args); i++ {
			params[fmt.Sprintf("p%d", i+1)] = args[i] // 索引从 1 开始。
		}
	} else {
		params = args[0].(map[string]any) // 本来就是命名参数传参。
	}

	return extendInParams(sqlText, params)
}

// extendInParams 用于处理 SQL IN 子句的参数展开
// 将切片类型的参数展开为多个参数。
func extendInParams(sqlText string, params map[string]any) (string, []any, error) {
	var newParams []any = make([]any, 0, len(params))
	var newSqlBuilder strings.Builder

	inString := false
	inName := false
	paramNameBuilder := strings.Builder{}
	lastIndex := utf8.RuneCountInString(sqlText) - 1 // sql 语句 bytes 的最后一个索引位置。

	hasParam := map[string]struct{}{} // 用于判断某个参数是否已经存在返回结果的参数列表中

	for i, r := range sqlText {
		if r == '\'' {
			inString = !inString
			newSqlBuilder.WriteRune(r)
			continue
		}

		if inString {
			newSqlBuilder.WriteRune(r)
			continue
		}

		if !inName && !strings.ContainsRune("@", r) {
			newSqlBuilder.WriteRune(r)
			continue
		}

		// 标识参数的开始。
		if r == '@' {
			// 连续2个@可以用来转义，需要跳出作用域。
			if inName && i > 0 && sqlText[i-1] == '@' {
				paramNameBuilder.WriteRune(r)
				inName = false
				continue
			}
			inName = true
			paramNameBuilder.Reset()
			continue
		}

		// 当前状态正处于参数名中，且当前字符不是合法的参数字符，标示着参数名范围结束。
		// 只要本字符不是最后一个字符，都会有后续的字符触发参数处理逻辑，都可以继续，但如果是最后一个字符，需要直接开始处理参数逻辑。
		if inName && isLegalParamNameCharter(r) {
			paramNameBuilder.WriteRune(r)
			if i < lastIndex {
				continue
			}
		}

		inName = isLegalParamNameCharter(r) // 如果最后一个字符是合法字符还走到这里，说明本字符其实是参数名的最后一个字。

		paramName := paramNameBuilder.String()

		// 走到这里就是一个参数名结束，需要开始处理参数的时候了。
		param := params[paramName]

		// 处理切片类型。
		// 排除 []byte，因为虽然 []byte 也是切片类型，但是它是二进制数据，不应该被展开。
		paramValue := reflect.ValueOf(param)
		if (paramValue.Kind() == reflect.Slice || paramValue.Kind() == reflect.Array) && !paramValue.Type().ConvertibleTo(reflect.TypeOf([]byte{})) {
			paramLen := paramValue.Len()
			if paramLen == 0 {
				// 空切片替换为 SQL 不可能条件。
				newSqlBuilder.WriteString("NULL")
				continue
			}

			_, hasThisParam := hasParam[paramName]
			for i := 0; i < paramLen; i++ {
				if i != 0 {
					newSqlBuilder.WriteByte(',')
				}

				// 生成占位符。@xxx 将被展开为 @xxx_0, @xxx_1, @xxx_2...
				newSqlBuilder.WriteByte('@')
				paramName := fmt.Sprintf("%s_%d", paramName, i)
				newSqlBuilder.WriteString(paramName)

				if hasThisParam { // 表明本参数已经在列表里面了，不需要再次添加。
					continue
				}

				// 插入参数。
				newParams = append(newParams, sql.Named(paramName, paramValue.Index(i).Interface()))
			}
		} else {
			newSqlBuilder.WriteRune('@')
			newSqlBuilder.WriteString(paramName)

			if _, ok := hasParam[paramName]; ok {
				continue
			}

			newParams = append(newParams, sql.Named(paramName, param))
		}

		// 标记为本参数已在列表中，且需要把触发参数结束的字符写入。
		hasParam[paramName] = struct{}{}

		if !inName { // 如果上面是因为最后一个字符触发的参数逻辑，这里的字符就是参数名的最后一个字符，不用再写入。
			newSqlBuilder.WriteRune(r)
		}
	}

	return newSqlBuilder.String(), newParams, nil
}

// 用于初始化合法字符集合 map，用于快速筛选合法字符。
var onceInitParamNameMap = sync.Once{}

// 定义 mssql 参数名允许的字符。
var legalParamNameCharactersMap map[rune]struct{}

// 用于快速判断某个字符是否是占位符参数合法字符。
func isLegalParamNameCharter(r rune) bool {
	onceInitParamNameMap.Do(func() {
		const legalParamNameCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
		legalParamNameCharactersMap = make(map[rune]struct{}, len(legalParamNameCharacters))
		for _, r := range legalParamNameCharacters {
			legalParamNameCharactersMap[r] = struct{}{}
		}
	})

	_, ok := legalParamNameCharactersMap[r]
	return ok
}
