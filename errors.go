package sqlmer

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrTran 是数据库执行事务操作遇到的错误。
	ErrTran = errors.New("dbTrans")

	// ErrConnect 是 DbClient 获取数据库连接时遇到的错误。
	ErrConnect = errors.New("dbConnect")

	// ErrGetEffectedRows 当数据库不支持获取影响行数时候，会返回改类型的错误。
	ErrGetEffectedRows = errors.New("dbClient: the db driver do not support getting effected rows")

	// ErrSqlParamParse 解析 SQL 语句中的参数遇到错误时候，会返回该类型错误。
	ErrParseParamFailed = errors.New("dbClient: failed to parse named params")

	// ErrExpectedSizeWrong 当执行语句时候，没有影响到预期行数，返回该类型错误。
	ErrExpectedSizeWrong = errors.New("dbClient: effected rows was wrong")

	// ErrExecutingSql 当执行 SQL 语句执行时遇到错误，返回该类型错误。
	ErrExecutingSql = errors.New("dbClient: failed to execute sql")
)

// MaxLengthErrorValue 用于限制错误中 Value 值的长度，超过该大小将会进行截断。
var MaxLengthErrorValue = 64 * 1024

// getExecutingSqlError 用于生成一个带着 SQL 和参数列表的 ErrExecutingSql。
// 错误内容中包含了：原始传入的 SQL，解析后的 SQL，参数列表。
func getExecutingSqlError(err error, rawSql string, fixedSql string, params []any) error {
	sb := printSqlParams(params)
	return fmt.Errorf("%w\nraw error: %s\nsql:\ninput sql=%s\nexecuting sql=%s\n%s", ErrExecutingSql, err.Error(), rawSql, fixedSql, sb)
}

// getSqlError 用于生成一个带着 SQL 和参数列表的指定错误。
func getSqlError(err error, rawSql string, params []any) error {
	sb := printSqlParams(params)
	return fmt.Errorf("%w\nsql:\ninput sql=%s\n%s", err, rawSql, sb)
}

// printSqlParams 用于打印 SQL 参数列表。
func printSqlParams(params []interface{}) string {
	var sb strings.Builder
	for i, param := range params {
		if i == 0 {
			sb.WriteString("params:\n")
		} else {
			sb.WriteString("\n")
		}

		switch v := param.(type) {
		// 如果是命名参数，打印出 name/value 对。
		case sql.NamedArg:
			logVal := v.Value

			stringVal, ok := v.Value.(fmt.Stringer)
			if ok {
				logStringValue := stringVal.String()
				// string 类型的日志，参考 MaxLengthErrorValue 的值，对输出长度进行截取，以避免 Value 长度过长时候，输出过大的日志。
				if len(logStringValue) > MaxLengthErrorValue {
					logStringValue = logStringValue[:MaxLengthErrorValue]
				}
				logVal = logStringValue
			}

			sb.WriteString(fmt.Sprintf("@%s=%v", v.Name, logVal))

		// 非命名参数，索引按顺序打印。
		default:
			sb.WriteString(fmt.Sprintf("@p%d=%v", i+1, v))
		}
	}
	return sb.String()
}
