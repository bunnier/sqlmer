package sqlmer

import (
	"context"
	"database/sql"

	"github.com/bunnier/sqlmer/sqlen"
)

var _ mustDbClient = (*AbstractDbClient)(nil)

// MustCreateTransaction 用于开始一个事务。
func (client *AbstractDbClient) MustCreateTransaction() TransactionKeeper {
	if trans, err := client.CreateTransaction(); err != nil {
		panic(err)
	} else {
		return trans
	}
}

// MustExecute 用于执行非查询 sql 语句，并返回所影响的行数。
func (client *AbstractDbClient) MustExecute(sqlText string, args ...interface{}) int64 {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustExecuteContext(ctx, sqlText, args...)
}

// MustExecuteContext 用于执行非查询 sql 语句，并返回所影响的行数。
func (client *AbstractDbClient) MustExecuteContext(ctx context.Context, sqlText string, args ...interface{}) int64 {
	if res, err := client.ExecuteContext(ctx, sqlText, args...); err != nil {
		panic(err)
	} else {
		return res
	}
}

// MustSizedExecute 用于执行非查询 sql 语句，并断言所影响的行数。
func (client *AbstractDbClient) MustSizedExecute(expectedSize int64, sqlText string, args ...interface{}) {
	ctx, _ := client.getExecTimeoutContext()
	client.MustSizedExecuteContext(ctx, expectedSize, sqlText, args...)
}

// MustSizedExecuteContext 用于执行非查询 sql 语句，并断言所影响的行数。
func (client *AbstractDbClient) MustSizedExecuteContext(ctx context.Context, expectedSize int64, sqlText string, args ...interface{}) {
	if err := client.SizedExecuteContext(ctx, expectedSize, sqlText, args...); err != nil {
		panic(err)
	}
}

// MustExists 用于判断给定的查询的结果是否至少包含 1 行。
// 注意：当查询不到行时候，将返回 false，而不是 panic。
func (client *AbstractDbClient) MustExists(sqlText string, args ...interface{}) bool {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustExistsContext(ctx, sqlText, args...)
}

// MustExistsContext 用于判断给定的查询的结果是否至少包含 1 行。
// 注意：当查询不到行时候，将返回 false，而不是 panic。
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
// 注意，sql.ErrNoRows 不会引发 panic，而通过第二个返回值区分，当查询不到数据的时候第二个返回值将为 false，否则为 true。
func (client *AbstractDbClient) MustScalar(sqlText string, args ...interface{}) (interface{}, bool) {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustScalarContext(ctx, sqlText, args...)
}

// MustScalarContext 用于获取查询的第一行第一列的值。
// 注意，sql.ErrNoRows 不会引发 panic，而通过第二个返回值区分，当查询不到数据的时候第二个返回值将为 false，否则为 true。
func (client *AbstractDbClient) MustScalarContext(ctx context.Context, sqlText string, args ...interface{}) (interface{}, bool) {
	if res, hit, err := client.ScalarContext(ctx, sqlText, args...); err != nil {
		panic(err)
	} else {
		return res, hit
	}
}

// MustGet 用于获取查询结果的第一行记录。
// 注意：当查询不到行时候，将返回 nil，而不是 panic。
func (client *AbstractDbClient) MustGet(sqlText string, args ...interface{}) map[string]interface{} {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustGetContext(ctx, sqlText, args...)
}

// MustGetContext 用于获取查询结果的第一行记录。
// 注意：当查询不到行时候，将返回 nil，而不是 panic。
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

// MustSliceGet 用于获取查询结果得行序列。
// 注意：当查询不到行时候，将返回 nil，而不是 panic。
func (client *AbstractDbClient) MustSliceGet(sqlText string, args ...interface{}) []map[string]interface{} {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustSliceGetContext(ctx, sqlText, args...)
}

// MustSliceGetContext 用于获取查询结果得行序列。
// 注意：当查询不到行时候，将返回 nil，而不是 panic。
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
func (client *AbstractDbClient) MustRow(sqlText string, args ...interface{}) *sqlen.EnhanceRow {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustRowContext(ctx, sqlText, args...)
}

// RowMustRowContextContext 用于获取单个查询结果行。
func (client *AbstractDbClient) MustRowContext(ctx context.Context, sqlText string, args ...interface{}) *sqlen.EnhanceRow {
	if res, err := client.RowContext(ctx, sqlText, args...); err != nil {
		panic(err)
	} else {
		return res
	}
}

// MustRows 用于获取读取数据的游标 sql.Rows。
func (client *AbstractDbClient) MustRows(sqlText string, args ...interface{}) *sqlen.EnhanceRows {
	ctx, _ := client.getExecTimeoutContext()
	return client.MustRowsContext(ctx, sqlText, args...)
}

// MustRowsContext 用于获取读取数据的游标 sql.Rows。
func (client *AbstractDbClient) MustRowsContext(ctx context.Context, sqlText string, args ...interface{}) *sqlen.EnhanceRows {
	if res, err := client.RowsContext(ctx, sqlText, args...); err != nil {
		panic(err)
	} else {
		return res
	}
}
