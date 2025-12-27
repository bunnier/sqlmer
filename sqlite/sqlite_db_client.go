package sqlite

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/sqlen"

	_ "modernc.org/sqlite"
)

// DriverName 是 Sqlite 驱动名称。
const DriverName = "sqlite"

var _ sqlmer.DbClient = (*SqliteDbClient)(nil)

// SqliteDbClient 是针对 Sqlite 的 DbClient 实现。
type SqliteDbClient struct {
	*sqlmer.AbstractDbClient
}

// NewSqliteDbClient 用于创建一个 SqliteDbClient。
func NewSqliteDbClient(dsn string, options ...sqlmer.DbClientOption) (*SqliteDbClient, error) {
	fixedOptions := []sqlmer.DbClientOption{
		sqlmer.WithDsn(DriverName, dsn),
		sqlmer.WithGetScanTypeFunc(getScanTypeFn()),        // 定制 Scan 类型逻辑。
		sqlmer.WithUnifyDataTypeFunc(getUnifyDataTypeFn()), // 定制类型转换逻辑。
		sqlmer.WithBindArgsFunc(bindArgs),                  // 定制参数绑定逻辑。
	}
	options = append(fixedOptions, options...) // 用户自定义选项放后面，以覆盖默认。

	config, err := sqlmer.NewDbClientConfig(options...)
	if err != nil {
		return nil, err
	}

	absDbClient, err := sqlmer.NewAbstractDbClient(config)
	if err != nil {
		return nil, err
	}

	return &SqliteDbClient{absDbClient}, nil
}

// bindArgs 用于对 SQL 语句和参数进行预处理。
// 第一个参数如果是 map，且仅且只有一个参数的情况下，做命名参数处理，其余情况做位置参数处理。
func bindArgs(sqlText string, args ...any) (string, []any, error) {
	namedParsedResult := parseSqliteNamedSql(sqlText)
	parsedNames := len(namedParsedResult.Names)
	argsCount := len(args)
	resultArgs := make([]any, 0, parsedNames)

	// map 按返回的 paramNames 顺序整理一个 slice 返回。
	if argsCount == 1 && reflect.ValueOf(args[0]).Kind() == reflect.Map {
		mapArgs := args[0].(map[string]any)
		for _, paramName := range namedParsedResult.Names { // 这里通过从 SQL 解析出来的参数名去取参数，刚好也解决了参数重复使用的问题。
			if value, ok := mapArgs[paramName]; ok {
				resultArgs = append(resultArgs, value)
			} else {
				return "", nil, fmt.Errorf("%w:\nlack of parameter\nsql = %s", sqlmer.ErrParseParamFailed, namedParsedResult.Sql)
			}
		}
		namedParsedResult.Sql, resultArgs = extendInParams(namedParsedResult.Sql, resultArgs)
		return namedParsedResult.Sql, resultArgs, nil
	}

	// slice 语句中使用的顺序未必是递增的，切可能重复引用同一个索引，所以这里也需要整理顺序。
	for _, paramName := range namedParsedResult.Names {
		if paramName[0] != 'p' { // 要求数字参数的格式为 @p1....@pN
			return "", nil, fmt.Errorf("%w: parsing parameter failed\nsql = %s", sqlmer.ErrParseParamFailed, namedParsedResult.Sql)
		}

		index, err := strconv.Atoi(paramName[1:]) // 取出索引值。
		if err != nil {
			return "", nil, fmt.Errorf("%w: parsing parameter failed\nsql = %s", sqlmer.ErrParseParamFailed, namedParsedResult.Sql)
		}

		index-- // 占位符从 1 开始，转成索引，需要 -1。

		if index < 0 { // 数字参数的数值从 1 开始，而不是 0，因此不会是负数。
			return "", nil, fmt.Errorf("%w: index must start at 1\nsql = %s", sqlmer.ErrParseParamFailed, namedParsedResult.Sql) // 索引对不上参数。
		}

		if index >= argsCount { // 索引值超过参数数量了。
			return "", nil, fmt.Errorf("%w: parameter index out of args range\nsql = %s", sqlmer.ErrParseParamFailed, namedParsedResult.Sql)
		}

		resultArgs = append(resultArgs, args[index])
	}

	namedParsedResult.Sql, resultArgs = extendInParams(namedParsedResult.Sql, resultArgs)
	return namedParsedResult.Sql, resultArgs, nil
}

