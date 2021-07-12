package sqlmer

import (
	"fmt"
)

var _ error = (*DbSqlError)(nil)

// DbSqlError 标示一个Sql执行错误。
type DbSqlError struct {
	Msg     string
	SqlText string
}

func (err *DbSqlError) Error() string {
	return fmt.Sprintf("sql err: msg=%s; sql=%s", err.Msg, err.SqlText)
}

// NewDbSqlError 用于创建一个Sql执行错误。
func NewDbSqlError(msg string, sqlText string) *DbSqlError {
	return &DbSqlError{msg, sqlText}
}

var _ error = (*DbConnError)(nil)

// DbSqlError 标示一个Sql连接错误。
type DbConnError struct {
	Msg              string
	ConnectionString string
}

func (err *DbConnError) Error() string {
	return fmt.Sprintf("sql err: msg=%s; connection string=%s", err.Msg, err.ConnectionString)
}

// NewDbSqlError 用于创建一个数据库连接错误。
func NewDbConnError(msg string, sqlText string) *DbConnError {
	return &DbConnError{msg, sqlText}
}

var _ error = (*DbTransError)(nil)

// DbTransError 标示一个数据库事务错误。
type DbTransError struct {
	Msg string
}

func (err *DbTransError) Error() string {
	return fmt.Sprintf("trans err: %s", err.Msg)
}

// NewDbSqlError 用于创建一个数据库连接错误。
func NewDbTransError(msg string) *DbTransError {
	return &DbTransError{msg}
}
