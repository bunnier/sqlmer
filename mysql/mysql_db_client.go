package mysql

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"

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

// 定义 mysql 参数名允许的字符。
var mysqlParamNameAllowRunes = []*unicode.RangeTable{unicode.Letter, unicode.Digit}

// 分析Sql语句，提取用到的命名参数名称（按顺序），并将 @ 占位参数转换为 mysql 驱动支持的 ? 形式。
func parseMySqlNamedSql(sqlText string) *mysqlNamedParsedResult {
	// 如果缓存中有数据，直接返回。
	if cacheResult, ok := mysqlNamedSqlParsedResult.Load(sqlText); ok {
		return cacheResult.(*mysqlNamedParsedResult)
	}

	sqlTextBytes := []byte(sqlText)                     // 原 sql 的 bytes
	fixedSqlBytes := make([]byte, 0, len(sqlTextBytes)) // 处理后 sql 的 bytes。
	names := make([]string, 0, 10)                      // sql 中所有的参数名称。

	var name []byte                    // 存放解析过程中的参数名称。
	inName := false                    // 标示当前字符是否正处于参数名称之中。
	inString := false                  // 标示当前字符是否正处于字符串之中。
	lastIndex := len(sqlTextBytes) - 1 // sql 语句 bytes 的最后一个索引位置。

	for i, b := range sqlTextBytes {
		switch {
		// 遇到字符串开始结束符号'需要设置当前字符串状态。
		case b == '\'':
			inString = !inString
			fixedSqlBytes = append(fixedSqlBytes, b)

		// 如果当前状态正处于字符串中，字符无须特殊处理。
		case inString:
			fixedSqlBytes = append(fixedSqlBytes, b)

		// @ 符号标示参数名称部分开始。
		case b == '@':
			// 连续2个@可以用来转义，需要跳出作用域。
			if inName && i > 0 && sqlTextBytes[i-1] == '@' {
				fixedSqlBytes = append(fixedSqlBytes, b)
				inName = false
				continue
			}
			inName = true
			name = make([]byte, 0, 10)

		// 当前状态正处于参数名中，且当前字符合法，应作为参数名的一部分。
		case inName && unicode.IsOneOf(mysqlParamNameAllowRunes, rune(b)):
			name = append(name, b)
			if lastIndex == i {
				fixedSqlBytes = append(fixedSqlBytes, '?') // 确保最后一个字符的占位也能被加上。
				names = append(names, string(name))
			}

		// 当前状态正处于参数名中，且当前字符不合法，标示着参数名范围结束。
		case inName && !unicode.IsOneOf(mysqlParamNameAllowRunes, rune(b)):
			inName = false
			fixedSqlBytes = append(fixedSqlBytes, '?', b)
			names = append(names, string(name))

		// 非上述情况即为普通 sql 字符部分，无须特殊处理。
		default:
			fixedSqlBytes = append(fixedSqlBytes, b)
		}
	}

	parsedResult := &mysqlNamedParsedResult{string(fixedSqlBytes), names}
	mysqlNamedSqlParsedResult.Store(sqlText, parsedResult) // 缓存结果。
	return parsedResult
}