// 用于缓存 parseSqliteNamedSql 解析的结果。
var sqliteNamedSqlParsedResult sync.Map

type sqliteNamedParsedResult struct {
	Sql   string   // 处理后的sql语句。
	Names []string // 原语句中用到的命名参数集合 (没有去重逻辑如果有同名参数，会有多项)。
}

// 用于初始化合法字符集合 map，用于快速筛选合法字符。
var onceInitParamNameMap = sync.Once{}

// 定义 sqlite 参数名允许的字符。
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

// 分析Sql语句，提取用到的命名参数名称（按顺序），并将 @ 占位参数转换为 sqlite 驱动支持的 ? 形式。
func parseSqliteNamedSql(sqlText string) sqliteNamedParsedResult {
	// 如果缓存中有数据，直接返回。
	if cacheResult, ok := sqliteNamedSqlParsedResult.Load(sqlText); ok {
		return cacheResult.(sqliteNamedParsedResult)
	}

	names := make([]string, 0, 10) // 存放 sql 中所有的参数名称。

	fixedSqlTextBuilder := strings.Builder{}
	paramNameBuilder := strings.Builder{}

	inName := false   // 标示当前字符是否正处于参数名称之中。
	inString := false // 标示当前字符是否正处于字符串之中。

	lastIndex := utf8.RuneCountInString(sqlText) - 1 // sql 语句 bytes 的最后一个索引位置。

	for i, currentRune := range sqlText {
		switch {
		// 遇到字符串开始结束符号'需要设置当前字符串状态。
		case currentRune == '\'':
			inString = !inString
			fixedSqlTextBuilder.WriteRune(currentRune)

		// 如果当前状态正处于字符串中，字符无须特殊处理。
		case inString:
			fixedSqlTextBuilder.WriteRune(currentRune)

		// @ 符号标示参数名称部分开始。
		case currentRune == '@':
			// 连续2个@可以用来转义，需要跳出作用域。
			if inName && i > 0 && sqlText[i-1] == '@' {
				fixedSqlTextBuilder.WriteRune(currentRune)
				inName = false
				continue
			}
			inName = true
			paramNameBuilder.Reset()

		// 当前状态正处于参数名中，且当前字符合法，应作为参数名的一部分。
		case inName && isLegalParamNameCharter(currentRune):
			paramNameBuilder.WriteRune(currentRune)
			if lastIndex == i { // 如果是最后一个字符，直接写入，并结束。
				fixedSqlTextBuilder.WriteString("?")
				names = append(names, paramNameBuilder.String())
			}

		// 当前状态正处于参数名中，且当前字符不是合法的参数字符，标示着参数名范围结束。
		case inName && !isLegalParamNameCharter(currentRune):
			inName = false
			fixedSqlTextBuilder.WriteString("?")
			fixedSqlTextBuilder.WriteRune(currentRune)
			names = append(names, paramNameBuilder.String())

		// 非上述情况即为普通 sql 字符部分，无须特殊处理。
		default:
			fixedSqlTextBuilder.WriteRune(currentRune)
		}
	}

	parsedResult := sqliteNamedParsedResult{fixedSqlTextBuilder.String(), names}
	sqliteNamedSqlParsedResult.Store(sqlText, parsedResult) // 缓存结果。
	return parsedResult
}

// extendInParams 用于处理 SQL IN 子句的参数展开
// 将切片类型的参数展开为多个问号占位符
func extendInParams(sqlText string, params []any) (string, []any) {
	var newParams []any = make([]any, 0, len(params))
	var newSqlBuilder strings.Builder

	paramIndex := 0
	inString := false
	for _, r := range sqlText {
		if r == '\'' {
			inString = !inString
			newSqlBuilder.WriteRune(r)
			continue
		}

		if inString {
			newSqlBuilder.WriteRune(r)
			continue
		}

		if r != '?' {
			newSqlBuilder.WriteRune(r)
			continue
		}

		if paramIndex >= len(params) { // 后面没参数了，无需判断了。
			newSqlBuilder.WriteRune(r)
			continue
		}

		param := params[paramIndex]
		paramValue := reflect.ValueOf(param)
		paramIndex++

		// 处理切片类型。
		// 排除 []byte，因为虽然 []byte 也是切片类型，但是它是二进制数据，不应该被展开。
		if (paramValue.Kind() == reflect.Slice || paramValue.Kind() == reflect.Array) && !paramValue.Type().ConvertibleTo(reflect.TypeOf([]byte{})) {
			paramLen := paramValue.Len()
			if paramLen == 0 {
				// 空切片替换为 SQL 不可能条件。
				newSqlBuilder.WriteString("NULL")
				continue
			}

			// 生成占位符。
			placeholders := strings.Repeat(",?", paramLen)
			newSqlBuilder.WriteString(placeholders[1:])

			// 展开参数。
			for i := 0; i < paramLen; i++ {
				newParams = append(newParams, paramValue.Index(i).Interface())
			}
		} else {
			newSqlBuilder.WriteRune('?')
			newParams = append(newParams, param)
		}
	}

	return newSqlBuilder.String(), newParams
}

