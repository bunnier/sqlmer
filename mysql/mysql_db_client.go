package mysql

import (
	"database/sql"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/bunnier/sqlmer"
	"github.com/pkg/errors"

	_ "github.com/go-sql-driver/mysql"
)

// DriverName 是 MySql 驱动名称。
const DriverName = "mysql"

var _ sqlmer.DbClient = (*MySqlDbClient)(nil)

// MySqlDbClient 是针对 MySql 的 DbClient 实现。
type MySqlDbClient struct {
	sqlmer.AbstractDbClient
}

// NewMySqlDbClient 用于创建一个 MySqlDbClient。
func NewMySqlDbClient(connectionString string, options ...sqlmer.DbClientOption) (*MySqlDbClient, error) {
	// 依赖的驱动中，只有连接字符串中设置了 parseTime=true 才会转换 Date / Datetime 类型到 time.Time，
	// 本库为了不同库中的类型一致，这里强制开启该设置。
	if !strings.Contains(connectionString, "parseTime") {
		if !strings.Contains(connectionString, "?") {
			connectionString += "?"
		} else {
			connectionString += "&"
		}
		connectionString += "parseTime=true"
	}

	options = append(options,
		sqlmer.WithConnectionString(DriverName, connectionString),
		sqlmer.WithUnifyDataTypeFunc(UnifyDataType),
		sqlmer.WithBindArgsFunc(bindMySqlArgs)) // mysql 的驱动不支持命名参数，这里需要进行处理。
	config, err := sqlmer.NewDbClientConfig(options...)
	if err != nil {
		return nil, err
	}

	internalDbClient, err := sqlmer.NewInternalDbClient(config)
	if err != nil {
		return nil, err
	}

	return &MySqlDbClient{
		internalDbClient,
	}, nil
}

// UnifyDataType 用于统一数据类型。
func UnifyDataType(colDbTypeName string, dest *interface{}) {
	switch colDbTypeName {
	case "VARCHAR", "DECIMAL":
		switch v := (*dest).(type) {
		case []byte:
			if v == nil {
				*dest = nil
			}
			*dest = string(v)
		case sql.RawBytes:
			if v == nil {
				*dest = nil
			}
			*dest = string(v)
		case *string:
			*dest = v
		case nil:
			*dest = nil
		}
	}
}

// bindMySqlArgs 用于对 sql 语句和参数进行预处理。
// 第一个参数如果是 map，且仅且只有一个参数的情况下，做命名参数处理，其余情况做位置参数处理。
func bindMySqlArgs(sqlText string, args ...interface{}) (string, []interface{}, error) {
	namedParsedResult := parseMySqlNamedSql(sqlText)
	paramNameCount := len(namedParsedResult.Names)
	argsCount := len(args)
	resultArgs := make([]interface{}, 0, paramNameCount)

	// map 按返回的paramNames顺序整理一个slice返回。
	if argsCount == 1 && reflect.ValueOf(args[0]).Kind() == reflect.Map {
		mapArgs := args[0].(map[string]interface{})
		for _, paramName := range namedParsedResult.Names {
			if value, ok := mapArgs[paramName]; ok {
				resultArgs = append(resultArgs, value)
			} else {
				return "", nil, errors.Wrap(sqlmer.ErrSql, "lack of parameter:"+namedParsedResult.Sql)
			}
		}
		return namedParsedResult.Sql, resultArgs, nil
	}

	// slice 语句中使用的顺序未必是递增的，所以这里也需要整理顺序。
	for _, paramName := range namedParsedResult.Names {
		// 从参数名称提取索引。
		if paramName[0] != 'p' {
			return "", nil, errors.Wrap(sqlmer.ErrSql, "parameter error:"+namedParsedResult.Sql)
		}
		index, err := strconv.Atoi(paramName[1:])
		if err != nil {
			return "", nil, errors.Wrap(sqlmer.ErrSql, "parameter error:"+namedParsedResult.Sql)
		}
		index-- // 占位符从0开始。
		if index < 0 || index > paramNameCount-1 {
			return "", nil, errors.Wrap(sqlmer.ErrSql, "lack of parameter:"+namedParsedResult.Sql) // 索引对不上参数。
		}

		if index >= argsCount {
			return "", nil, errors.Wrap(sqlmer.ErrSql, "parameter error:"+namedParsedResult.Sql)
		}

		resultArgs = append(resultArgs, args[index])
	}

	if paramNameCount > len(resultArgs) {
		return "", nil, errors.Wrap(sqlmer.ErrSql, "parameter error:"+namedParsedResult.Sql)
	}

	return namedParsedResult.Sql, resultArgs, nil
}

// 用于缓存 parseMySqlNamedSql 解析的结果。
var mysqlNamedSqlParsedResult sync.Map

type mysqlNamedParsedResult struct {
	Sql   string   // 处理后的sql语句。
	Names []string // 原语句中用到的命名参数。
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
func parseMySqlNamedSql(sqlText string) *mysqlNamedParsedResult {
	// 如果缓存中有数据，直接返回。
	if cacheResult, ok := mysqlNamedSqlParsedResult.Load(sqlText); ok {
		return cacheResult.(*mysqlNamedParsedResult)
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

	parsedResult := &mysqlNamedParsedResult{fixedSqlTextBuilder.String(), names}
	mysqlNamedSqlParsedResult.Store(sqlText, parsedResult) // 缓存结果。
	return parsedResult
}
