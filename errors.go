package sqlmer

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrTran 是数据库执行事务操作遇到的错误。
var ErrTran = errors.New("dbTrans")

// ErrConnect 是 DbClient 获取数据库连接时遇到的错误。
var ErrConnect = errors.New("dbConnect")

// ErrGetEffectedRows 当数据库不支持获取影响行数时候，会返回改类型的错误。
var ErrGetEffectedRows = errors.New("dbClient: the db driver do not support getting effected rows")

// ErrSqlParamParse 解析 SQL 语句中的参数遇到错误时候，会返回该类型错误。
var ErrParseParamFailed = errors.New("dbClient: failed to parse named params")

// ErrExpectedSizeWrong 当执行语句时候，没有影响到预期行数，返回该类型错误。
var ErrExpectedSizeWrong = errors.New("dbClient: effected rows was wrong")

// ErrExecutingSql 当执行 SQL 语句执行时遇到错误，返回该类型错误。
var ErrExecutingSql = errors.New("dbClient: failed to execute sql")

// getExecutingSqlError 用于生成一个带着完整执行参数、语句的 ErrExecutingSql。
func getExecutingSqlError(err error, rawSql string, fixedSql string, params []interface{}) error {
	var sb strings.Builder
	for i, param := range params {
		if i == 0 {
			sb.WriteString("params:\n")
		} else {
			sb.WriteString("\n")
		}

		switch v := param.(type) {
		case sql.NamedArg: // 如果是命名参数，打印出 name/value 对。
			sb.WriteString(fmt.Sprintf("@%s=%s", v.Name, v.Value))
		default:
			sb.WriteString(fmt.Sprintf("@p%d=%v", i+1, v))
		}
	}
	return fmt.Errorf("%w:\ninput sql = %s\nexecuting sql = %s%s", err, rawSql, fixedSql, sb.String())
}