// getScanTypeFn 根据驱动配置返回一个可以正确获取 Scan 类型的函数。
func getScanTypeFn() sqlen.GetScanTypeFunc {
	// sqlite 的 ScanType 可能不够准确，这里尽量使用通用的处理。
	// modernc.org/sqlite 可能会返回 nil 或者 interface{} 如果类型未知。
	return func(columnType *sql.ColumnType) reflect.Type {
		// 如果 ScanType 返回 nil，使用 interface{} 或者 sql.RawBytes。
		// 使用 interface{} 可以让驱动决定返回什么类型。
		t := columnType.ScanType()
		if t == nil {
			return reflect.TypeOf(new(any)).Elem()
		}
		return t
	}
}

// getUnifyDataTypeFn 根据驱动配置返回一个统一处理数据类型的函数。
func getUnifyDataTypeFn() sqlen.UnifyDataTypeFn {
	return func(columnType *sql.ColumnType, dest *any) {
		// 获取数据库类型名称，注意 SQLite 类型可能是 "VARCHAR(20)" 这种。
		typeName := strings.ToUpper(columnType.DatabaseTypeName())

		// 处理字符串类型。
		if strings.Contains(typeName, "CHAR") || strings.Contains(typeName, "TEXT") || strings.Contains(typeName, "CLOB") {
			switch v := (*dest).(type) {
			case sql.RawBytes:
				if v == nil {
					*dest = nil
					break
				}
				*dest = string(v)
			case []byte: // 某些情况下可能是 []byte
				if v == nil {
					*dest = nil
					break
				}
				*dest = string(v)
			case string:
				// 已经是 string 了，不需要处理
			}
			return
		}

		// 处理时间类型。
		// SQLite 默认没有时间类型，通常存为 TEXT 或 REAL 或 INTEGER。
		// 如果定义表时使用了 DATETIME 等类型，driver 可能会有所提示。
		// modernc.org/sqlite 对 DATETIME 类型好像会返回 string。
		if strings.Contains(typeName, "TIME") || strings.Contains(typeName, "DATE") {
			switch v := (*dest).(type) {
			case string:
				// 尝试解析常见时间格式。
				// 这里假设是标准格式，如果不是，可能需要更多逻辑。
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					*dest = t
				} else if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
					*dest = t
				} else if t, err := time.Parse("2006-01-02 15:04:05.999999999", v); err == nil {
					*dest = t
				} else if t, err := time.Parse("2006-01-02", v); err == nil {
					*dest = t
				} else if t, err := time.Parse("15:04:05.999999999", v); err == nil {
					// Time only
					// Usually we return time.Time with date 0000-01-01 or similar?
					// Use a base date.
					baseDate, _ := time.Parse("2006-01-02", "0001-01-01")
					*dest = baseDate.Add(time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute + time.Duration(t.Second())*time.Second + time.Duration(t.Nanosecond()))
				}
				// 否则保留 string。
			case int64:
				// 假设是 Unix timestamp。
				*dest = time.Unix(v, 0)
			case []byte:
				s := string(v)
				if t, err := time.Parse(time.RFC3339, s); err == nil {
					*dest = t
				} else if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
					*dest = t
				} else if t, err := time.Parse("2006-01-02 15:04:05.999999999", s); err == nil {
					*dest = t
				}
			}
			return
		}

		// 默认处理，处理 sql.RawBytes。
		switch v := (*dest).(type) {
		case sql.RawBytes:
			if v == nil {
				*dest = nil
				break
			}
			*dest = []byte(v)
		case *any:
			if v == nil {
				*dest = nil
				break
			}
			*dest = v
		}
	}
}
