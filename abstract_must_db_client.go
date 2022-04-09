package sqlmer

import (
	"context"
	"database/sql"

	"github.com/bunnier/sqlmer/sqlen"
)

var _ MustDbClient = (*AbstractDbClient)(nil)

// MustCreateTransaction 用于开始一个事务。
// 	returns:
// 		@tran 返回一个实现了 TransactionKeeper（内嵌 DbClient 接口） 接口的对象，在上面执行的语句会在同一个事务中执行。
func (client *AbstractDbClient) MustCreateTransaction() TransactionKeeper {
	if trans, err := client.CreateTransaction(); err != nil {
		panic(err)
	} else {
		return trans
	}
}

// MustExecute 用于执行非查询 sql 语句，并返回所影响的行数。
// 	params:
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
// 		@effectedRows 语句影响的行数。
//	可能 panic 出的 error （可以通过 errors.Is 判断）：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
// 		@sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
func (client *AbstractDbClient) MustExecute(sqlText string, args ...interface{}) int64 {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustExecuteContext(ctx, sqlText, args...)
}

// MustExecuteContext 用于执行非查询SQL语句，并返回所影响的行数。
// 	params:
// 		@ctx context。
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
// 		@effectedRows 语句影响的行数。
//	可能 panic 出的 error （可以通过 errors.Is 判断）：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
// 		@sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
func (client *AbstractDbClient) MustExecuteContext(ctx context.Context, sqlText string, args ...interface{}) int64 {
	if res, err := client.ExecuteContext(ctx, sqlText, args...); err != nil {
		panic(err)
	} else {
		return res
	}
}

// MustSizedExecute 用于执行非查询 sql 语句，并断言所影响的行数。
// 	params:
// 		@expectedSize 预期的影响行数，当
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
//	可能 panic 出的 error （可以通过 errors.Is 判断）：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
// 		@sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
// 		@sqlmer.ErrExpectedSizeWrong: 当没有影响到预期行数时返回该类错误。
func (client *AbstractDbClient) MustSizedExecute(expectedSize int64, sqlText string, args ...interface{}) {
	ctx, _ := client.getExecTimeoutContext()
	client.MustSizedExecuteContext(ctx, expectedSize, sqlText, args...)
}

// MustSizedExecuteContext 用于执行非查询 sql 语句，并断言所影响的行数。
// 	params:
// 		@ctx context。
// 		@expectedSize 预期的影响行数，当
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
//	可能 panic 出的 error （可以通过 errors.Is 判断）：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
// 		@sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
// 		@sqlmer.ErrExpectedSizeWrong: 当没有影响到预期行数时返回该类错误。
func (client *AbstractDbClient) MustSizedExecuteContext(ctx context.Context, expectedSize int64, sqlText string, args ...interface{}) {
	if err := client.SizedExecuteContext(ctx, expectedSize, sqlText, args...); err != nil {
		panic(err)
	}
}

// MustExists 用于判断给定的查询的结果是否至少包含 1 行。
// 	params:
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
//		@ok 结果至少包含一行。
//	可能 panic 出的 error （可以通过 errors.Is 判断）：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
func (client *AbstractDbClient) MustExists(sqlText string, args ...interface{}) bool {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustExistsContext(ctx, sqlText, args...)
}

// MustExistsContext 用于判断给定的查询的结果是否至少包含 1 行。
// 	params:
// 		@ctx context。
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
//		@ok 结果至少包含一行。
//	可能 panic 出的 error （可以通过 errors.Is 判断）：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
func (client *AbstractDbClient) MustExistsContext(ctx context.Context, sqlText string, args ...interface{}) bool {
	if res, err := client.ExistsContext(ctx, sqlText, args...); err != nil {
		if err == sql.ErrNoRows { // 找不到行时候，仅返回 nil，不做 panic。
			return false
		} else {
			panic(err)
		}
	} else {
		return res
	}
}

// MustScalar 用于获取查询的第一行第一列的值。
// Scalar 用于获取查询的第一行第一列的值。
// 	params:
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
//		@cell 目标查询第一行第一列的值。
//		@hit true 表明有命中数据，false 则为没有命中数据，可通过该值区分是否为数据库空值。
//	可能 panic 出的 error（可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，可以通过第二个返回值区分：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
func (client *AbstractDbClient) MustScalar(sqlText string, args ...interface{}) (interface{}, bool) {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustScalarContext(ctx, sqlText, args...)
}

