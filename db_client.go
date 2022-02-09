package sqlmer

import (
	"context"

	"github.com/bunnier/sqlmer/internal/sqlen"
)

// DbClient 定义了数据库访问客户端。
type DbClient interface {
	ConnectionString() string // ConnectionString 用于获取当前实例所使用的数据库连接字符串。
	errorDbClient             // error 版本 API。
	mustDbClient              // panic 版本 API。
}

// errorDbClient 为 error 版本 API。
type errorDbClient interface {
	// CreateTransaction 用于开始一个事务。
	CreateTransaction() (TransactionKeeper, error)

	// Scalar 用于获取查询的第一行第一列的值。
	Scalar(sqlText string, args ...interface{}) (interface{}, error)

	// ScalarContext 用于获取查询的第一行第一列的值。
	ScalarContext(context context.Context, sqlText string, args ...interface{}) (interface{}, error)

	// Execute 用于执行非查询SQL语句，并返回所影响的行数。
	Execute(sqlText string, args ...interface{}) (int64, error)

	// ExecuteContext 用于执行非查询SQL语句，并返回所影响的行数。
	ExecuteContext(context context.Context, sqlText string, args ...interface{}) (int64, error)

	// SizedExecute 用于执行非查询SQL语句，并断言所影响的行数。若影响的函数不正确，抛出异常。
	SizedExecute(expectedSize int64, sqlText string, args ...interface{}) error

	// SizedExecuteContext 用于执行非查询SQL语句，并断言所影响的行数。若影响的函数不正确，抛出异常。
	SizedExecuteContext(context context.Context, expectedSize int64, sqlText string, args ...interface{}) error

	// Exists 用于判断给定的查询的结果是否至少包含1行。
	Exists(sqlText string, args ...interface{}) (bool, error)

	// ExistsContext 用于判断给定的查询的结果是否至少包含1行。
	ExistsContext(context context.Context, sqlText string, args ...interface{}) (bool, error)

	// Get 用于获取查询结果的第一行记录。
	Get(sqlText string, args ...interface{}) (map[string]interface{}, error)

	// GetContext 用于获取查询结果的第一行记录。
	GetContext(context context.Context, sqlText string, args ...interface{}) (map[string]interface{}, error)

	// SliceGet 用于获取查询结果得行序列。
	SliceGet(sqlText string, args ...interface{}) ([]map[string]interface{}, error)

	// SliceGetContext 用于获取查询结果得行序列。
	SliceGetContext(context context.Context, sqlText string, args ...interface{}) ([]map[string]interface{}, error)

	// Rows 用于获取查询结果得行序列。
	Rows(sqlText string, args ...interface{}) (*sqlen.EnhanceRows, error)

	// RowsContext 用于获取查询结果得行序列。
	RowsContext(context context.Context, sqlText string, args ...interface{}) (*sqlen.EnhanceRows, error)
}

// mustDbClient 为 panic 版本 API
type mustDbClient interface {
	// MustCreateTransaction 用于开始一个事务。
	MustCreateTransaction() TransactionKeeper

	// MustScalar 用于获取查询的第一行第一列的值。
	MustScalar(sqlText string, args ...interface{}) interface{}

	// MustScalarContext 用于获取查询的第一行第一列的值。
	MustScalarContext(context context.Context, sqlText string, args ...interface{}) interface{}

	// MustExecute 用于执行非查询SQL语句，并返回所影响的行数。
	MustExecute(sqlText string, args ...interface{}) int64

	// MustExecuteContext 用于执行非查询SQL语句，并返回所影响的行数。
	MustExecuteContext(context context.Context, sqlText string, args ...interface{}) int64

	// MustSizedExecute 用于执行非查询SQL语句，并断言所影响的行数。若影响的函数不正确，抛出异常。
	MustSizedExecute(expectedSize int64, sqlText string, args ...interface{})

	// MustSizedExecuteContext 用于执行非查询SQL语句，并断言所影响的行数。若影响的函数不正确，抛出异常。
	MustSizedExecuteContext(context context.Context, expectedSize int64, sqlText string, args ...interface{})

	// MustExists 用于判断给定的查询的结果是否至少包含1行。
	MustExists(sqlText string, args ...interface{}) bool

	// MustExistsContext 用于判断给定的查询的结果是否至少包含1行。
	MustExistsContext(context context.Context, sqlText string, args ...interface{}) bool

	// MustGet 用于获取查询结果的第一行记录。
	MustGet(sqlText string, args ...interface{}) map[string]interface{}

	// MustGetContext 用于获取查询结果的第一行记录。
	MustGetContext(context context.Context, sqlText string, args ...interface{}) map[string]interface{}

	// MustSliceGet 用于获取查询结果得行序列。
	MustSliceGet(sqlText string, args ...interface{}) []map[string]interface{}

	// MustSliceGetContext 用于获取查询结果得行序列。
	MustSliceGetContext(context context.Context, sqlText string, args ...interface{}) []map[string]interface{}

	// MustRows 用于获取查询结果得行序列。
	MustRows(sqlText string, args ...interface{}) *sqlen.EnhanceRows

	// MustRowsContext 用于获取查询结果得行序列。
	MustRowsContext(context context.Context, sqlText string, args ...interface{}) *sqlen.EnhanceRows
}

// TransactionKeeper 是一个定义数据库事务容器。
type TransactionKeeper interface {
	DbClient               // DbClient 实现了数据库访问客户端的功能。
	errorTransactionKeeper // error 版本 API。
	mustTransactionKeeper  // panic 版本 API。
}

// errorTransactionKeeper 为 error 版本 API。
type errorTransactionKeeper interface {
	// Commit 用于提交事务。
	Commit() error

	// Rollback 用于回滚事务。
	Rollback() error

	// Close 用于优雅关闭事务，创建事务后应 defer 执行本方法。
	Close() error
}

// mustTransactionKeeper 为 panic 版本 API。
type mustTransactionKeeper interface {
	// MustClose 用于提交事务。
	MustCommit()

	// MustClose 用于回滚事务。
	MustRollback()

	// MustClose 用于优雅关闭事务，创建事务后应 defer 执行本方法。
	MustClose()
}
