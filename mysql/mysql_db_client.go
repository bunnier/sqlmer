package mysql

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

	mysqlDriver "github.com/go-sql-driver/mysql"
)

// DriverName 是 MySql 驱动名称。
const DriverName = "mysql"

var _ sqlmer.DbClient = (*MySqlDbClient)(nil)

// MySqlDbClient 是针对 MySql 的 DbClient 实现。
type MySqlDbClient struct {
	*sqlmer.AbstractDbClient
	dsnConfig *mysqlDriver.Config
}

// NewMySqlDbClient 用于创建一个 MySqlDbClient。
func NewMySqlDbClient(dsn string, options ...sqlmer.DbClientOption) (*MySqlDbClient, error) {
	var dsnConfig *mysqlDriver.Config
	var err error

	if dsnConfig, err = mysqlDriver.ParseDSN(dsn); err != nil {
		return nil, err
	}

	// 影响行数的处理，采用和 SQL Server 一样的逻辑，即：
	// UPDATE 时，只要找到行，即使原值和目标值一致，也作为影响到行处理。
	dsnConfig.ClientFoundRows = true
	dsn = dsnConfig.FormatDSN()

	fixedOptions := []sqlmer.DbClientOption{
		sqlmer.WithDsn(DriverName, dsn),
		sqlmer.WithGetScanTypeFunc(getScanTypeFn(dsnConfig)),        // 定制 Scan 类型逻辑。
		sqlmer.WithUnifyDataTypeFunc(getUnifyDataTypeFn(dsnConfig)), // 定制类型转换逻辑。
		sqlmer.WithBindArgsFunc(bindArgs),                           // 定制参数绑定逻辑。
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

	// 自动设置连接池的超时时间，
	err = autoSetConnMaxLifetime(absDbClient.Db)
	if err != nil {
		return nil, err
	}

	return &MySqlDbClient{absDbClient, dsnConfig}, nil
}

// autoSetConnMaxLifetime 用于从数据库读取 wait_timeout 设置，并自动设置连接最大生命周期。
// 当数据库的超时时间有容错时间时，连接的使用时间短一些，，以尽量在数据库掐断连接之前主动放弃连接。
func autoSetConnMaxLifetime(db *sqlen.DbEnhance) error {
	timeSettingRows := db.EnhancedQueryRow(`SHOW VARIABLES WHERE Variable_name IN ('wait_timeout')`) // 连接测试。
	if timeSettingRows.Err() != nil {
		return fmt.Errorf("%w: get 'wait_timeout' variable error: %w", sqlmer.ErrConnect, timeSettingRows.Err())
	}

	var waitTimeoutName string
	var waitTimeout int
	err := timeSettingRows.Scan(&waitTimeoutName, &waitTimeout)
	if err != nil {
		return fmt.Errorf("%w: get 'wait_timeout' variable error: %w", sqlmer.ErrConnect, err)
	}

	// 根据 wait_timeout 设置连接最大生命周期。
	// SetConnMaxLifetime 小于 wait_timeout 应该是最佳实践，可以避免数据库强制关闭连接导致的错误（driver: bad connection）。
	// 参考：
	// https://go.dev/doc/database/manage-connections
	// https://github.com/go-sql-driver/mysql?tab=readme-ov-file#important-settings
	switch {
	case waitTimeout == 0:
		return nil

	case waitTimeout > 60:
		waitTimeout = waitTimeout - 5

	case waitTimeout > 1:
		waitTimeout = waitTimeout - 1
	}

	db.SetConnMaxLifetime(time.Duration(waitTimeout) * time.Second)
	return nil
}

// bindArgs 用于对 SQL 语句和参数进行预处理。
// 第一个参数如果是 map，且仅且只有一个参数的情况下，做命名参数处理，其余情况做位置参数处理。
func bindArgs(sqlText string, args ...any) (string, []any, error) {
	namedParsedResult := parseMySqlNamedSql(sqlText)
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

// 用于缓存 parseMySqlNamedSql 解析的结果。
var mysqlNamedSqlParsedResult sync.Map

type mysqlNamedParsedResult struct {
	Sql   string   // 处理后的sql语句。
	Names []string // 原语句中用到的命名参数集合 (没有去重逻辑如果有同名参数，会有多项)。
}

// 用于初始化合法字符集合 map，用于快速筛选合法字符。
var onceInitParamNameMap = sync.Once{}

// 定义 mysql 参数名允许的字符。
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

// 分析Sql语句，提取用到的命名参数名称（按顺序），并将 @ 占位参数转换为 mysql 驱动支持的 ? 形式。
func parseMySqlNamedSql(sqlText string) mysqlNamedParsedResult {
	// 如果缓存中有数据，直接返回。
	if cacheResult, ok := mysqlNamedSqlParsedResult.Load(sqlText); ok {
		return cacheResult.(mysqlNamedParsedResult)
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

	parsedResult := mysqlNamedParsedResult{fixedSqlTextBuilder.String(), names}
	mysqlNamedSqlParsedResult.Store(sqlText, parsedResult) // 缓存结果。
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
func getScanTypeFn(cfg *mysqlDriver.Config) sqlen.GetScanTypeFunc {
	var scanTypeRawBytes = reflect.TypeOf(sql.RawBytes{})
	return func(columnType *sql.ColumnType) reflect.Type {
		if !cfg.ParseTime && isTimeColumn(columnType.DatabaseTypeName()) {
			return scanTypeRawBytes // MySql 的驱动，如果没有开启 ParseTime，只能通过 sql.RawBytes 进行 Scan，否则会失败。
		}
		return columnType.ScanType()
	}
}

// isTimeColumn 判断某个列是否是时间类型。
func isTimeColumn(colTypeName string) bool {
	return colTypeName == "TIMESTAMP" ||
		colTypeName == "TIME" ||
		colTypeName == "DATETIME" ||
		colTypeName == "DATE"
}

// getUnifyDataTypeFn 根据驱动配置返回一个统一处理数据类型的函数。
func getUnifyDataTypeFn(cfg *mysqlDriver.Config) sqlen.UnifyDataTypeFn {
	return func(columnType *sql.ColumnType, dest *any) {
		switch columnType.DatabaseTypeName() {
		case "VARCHAR", "CHAR", "TEXT", "DECIMAL":
			switch v := (*dest).(type) {
			case sql.RawBytes:
				if v == nil {
					*dest = nil
					break
				}
				*dest = string(v)
			}

		case "TIMESTAMP", "TIME", "DATETIME", "DATE":
			if cfg.ParseTime {
				break // 如果驱动开启了 ParseTime，就不需要再进行转换了。
			}
			switch v := (*dest).(type) {
			case sql.RawBytes:
				if v == nil {
					*dest = nil
					break
				}
				// 用 RawBytes 接收的 Time 系列类型，需要转为 time.Time{}，下面这个时间格式是从 MySQL 驱动里拷来的。
				const timeFormat = "2006-01-02 15:04:05.999999"
				timeStr := string(v)
				if time, err := time.ParseInLocation(timeFormat[:len(timeStr)], timeStr, cfg.Loc); err != nil {
					panic(err)
				} else {
					*dest = time
				}
			}

		default: // 将 sql.RawBytes 统一转为 []byte。
			switch v := (*dest).(type) {
			case sql.RawBytes:
				if v == nil {
					*dest = nil
					break
				}
				*dest = []byte(v)
			case *any:
				if v == nil { // 直接 SELECT null 时候会使用 *any 进行 Scan 并在最后返回一个指向 nil 的值，这里将其统一到无类型 nil。
					*dest = nil
					break
				}
				*dest = v
			}

		}
	}
}
