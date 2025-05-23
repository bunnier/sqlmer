package sqlmer

import (
	"context"

	"github.com/bunnier/sqlmer/sqlen"
)

// MustDbClient 提供了 DBClient 的 panic 版本 API
type MustDbClient interface {

	// MustCreateTransactionEx（和 MustCreateTransaction 一致） 用于开始一个事务。
	// returns:
	//	@tran 返回一个TransactionKeeperEx 实例（实现了 TransactionKeeper、DbClient 接口） 接口的对象，在上面执行的语句会在同一个事务中执行。
	MustCreateTransaction() (tran *TransactionKeeperEx)

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
