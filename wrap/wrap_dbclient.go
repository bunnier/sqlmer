package wrap

import (
	"context"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/sqlen"
)

var _ sqlmer.DbClient = (*WrappedDbClient)(nil)

// WrappedDbClient 将包包裹 DbClient 的所有 SQL 执行方法，以注入慢日志/统计指标等能力。
type WrappedDbClient struct {
	dbClient sqlmer.DbClient // 原始的 DbClient 实例。
	wrapFunc WrapFunc        // 包裹函数。
}

// WrapFunc 用于包裹 SQL 执行方法。
type WrapFunc func(sql string, args []any) func(error)

// Extend 加强 DbClient，在 DbClient 的数据库访问上，提供一层装饰器包裹，以注入慢日志/统计指标等能力。
// 注意：
//   - 事务是内部语句独立包裹，而不是一整个事务包裹；
//   - Rows 方法，在返回游标对象时包裹上下文即结束；
func Extend(raw sqlmer.DbClient, execWrapFunc WrapFunc) *WrappedDbClient {
	return &WrappedDbClient{raw, execWrapFunc}
}

// Dsn 用于获取当
// Dsn 用于获取当前实例所使用的数据库连接字符串。
func (c *WrappedDbClient) Dsn() string {
	return c.dbClient.Dsn()
}

// GetConnTimeout 用于获取当前 DbClient 实例的获取连接的超时时间。
func (c *WrappedDbClient) GetConnTimeout() time.Duration {
	return c.dbClient.GetConnTimeout()
}

// GetExecTimeout 用于获取当前 DbClient 实例的执行超时时间。
func (c *WrappedDbClient) GetExecTimeout() time.Duration {
	return c.dbClient.GetExecTimeout()
}

// CreateTransaction 用于开始一个事务。
// returns:
//
//	@tran 返回一个实现了 TransactionKeeper（内嵌 DbClient 接口） 接口的对象，在上面执行的语句会在同一个事务中执行。
//	@err 创建事务时遇到的错误。
func (c *WrappedDbClient) CreateTransaction() (tran sqlmer.TransactionKeeper, err error) {
	tran, err = c.dbClient.CreateTransaction()
	if err != nil {
		return
	}

	// 将事务上的方法也包裹上包裹函数。。
	tran = extendExecWrapTx(tran, c.wrapFunc)
	return
}

// Execute 用于执行非查询SQL语句，并返回所影响的行数。
// params:
//
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@effectedRows 语句影响的行数。
//	@err 执行语句时遇到的错误。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) Execute(sqlText string, args ...any) (rowsEffected int64, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	rowsEffected, err = c.dbClient.Execute(sqlText, args...)
	return
}

// ExecuteContext 用于执行非查询 sql 语句，并返回所影响的行数。
// params:
//
//	@ctx context。
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@effectedRows 语句影响的行数。
//	@err 执行语句时遇到的错误。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) ExecuteContext(ctx context.Context, sqlText string, args ...any) (rowsEffected int64, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	rowsEffected, err = c.dbClient.ExecuteContext(ctx, sqlText, args...)
	return
}

// SizedExecute 用于执行非查询SQL语句，并断言所影响的行数。
// params:
//
//	@expectedSize 预期的影响行数，当
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@err 执行语句时遇到的错误。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
//   - sqlmer.ErrExpectedSizeWrong: 当没有影响到预期行数时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) SizedExecute(expectedSize int64, sqlText string, args ...any) (err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	err = c.dbClient.SizedExecute(expectedSize, sqlText, args...)
	return
}

// SizedExecuteContext 用于执行非查询SQL语句，并断言所影响的行数。
// params:
//
//	@ctx context。
//	@expectedSize 预期的影响行数，当
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@err 执行语句时遇到的错误。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
//   - sqlmer.ErrExpectedSizeWrong: 当没有影响到预期行数时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) SizedExecuteContext(ctx context.Context, expectedSize int64, sqlText string, args ...any) (err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	err = c.dbClient.SizedExecuteContext(ctx, expectedSize, sqlText, args...)
	return
}

// Exists 用于判断给定的查询的结果是否至少包含 1 行。
// params:
//
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@ok 结果至少包含一行。
//	@err 执行语句时遇到的错误。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) Exists(sqlText string, args ...any) (ok bool, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	ok, err = c.dbClient.Exists(sqlText, args...)
	return
}

// ExistsContext 用于判断给定的查询的结果是否至少包含 1 行。
// params:
//
//	@ctx context。
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@ok 结果至少包含一行。
//	@err 执行语句时遇到的错误。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) ExistsContext(ctx context.Context, sqlText string, args ...any) (ok bool, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	ok, err = c.dbClient.ExistsContext(ctx, sqlText, args...)
	return
}

