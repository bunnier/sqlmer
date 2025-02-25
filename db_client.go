package sqlmer

import (
	"context"
	"time"

	"github.com/bunnier/sqlmer/sqlen"
)

type any = interface{} // 空接口的别名，同时兼容 Go 1.18 。

// ErrorDbClient 为 error 版本 API。
type DbClient interface {
	// Dsn 用于获取当前实例所使用的数据库连接字符串。
	Dsn() string

	// GetConnTimeout 用于获取当前 DbClient 实例的获取连接的超时时间。
	GetConnTimeout() time.Duration

	// GetExecTimeout 用于获取当前 DbClient 实例的执行超时时间。
	GetExecTimeout() time.Duration

	// CreateTransaction 用于开始一个事务。
	// returns:
	//  @tran 返回一个实现了 TransactionKeeper（内嵌 DbClient 接口） 接口的对象，在上面执行的语句会在同一个事务中执行。
	//  @err 创建事务时遇到的错误。
	CreateTransaction() (tran TransactionKeeper, err error)

	// Execute 用于执行非查询SQL语句，并返回所影响的行数。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @effectedRows 语句影响的行数。
	//  @err 执行语句时遇到的错误。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	Execute(sqlText string, args ...any) (rowsEffected int64, err error)

	// ExecuteContext 用于执行非查询 sql 语句，并返回所影响的行数。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @effectedRows 语句影响的行数。
	//  @err 执行语句时遇到的错误。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	ExecuteContext(ctx context.Context, sqlText string, args ...any) (rowsEffected int64, err error)

	// SizedExecute 用于执行非查询SQL语句，并断言所影响的行数。
	// params:
	//  @expectedSize 预期的影响行数，当
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @err 执行语句时遇到的错误。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
	//  - sqlmer.ErrExpectedSizeWrong: 当没有影响到预期行数时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	SizedExecute(expectedSize int64, sqlText string, args ...any) error

	// SizedExecuteContext 用于执行非查询SQL语句，并断言所影响的行数。
	// params:
	//  @ctx context。
	//  @expectedSize 预期的影响行数，当
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @err 执行语句时遇到的错误。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
	//  - sqlmer.ErrExpectedSizeWrong: 当没有影响到预期行数时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	SizedExecuteContext(ctx context.Context, expectedSize int64, sqlText string, args ...any) error

	// Exists 用于判断给定的查询的结果是否至少包含 1 行。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @ok 结果至少包含一行。
	//  @err 执行语句时遇到的错误。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	Exists(sqlText string, args ...any) (ok bool, err error)

	// ExistsContext 用于判断给定的查询的结果是否至少包含 1 行。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @ok 结果至少包含一行。
	//  @err 执行语句时遇到的错误。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	ExistsContext(ctx context.Context, sqlText string, args ...any) (ok bool, err error)

	// Scalar 用于获取查询的第一行第一列的值。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @cell 目标查询第一行第一列的值。
	//  @hit true 表明有命中数据，false 则为没有命中数据，可通过该值区分是否为数据库空值。
	//  @err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，而通过第二个返回值区分。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	Scalar(sqlText string, args ...any) (cell any, hit bool, err error)

	// ScalarContext 用于获取查询的第一行第一列的值。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @cell 目标查询第一行第一列的值。
	//  @hit true 表明有命中数据，false 则为没有命中数据，可通过该值区分是否为数据库空值。
	//  @err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，而通过第二个返回值区分。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	ScalarContext(ctx context.Context, sqlText string, args ...any) (cell any, hit bool, err error)

	// Get 用于获取查询结果的第一行记录。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @mapRow 目标查询第一行的结果。
	//  @err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，要知道是否有数据可直接判断第一个返回值是否为 nil。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	Get(sqlText string, args ...any) (mapRow map[string]any, err error)

	// GetContext 用于获取查询结果的第一行记录。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @mapRow 目标查询第一行的结果。
	//  @err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，要知道是否有数据可直接判断第一个返回值是否为 nil。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	GetContext(ctx context.Context, sqlText string, args ...any) (mapRow map[string]any, err error)

	// SliceGet 用于获取查询结果的所有行。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @mapRows 目标查询结果的所有行。
	//  @err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，要知道是否有数据可直接判断第一个返回值的 len。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	SliceGet(sqlText string, args ...any) (mapRows []map[string]any, err error)

	// SliceGetContext 用于获取查询结果得行序列。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @mapRows 目标查询结果的所有行。
	//  @err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，要知道是否有数据可直接判断第一个返回值的 len。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	SliceGetContext(ctx context.Context, sqlText string, args ...any) (mapRows []map[string]any, err error)

	// Row 用于获取单个查询结果行。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @row 返回目标行的 EnhanceRow 对象（是对 sql.Row 的增强包装对象）。
	//  @err 执行语句时遇到的错误。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	Row(sqlText string, args ...any) (row *sqlen.EnhanceRow, err error)

	// RowContext 用于获取单个查询结果行。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @row 返回目标行的 EnhanceRow 对象（是对 sql.Row 的增强包装对象）。
	//  @err 执行语句时遇到的错误。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	RowContext(context context.Context, sqlText string, args ...any) (row *sqlen.EnhanceRow, err error)

	// Rows 用于获取查询结果行的游标对象。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @row 返回目标行的 EnhanceRows 对象（是对 sql.Rows 的增强包装对象）。
	//  @err 执行语句时遇到的错误。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	Rows(sqlText string, args ...any) (rows *sqlen.EnhanceRows, err error)

	// RowsContext 用于获取查询结果行的游标对象。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @row 返回目标行的 EnhanceRows 对象（是对 sql.Rows 的增强包装对象）。
	//  @err 执行语句时遇到的错误。
	// 可以通过 errors.Is 判断的特殊 err：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	RowsContext(context context.Context, sqlText string, args ...any) (rows *sqlen.EnhanceRows, err error)
}

// TransactionKeeper 是一个定义数据库事务容器。
type TransactionKeeper interface {
	DbClient // DbClient 实现了数据库访问客户端的功能。
	// Commit 用于提交事务。
	Commit() error

	// Rollback 用于回滚事务。
	Rollback() error

	// Close 用于优雅关闭事务，创建事务后可 defer 执行本方法。
	Close() error
}
