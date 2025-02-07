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
			sb.WriteString(fmt.Sprintf("@%s=%v", v.Name, cutLongStringParams(v.Value)))

		// 非命名参数，索引按顺序打印。
		default:
			sb.WriteString(fmt.Sprintf("@p%d=%v", i+1, cutLongStringParams(v)))
		}
	}
	return sb.String()
}

// MaxLengthErrorValue 用于限制错误输出中参数的长度，超过该大小将会进行截断。
// NOTE: 可自行调整该值。
var MaxLengthErrorValue = 64 * 1024

// 本方法用于对参数进行处理，以避免在 error 中输出过大的字符串。
func cutLongStringParams(paramVal any) any {
	var paramValString string
	switch v := paramVal.(type) {
	case string:
		paramValString = v
	case fmt.Stringer:
		paramValString = v.String()
	default:
		return paramVal
	}

	if len(paramValString) <= MaxLengthErrorValue {
		return paramValString
	}

	// 超过长度的字符串，截取前 MaxLengthErrorValue 个字符，后面用 ... 填充，并注明参数大小。
	var paramStringBuilder strings.Builder
	paramStringBuilder.Grow(MaxLengthErrorValue + 24)
	paramStringBuilder.WriteString(paramValString[:MaxLengthErrorValue])
	paramStringBuilder.WriteString("...")
	paramStringBuilder.WriteString(fmt.Sprintf("(length=%d)", len(paramValString)))

	return paramStringBuilder.String()
}