// Scalar 用于获取查询的第一行第一列的值。
// params:
//
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@cell 目标查询第一行第一列的值。
//	@hit true 表明有命中数据，false 则为没有命中数据，可通过该值区分是否为数据库空值。
//	@err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，而通过第二个返回值区分。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) Scalar(sqlText string, args ...any) (cell any, hit bool, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	cell, hit, err = c.dbClient.Scalar(sqlText, args...)
	return
}

// ScalarContext 用于获取查询的第一行第一列的值。
// params:
//
//	@ctx context。
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@cell 目标查询第一行第一列的值。
//	@hit true 表明有命中数据，false 则为没有命中数据，可通过该值区分是否为数据库空值。
//	@err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，而通过第二个返回值区分。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) ScalarContext(ctx context.Context, sqlText string, args ...any) (cell any, hit bool, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	cell, hit, err = c.dbClient.ScalarContext(ctx, sqlText, args...)
	return
}

// Get 用于获取查询结果的第一行记录。
// params:
//
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@mapRow 目标查询第一行的结果。
//	@err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，可知是否有数据可直接判断第一个返回值是否为 nil。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) Get(sqlText string, args ...any) (mapRow map[string]any, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	mapRow, err = c.dbClient.Get(sqlText, args...)
	return
}

// GetContext 用于获取查询结果的第一行记录。
// params:
//
//	@ctx context。
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@mapRow 目标查询第一行的结果。
//	@err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，可知是否有数据可直接判断第一个返回值是否为 nil。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) GetContext(ctx context.Context, sqlText string, args ...any) (mapRow map[string]any, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	mapRow, err = c.dbClient.GetContext(ctx, sqlText, args...)
	return
}

// SliceGet 用于获取查询结果的所有行。
// params:
//
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@mapRows 目标查询结果的所有行。
//	@err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，可知是否有数据可直接判断第一个返回值的 len。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) SliceGet(sqlText string, args ...any) (mapRows []map[string]any, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	mapRows, err = c.dbClient.SliceGet(sqlText, args...)
	return
}

// SliceGetContext 用于获取查询结果得行序列。
// params:
//
//	@ctx context。
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@mapRows 目标查询结果的所有行。
//	@err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，可知是否有数据可直接判断第一个返回值的 len。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) SliceGetContext(ctx context.Context, sqlText string, args ...any) (mapRows []map[string]any, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	mapRows, err = c.dbClient.SliceGetContext(ctx, sqlText, args...)
	return
}

// Row 用于获取单个查询结果行。
// params:
//
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@row 返回目标行的 EnhanceRow 对象（是对 sql.Row 的增强包装对象）。
//	@err 执行语句时遇到的错误。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) Row(sqlText string, args ...any) (row *sqlen.EnhanceRow, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	row, err = c.dbClient.Row(sqlText, args...)
	return
}

// RowContext 用于获取单个查询结果行。
// params:
//
//	@ctx context。
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@row 返回目标行的 EnhanceRow 对象（是对 sql.Row 的增强包装对象）。
//	@err 执行语句时遇到的错误。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) RowContext(context context.Context, sqlText string, args ...any) (row *sqlen.EnhanceRow, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	row, err = c.dbClient.RowContext(context, sqlText, args...)
	return
}

// Rows 用于获取查询结果行的游标对象。
// params:
//
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@row 返回目标行的 EnhanceRows 对象（是对 sql.Rows 的增强包装对象）。
//	@err 执行语句时遇到的错误。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) Rows(sqlText string, args ...any) (rows *sqlen.EnhanceRows, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	rows, err = c.dbClient.Rows(sqlText, args...)
	return
}

// RowsContext 用于获取查询结果行的游标对象。
// params:
//
//	@ctx context。
//	@sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//	@args SQL 语句的参数，支持通过 map[string]any 提供命名参数值 或 通过变长参数提供索引参数值。
//
// returns:
//
//	@row 返回目标行的 EnhanceRows 对象（是对 sql.Rows 的增强包装对象）。
//	@err 执行语句时遇到的错误。
//
// 可以通过 errors.Is 判断的特殊 err：
//   - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//   - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (c *WrappedDbClient) RowsContext(context context.Context, sqlText string, args ...any) (rows *sqlen.EnhanceRows, err error) {
	deferFunc := c.wrapFunc(sqlText, args)
	defer func() { // 这个 defer 是为了获取到执行结果的 err。
		deferFunc(err)
	}()
	rows, err = c.dbClient.RowsContext(context, sqlText, args...)
	return
}
