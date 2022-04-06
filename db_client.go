package sqlmer

import (
	"context"

	"github.com/bunnier/sqlmer/sqlen"
)

type any = interface{} // 空接口的别名，同时兼容 Go 1.18 。

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

	// Execute 用于执行非查询SQL语句，并返回所影响的行数。
	Execute(sqlText string, args ...interface{}) (int64, error)

	// ExecuteContext 用于执行非查询SQL语句，并返回所影响的行数。
	ExecuteContext(context context.Context, sqlText string, args ...interface{}) (int64, error)

	// SizedExecute 用于执行非查询SQL语句，并断言所影响的行数。若影响的函数不正确，抛出异常。
	SizedExecute(expectedSize int64, sqlText string, args ...interface{}) error

	// SizedExecuteContext 用于执行非查询SQL语句，并断言所影响的行数。若影响的函数不正确，抛出异常。
	SizedExecuteContext(context context.Context, expectedSize int64, sqlText string, args ...interface{}) error

	// Exists 用于判断给定的查询的结果是否至少包含 1 行。
	Exists(sqlText string, args ...interface{}) (bool, error)

	// ExistsContext 用于判断给定的查询的结果是否至少包含 1 行。
	ExistsContext(context context.Context, sqlText string, args ...interface{}) (bool, error)

	// Scalar 用于获取查询的第一行第一列的值。
	// 注意，sql.ErrNoRows 不放 error 中返回，而通过第二个返回值区分，当查询不到数据的时候第二个返回值将为 false，否则为 true。
	Scalar(sqlText string, args ...interface{}) (interface{}, bool, error)

	// ScalarContext 用于获取查询的第一行第一列的值。
	// 注意，sql.ErrNoRows 不放 error 中返回，而通过第二个返回值区分，当查询不到数据的时候第二个返回值将为 false，否则为 true。
	ScalarContext(context context.Context, sqlText string, args ...interface{}) (interface{}, bool, error)

	// Get 用于获取查询结果的第一行记录。
	Get(sqlText string, args ...interface{}) (map[string]interface{}, error)

	// GetContext 用于获取查询结果的第一行记录。
	GetContext(context context.Context, sqlText string, args ...interface{}) (map[string]interface{}, error)

	// SliceGet 用于获取查询结果得行序列。
	SliceGet(sqlText string, args ...interface{}) ([]map[string]interface{}, error)

	// SliceGetContext 用于获取查询结果得行序列。
	SliceGetContext(context context.Context, sqlText string, args ...interface{}) ([]map[string]interface{}, error)

	// Row 用于获取单个查询结果行。
	Row(sqlText string, args ...interface{}) (*sqlen.EnhanceRow, error)

	// RowsContext 用于获取单个查询结果行。
	RowContext(context context.Context, sqlText string, args ...interface{}) (*sqlen.EnhanceRow, error)

	// Rows 用于获取查询结果行序列。
	Rows(sqlText string, args ...interface{}) (*sqlen.EnhanceRows, error)

	// RowsContext 用于获取查询结果行序列。
	RowsContext(context context.Context, sqlText string, args ...interface{}) (*sqlen.EnhanceRows, error)
}

// mustDbClient 为 panic 版本 API
type mustDbClient interface {
	// MustCreateTransaction 用于开始一个事务。
	MustCreateTransaction() TransactionKeeper

	// MustExecute 用于执行非查询 sql 语句，并返回所影响的行数。
	MustExecute(sqlText string, args ...interface{}) int64

	// MustExecuteContext 用于执行非查询SQL语句，并返回所影响的行数。
	MustExecuteContext(context context.Context, sqlText string, args ...interface{}) int64

	// MustSizedExecute 用于执行非查询 sql 语句，并断言所影响的行数。
	MustSizedExecute(expectedSize int64, sqlText string, args ...interface{})

	// MustSizedExecuteContext 用于执行非查询 sql 语句，并断言所影响的行数。
	MustSizedExecuteContext(context context.Context, expectedSize int64, sqlText string, args ...interface{})

	// MustExists 用于判断给定的查询的结果是否至少包含 1 行。
	// 注意：当查询不到行时候，将返回 false，而不是 panic。
	MustExists(sqlText string, args ...interface{}) bool

	// MustExistsContext 用于判断给定的查询的结果是否至少包含 1 行。
	// 注意：当查询不到行时候，将返回 false，而不是 panic。
	MustExistsContext(context context.Context, sqlText string, args ...interface{}) bool

	// MustScalar 用于获取查询的第一行第一列的值。
	// 注意，sql.ErrNoRows 不会引发 panic，而通过第二个返回值区分，当查询不到数据的时候第二个返回值将为 false，否则为 true。
	MustScalar(sqlText string, args ...interface{}) (interface{}, bool)

	// MustScalarContext 用于获取查询的第一行第一列的值。
	// 注意，sql.ErrNoRows 不会引发 panic，而通过第二个返回值区分，当查询不到数据的时候第二个返回值将为 false，否则为 true。
	MustScalarContext(context context.Context, sqlText string, args ...interface{}) (interface{}, bool)

	// MustGet 用于获取查询结果的第一行记录。
	// 注意：当查询不到行时候，将返回 nil，而不是 panic。
	MustGet(sqlText string, args ...interface{}) map[string]interface{}

	// MustGetContext 用于获取查询结果的第一行记录。
	// 注意：当查询不到行时候，将返回 nil，而不是 panic。
	MustGetContext(context context.Context, sqlText string, args ...interface{}) map[string]interface{}

	// MustSliceGet 用于获取查询结果得行序列。
	// 注意：当查询不到行时候，将返回 nil，而不是 panic。
	MustSliceGet(sqlText string, args ...interface{}) []map[string]interface{}

	// MustSliceGetContext 用于获取查询结果得行序列。
	// 注意：当查询不到行时候，将返回 nil，而不是 panic。
	MustSliceGetContext(context context.Context, sqlText string, args ...interface{}) []map[string]interface{}

	// MustRow 用于获取单个查询结果行。
	MustRow(sqlText string, args ...interface{}) *sqlen.EnhanceRow

	// MustRowContext 用于获取单个查询结果行。
	MustRowContext(context context.Context, sqlText string, args ...interface{}) *sqlen.EnhanceRow

	// MustRows 用于获取读取数据的游标 sql.Rows。
	MustRows(sqlText string, args ...interface{}) *sqlen.EnhanceRows

	// MustRowsContext 用于获取读取数据的游标 sql.Rows。
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