// MustScalarContext 用于获取查询的第一行第一列的值。
// 	params:
// 		@ctx context。
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
//		@cell 目标查询第一行第一列的值。
//		@hit true 表明有命中数据，false 则为没有命中数据，可通过该值区分是否为数据库空值。
//	可能 panic 出的 error（可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，可以通过第二个返回值区分：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
func (client *AbstractDbClient) MustScalarContext(ctx context.Context, sqlText string, args ...interface{}) (interface{}, bool) {
	if res, hit, err := client.ScalarContext(ctx, sqlText, args...); err != nil {
		panic(err)
	} else {
		return res, hit
	}
}

// MustGet 用于获取查询结果的第一行记录。
// 	params:
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
//		@mapRow 目标查询第一行的结果。
//	可能 panic 出的 error （可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，可以判断返回值是否为 nil：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
func (client *AbstractDbClient) MustGet(sqlText string, args ...interface{}) map[string]interface{} {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustGetContext(ctx, sqlText, args...)
}

// MustGetContext 用于获取查询结果的第一行记录。
// 	params:
// 		@ctx context。
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
//		@mapRow 目标查询第一行的结果。
//	可能 panic 出的 error （可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，可以判断返回值是否为 nil：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
func (client *AbstractDbClient) MustGetContext(ctx context.Context, sqlText string, args ...interface{}) map[string]interface{} {
	if res, err := client.GetContext(ctx, sqlText, args...); err != nil {
		if err == sql.ErrNoRows { // 找不到行时候，仅返回 nil，不做 panic。
			return nil
		} else {
			panic(err)
		}
	} else {
		return res
	}
}

// MustSliceGet 用于获取查询结果的所有行。
// 	params:
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
//		@mapRows 目标查询结果的所有行。
//	可能 panic 出的 error （可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，可以判断返回值是否为 nil：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
func (client *AbstractDbClient) MustSliceGet(sqlText string, args ...interface{}) []map[string]interface{} {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustSliceGetContext(ctx, sqlText, args...)
}

// MustSliceGetContext 用于获取查询结果的所有行。
// 	params:
// 		@ctx context。
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
//		@mapRows 目标查询结果的所有行。
//	可能 panic 出的 error （可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，要知道是否有数据可直接判断返回值的 len。
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
func (client *AbstractDbClient) MustSliceGetContext(ctx context.Context, sqlText string, args ...interface{}) []map[string]interface{} {
	if res, err := client.SliceGetContext(ctx, sqlText, args...); err != nil {
		if err == sql.ErrNoRows { // 找不到行时候，仅返回 nil，不做 panic。
			return nil
		} else {
			panic(err)
		}
	} else {
		return res
	}
}

// MustRow 用于获取单个查询结果行。
// 	params:
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
//		@row 返回目标行的 EnhanceRow 对象（是对 sql.Row 的增强包装对象）。
//	可能 panic 出的 error （可以通过 errors.Is 判断），注意，sql.ErrNoRows 不会 panic，要知道是否有数据可直接判断返回值的 len。
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
func (client *AbstractDbClient) MustRow(sqlText string, args ...interface{}) *sqlen.EnhanceRow {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustRowContext(ctx, sqlText, args...)
}

// MustRowContext 用于获取单个查询结果行。
// 	params:
// 		@ctx context。
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
//		@row 返回目标行的 EnhanceRow 对象（是对 sql.Row 的增强包装对象）。
//	可能 panic 出的 error （可以通过 errors.Is 判断）：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
func (client *AbstractDbClient) MustRowContext(ctx context.Context, sqlText string, args ...interface{}) *sqlen.EnhanceRow {
	if res, err := client.RowContext(ctx, sqlText, args...); err != nil {
		panic(err)
	} else {
		return res
	}
}

// MustRows 用于获取查询结果行的游标对象。
// 	params:
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
//		@row 返回目标行的 EnhanceRows 对象（是对 sql.Rows 的增强包装对象）。
//	可能 panic 出的 error （可以通过 errors.Is 判断）：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
func (client *AbstractDbClient) MustRows(sqlText string, args ...interface{}) *sqlen.EnhanceRows {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustRowsContext(ctx, sqlText, args...)
}

// MustRowsContext 用于获取查询结果行的游标对象。
// 	params:
// 		@ctx context。
// 		@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
// 		@args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// 	returns:
//		@row 返回目标行的 EnhanceRows 对象（是对 sql.Rows 的增强包装对象）。
//	可能 panic 出的 error （可以通过 errors.Is 判断）：
// 		@sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
func (client *AbstractDbClient) MustRowsContext(ctx context.Context, sqlText string, args ...interface{}) *sqlen.EnhanceRows {
	if res, err := client.RowsContext(ctx, sqlText, args...); err != nil {
		panic(err)
	} else {
		return res
	}
}
