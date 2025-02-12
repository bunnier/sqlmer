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

	inString := false                                // 用于判断当前字符是否在字符串中。
	inName := false                                  // 用于判断当前字符是否属于参数名的一部分。
	paramNameBuilder := strings.Builder{}            // 用于存储参数名。
	lastIndex := utf8.RuneCountInString(sqlText) - 1 // sql 语句 bytes 的最后一个索引位置。
	hasParam := map[string]struct{}{}                // 用于判断某个参数是否已经存在返回结果的参数列表中。

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

		// 处理参数标识符。
		if r == '@' {
			// 处理转义的 @@ 情况。
			if inName && i > 0 && sqlText[i-1] == '@' {
				paramNameBuilder.WriteRune(r)
				inName = false
				continue
			}

			// 开始新的参数名
			inName = true
			paramNameBuilder.Reset()
			continue
		}

		// 普通语句字符，直接入库即可。
		if !inName {
			newSqlBuilder.WriteRune(r)
			continue
		}

		// 处理参数名称字符。
		isLegalChar := isLegalParamNameCharter(r)
		if isLegalChar {
			paramNameBuilder.WriteRune(r)
			if i < lastIndex {
				continue
			}
		}

		// 参数名称结束，处理参数。
		inName = isLegalChar // 更新参数名状态。
		paramName := paramNameBuilder.String()
		param := params[paramName]
		paramValue := reflect.ValueOf(param)

		// 处理需要展开的参数。
		_, hasThisParam := hasParam[paramName]
		if (paramValue.Kind() == reflect.Slice || paramValue.Kind() == reflect.Array) &&
			!paramValue.Type().ConvertibleTo(reflect.TypeOf([]byte{})) {
			paramLen := paramValue.Len()
			if paramLen == 0 {
				newSqlBuilder.WriteString("NULL")
				continue
			}

			for j := 0; j < paramLen; j++ {
				if j != 0 {
					newSqlBuilder.WriteByte(',')
				}

				// 生成展开参数的参数占位符。
				newParamName := fmt.Sprintf("%s_%d", paramName, j)
				newSqlBuilder.WriteByte('@')
				newSqlBuilder.WriteString(newParamName)

				// 只在参数首次出现时添加到参数列表。
				if !hasThisParam {
					newParams = append(newParams, sql.Named(newParamName, paramValue.Index(j).Interface()))
				}
			}
		} else {
			// 处理非切片类型参数
			newSqlBuilder.WriteRune('@')
			newSqlBuilder.WriteString(paramName)

			// 只在参数首次出现时添加到参数列表。
			if !hasThisParam {
				newParams = append(newParams, sql.Named(paramName, param))
			}
		}

		// 更新参数状态并处理结束字符
		hasParam[paramName] = struct{}{}

		// 如果是参数名的下一个字符触发的参数处理逻辑，该字符需要入语句。
		if !inName {
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
