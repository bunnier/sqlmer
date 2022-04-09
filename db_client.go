package sqlmer

import (
	"context"

	"github.com/bunnier/sqlmer/sqlen"
)

// DbClient 定义了数据库访问客户端。
type DbClient interface {
	Dsn() string  // Dsn 用于获取当前实例所使用的数据库连接字符串。
	ErrorDbClient // error 版本 API。
	MustDbClient  // panic 版本 API。
}

// ErrorDbClient 为 error 版本 API。
type ErrorDbClient interface {

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

// MustDbClient 为 panic 版本 API
type MustDbClient interface {

	// MustCreateTransaction 用于开始一个事务。
	// returns:
	//  @tran 返回一个实现了 TransactionKeeper（内嵌 DbClient 接口） 接口的对象，在上面执行的语句会在同一个事务中执行。
	MustCreateTransaction() (tran TransactionKeeper)

	// MustExecute 用于执行非查询 sql 语句，并返回所影响的行数。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @effectedRows 语句影响的行数。
	// 可能 panic 出的 error （可以通过 errors.Is 判断）：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustExecute(sqlText string, args ...any) (rowsEffected int64)

	// MustExecuteContext 用于执行非查询SQL语句，并返回所影响的行数。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @effectedRows 语句影响的行数。
	// 可能 panic 出的 error （可以通过 errors.Is 判断）：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustExecuteContext(context context.Context, sqlText string, args ...any) (rowsEffected int64)

	// MustSizedExecute 用于执行非查询 sql 语句，并断言所影响的行数。
	// params:
	//  @expectedSize 预期的影响行数，当
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// 可能 panic 出的 error （可以通过 errors.Is 判断）：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
	//  - sqlmer.ErrExpectedSizeWrong: 当没有影响到预期行数时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustSizedExecute(expectedSize int64, sqlText string, args ...any)

	// MustSizedExecuteContext 用于执行非查询 sql 语句，并断言所影响的行数。
	// params:
	//  @ctx context。
	//  @expectedSize 预期的影响行数，当
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// 可能 panic 出的 error （可以通过 errors.Is 判断）：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
	//  - sqlmer.ErrExpectedSizeWrong: 当没有影响到预期行数时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustSizedExecuteContext(context context.Context, expectedSize int64, sqlText string, args ...any)

	// MustExists 用于判断给定的查询的结果是否至少包含 1 行。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @ok 结果至少包含一行。
	// 可能 panic 出的 error （可以通过 errors.Is 判断）：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustExists(sqlText string, args ...any) (ok bool)

	// MustExistsContext 用于判断给定的查询的结果是否至少包含 1 行。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @ok 结果至少包含一行。
	// 可能 panic 出的 error （可以通过 errors.Is 判断）：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustExistsContext(context context.Context, sqlText string, args ...any) (ok bool)

	// MustScalar 用于获取查询的第一行第一列的值。
	// Scalar 用于获取查询的第一行第一列的值。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @cell 目标查询第一行第一列的值。
	//  @hit true 表明有命中数据，false 则为没有命中数据，可通过该值区分是否为数据库空值。
	// 可能 panic 出的 error（可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，可以通过第二个返回值区分：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustScalar(sqlText string, args ...any) (cell any, hit bool)

	// MustScalarContext 用于获取查询的第一行第一列的值。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @cell 目标查询第一行第一列的值。
	//  @hit true 表明有命中数据，false 则为没有命中数据，可通过该值区分是否为数据库空值。
	// 可能 panic 出的 error（可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，可以通过第二个返回值区分：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustScalarContext(context context.Context, sqlText string, args ...any) (cell any, hit bool)

	// MustGet 用于获取查询结果的第一行记录。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @mapRow 目标查询第一行的结果。
	// 可能 panic 出的 error （可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，可以判断返回值是否为 nil：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustGet(sqlText string, args ...any) (mapRow map[string]any)

	// MustGetContext 用于获取查询结果的第一行记录。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @mapRow 目标查询第一行的结果。
	// 可能 panic 出的 error （可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，可以判断返回值是否为 nil：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustGetContext(context context.Context, sqlText string, args ...any) (mapRow map[string]any)

	// MustSliceGet 用于获取查询结果的所有行。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @mapRows 目标查询结果的所有行。
	// 可能 panic 出的 error （可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，可以判断返回值是否为 nil：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustSliceGet(sqlText string, args ...any) (mapRows []map[string]any)

	// MustSliceGetContext 用于获取查询结果的所有行。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @mapRows 目标查询结果的所有行。
	// 可能 panic 出的 error （可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，要知道是否有数据可直接判断返回值的 len。
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustSliceGetContext(context context.Context, sqlText string, args ...any) (mapRows []map[string]any)

	// MustRow 用于获取单个查询结果行。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @row 返回目标行的 EnhanceRow 对象（是对 sql.Row 的增强包装对象）。
	// 可能 panic 出的 error （可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，要知道是否有数据可直接判断返回值的 len。
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustRow(sqlText string, args ...any) (row *sqlen.EnhanceRow)

	// MustRowContext 用于获取单个查询结果行。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @row 返回目标行的 EnhanceRow 对象（是对 sql.Row 的增强包装对象）。
	// 可能 panic 出的 error （可以通过 errors.Is 判断）：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustRowContext(context context.Context, sqlText string, args ...any) (row *sqlen.EnhanceRow)

	// MustRows 用于获取查询结果行的游标对象。
	// params:
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @row 返回目标行的 EnhanceRows 对象（是对 sql.Rows 的增强包装对象）。
	// 可能 panic 出的 error （可以通过 errors.Is 判断）：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustRows(sqlText string, args ...any) (rows *sqlen.EnhanceRows)

	// MustRowsContext 用于获取查询结果行的游标对象。
	// params:
	//  @ctx context。
	//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
	//  @args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
	// returns:
	//  @row 返回目标行的 EnhanceRows 对象（是对 sql.Rows 的增强包装对象）。
	// 可能 panic 出的 error （可以通过 errors.Is 判断）：
	//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
	//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
	MustRowsContext(context context.Context, sqlText string, args ...any) (rows *sqlen.EnhanceRows)
}

// TransactionKeeper 是一个定义数据库事务容器。
type TransactionKeeper interface {
	DbClient               // DbClient 实现了数据库访问客户端的功能。
	ErrorTransactionKeeper // error 版本 API。
	MustTransactionKeeper  // panic 版本 API。
}

// ErrorTransactionKeeper 为 error 版本 API。
type ErrorTransactionKeeper interface {
	// Commit 用于提交事务。
	Commit() error

	// Rollback 用于回滚事务。
	Rollback() error

	// Close 用于优雅关闭事务，创建事务后可 defer 执行本方法。
	Close() error
}

// MustTransactionKeeper 为 panic 版本 API。
type MustTransactionKeeper interface {
	// MustClose 用于提交事务。
	MustCommit()

	// MustClose 用于回滚事务。
	MustRollback()

	// MustClose 用于优雅关闭事务，创建事务后可 defer 执行本方法。
	MustClose()
}
