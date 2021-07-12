package sqlmer

import (
	"context"

	"github.com/bunnier/sqlmer/internal/sqlen"
)

// DbClient 定义了数据库访问客户端。
type DbClient interface {
	// CreateTransaction 用于开始一个事务。
	CreateTransaction() (TransactionKeeper, error)

	// ConnectionString 用于获取当前实例所使用的数据库连接字符串。
	ConnectionString() string

	// Scalar 用于获取查询的第一行第一列的值。
	Scalar(sqlText string, args ...interface{}) (interface{}, error)

	// Scalar 用于获取查询的第一行第一列的值。
	ScalarContext(context context.Context, sqlText string, args ...interface{}) (interface{}, error)

	// Execute 用于执行非查询SQL语句，并返回所影响的行数。
	Execute(sqlText string, args ...interface{}) (int64, error)

	// Execute 用于执行非查询SQL语句，并返回所影响的行数。
	ExecuteContext(context context.Context, sqlText string, args ...interface{}) (int64, error)

	// SizedExecute 用于执行非查询SQL语句，并断言所影响的行数。若影响的函数不正确，抛出异常。
	SizedExecute(expectedSize int64, sqlText string, args ...interface{}) error

	// SizedExecute 用于执行非查询SQL语句，并断言所影响的行数。若影响的函数不正确，抛出异常。
	SizedExecuteContext(context context.Context, expectedSize int64, sqlText string, args ...interface{}) error

	// Exists 用于判断给定的查询的结果是否至少包含1行。
	Exists(sqlText string, args ...interface{}) (bool, error)

	// Exists 用于判断给定的查询的结果是否至少包含1行。
	ExistsContext(context context.Context, sqlText string, args ...interface{}) (bool, error)

	// Get 用于获取查询结果的第一行记录。
	Get(sqlText string, args ...interface{}) (map[string]interface{}, error)

	// GetContext 用于获取查询结果的第一行记录。
	GetContext(context context.Context, sqlText string, args ...interface{}) (map[string]interface{}, error)

	// SliceGet 用于获取查询结果得行序列。
	SliceGet(sqlText string, args ...interface{}) ([]map[string]interface{}, error)

	// SliceGetContext 用于获取查询结果得行序列。
	SliceGetContext(context context.Context, sqlText string, args ...interface{}) ([]map[string]interface{}, error)

	// SliceGet 用于获取查询结果得行序列。
	Rows(sqlText string, args ...interface{}) (*sqlen.EnhanceRows, error)

	// SliceGetContext 用于获取查询结果得行序列。
	RowsContext(context context.Context, sqlText string, args ...interface{}) (*sqlen.EnhanceRows, error)
}

// ITransactionKeeper 是一个定义数据库事务容器。
type TransactionKeeper interface {
	// 实现了数据库访问客户端的功能。
	DbClient

	// Commit 用于提交事务。
	Commit() error

	// Rollback 用于回滚事务。
	Rollback() error

	// Close 用于优雅关闭事务，创建事务后应defer执行本方法。
	Close() error
}
