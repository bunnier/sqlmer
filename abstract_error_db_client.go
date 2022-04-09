package sqlmer

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bunnier/sqlmer/sqlen"
)

var _ ErrorDbClient = (*AbstractDbClient)(nil)

// CreateTransaction 用于开始一个事务。
// returns:
//  @tran 返回一个实现了 TransactionKeeper（内嵌 DbClient 接口） 接口的对象，在上面执行的语句会在同一个事务中执行。
//  @err 创建事务时遇到的错误。
func (client *AbstractDbClient) CreateTransaction() (TransactionKeeper, error) {
	tx, err := client.Db.Begin()
	if err != nil {
		return nil, err
	}

	txDbClient := &AbstractDbClient{
		config: client.config,
		Db:     client.Db,                         // Db 对象。
		Exer:   sqlen.NewTxEnhance(tx, client.Db), // 新的client中的实际执行对象使用开启的事务。
	}

	return &abstractTransactionKeeper{
		txDbClient, tx, false, 0,
	}, nil
}

// Execute 用于执行非查询SQL语句，并返回所影响的行数。
// params:
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @effectedRows 语句影响的行数。
//  @err 执行语句时遇到的错误。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
func (client *AbstractDbClient) Execute(sqlText string, args ...interface{}) (int64, error) {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.ExecuteContext(ctx, sqlText, args...)
}

// ExecuteContext 用于执行非查询 sql 语句，并返回所影响的行数。
// params:
//  @ctx context。
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @effectedRows 语句影响的行数。
//  @err 执行语句时遇到的错误。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) ExecuteContext(ctx context.Context, sqlText string, args ...interface{}) (int64, error) {
	fixedSqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return 0, err
	}
	sqlResult, err := client.Exer.ExecContext(ctx, fixedSqlText, args...)
	if err != nil {
		return 0, getExecutingSqlError(err, sqlText, fixedSqlText, args)
	}

	if effectRows, err := sqlResult.RowsAffected(); err != nil {
		return 0, fmt.Errorf("%w: %s", ErrGetEffectedRows, err.Error())
	} else {
		return effectRows, nil
	}
}

// SizedExecute 用于执行非查询SQL语句，并断言所影响的行数。
// params:
//  @expectedSize 预期的影响行数，当
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @err 执行语句时遇到的错误。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
//  - sqlmer.ErrExpectedSizeWrong: 当没有影响到预期行数时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) SizedExecute(expectedSize int64, sqlText string, args ...interface{}) error {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.SizedExecuteContext(ctx, expectedSize, sqlText, args...)
}

// SizedExecuteContext 用于执行非查询SQL语句，并断言所影响的行数。
// params:
//  @ctx context。
//  @expectedSize 预期的影响行数，当
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @err 执行语句时遇到的错误。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrGetEffectedRows: 当执行成功，但驱动不支持获取影响行数时候，返回该类型错误。
//  - sqlmer.ErrExpectedSizeWrong: 当没有影响到预期行数时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) SizedExecuteContext(ctx context.Context, expectedSize int64, sqlText string, args ...interface{}) error {
	effectedRow, err := client.ExecuteContext(ctx, sqlText, args...)
	if err != nil {
		return err
	}
	if effectedRow != expectedSize {
		return fmt.Errorf("%w: expected: %d, actually: %d\nsql = %s", ErrExpectedSizeWrong, expectedSize, effectedRow, sqlText)
	}
	return nil
}

// Exists 用于判断给定的查询的结果是否至少包含 1 行。
// params:
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @ok 结果至少包含一行。
//  @err 执行语句时遇到的错误。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) Exists(sqlText string, args ...interface{}) (bool, error) {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.ExistsContext(ctx, sqlText, args...)
}

// ExistsContext 用于判断给定的查询的结果是否至少包含 1 行。
// params:
//  @ctx context。
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @ok 结果至少包含一行。
//  @err 执行语句时遇到的错误。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) ExistsContext(ctx context.Context, sqlText string, args ...interface{}) (bool, error) {
	fixedSqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return false, err
	}

	rows, err := client.Exer.EnhancedQueryContext(ctx, fixedSqlText, args...)
	if err != nil {
		return false, getExecutingSqlError(err, sqlText, fixedSqlText, args)
	}
	defer rows.Close()

	hasData := rows.Next()
	if err = rows.Err(); err != nil {
		return false, err
	}
	return hasData, nil
}

// Scalar 用于获取查询的第一行第一列的值。
// params:
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @cell 目标查询第一行第一列的值。
//  @hit true 表明有命中数据，false 则为没有命中数据，可通过该值区分是否为数据库空值。
//  @err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，而通过第二个返回值区分。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) Scalar(sqlText string, args ...interface{}) (interface{}, bool, error) {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.ScalarContext(ctx, sqlText, args...)
}

// ScalarContext 用于获取查询的第一行第一列的值。
// params:
//  @ctx context。
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @cell 目标查询第一行第一列的值。
//  @hit true 表明有命中数据，false 则为没有命中数据，可通过该值区分是否为数据库空值。
//  @err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，而通过第二个返回值区分。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) ScalarContext(ctx context.Context, sqlText string, args ...interface{}) (interface{}, bool, error) {
	fixedSqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return nil, false, err
	}

	if result, err := client.Exer.EnhancedQueryRowContext(ctx, fixedSqlText, args...).SliceScan(); err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil // 没有命中行时候，不用 error 返回，而是通过第二个参数标识。
		}
		return nil, false, getExecutingSqlError(err, sqlText, fixedSqlText, args)
	} else {
		return result[0], true, nil // 只要没有 error，至少有 1 列的。
	}
}

// Get 用于获取查询结果的第一行记录。
// params:
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @mapRow 目标查询第一行的结果。
//  @err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，要知道是否有数据可直接判断第一个返回值是否为 nil。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) Get(sqlText string, args ...interface{}) (map[string]interface{}, error) {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.GetContext(ctx, sqlText, args...)
}

// GetContext 用于获取查询结果的第一行记录。
// params:
//  @ctx context。
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @mapRow 目标查询第一行的结果。
//  @err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，要知道是否有数据可直接判断第一个返回值是否为 nil。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) GetContext(ctx context.Context, sqlText string, args ...interface{}) (map[string]interface{}, error) {
	fixedSqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return nil, err
	}
	rows, err := client.Exer.EnhancedQueryContext(ctx, fixedSqlText, args...)
	if err != nil {
		return nil, getExecutingSqlError(err, sqlText, fixedSqlText, args)
	}
	defer rows.Close()

	if rows.Next() { // 这个地方不直接 EnhancedQueryRowContext.MapScan 主要是可以在没有行时候省略一次 make map。
		result := make(map[string]interface{})
		err = rows.MapScan(result)
		return result, err
	}
	return nil, rows.Err()
}

// SliceGet 用于获取查询结果的所有行。
// params:
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @mapRows 目标查询结果的所有行。
//  @err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，要知道是否有数据可直接判断第一个返回值的 len。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) SliceGet(sqlText string, args ...interface{}) ([]map[string]interface{}, error) {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.SliceGetContext(ctx, sqlText, args...)
}

// SliceGetContext 用于获取查询结果得行序列。
// params:
//  @ctx context。
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @mapRows 目标查询结果的所有行。
//  @err 执行语句时遇到的错误。注意，sql.ErrNoRows 不放 error 中返回，要知道是否有数据可直接判断第一个返回值的 len。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) SliceGetContext(ctx context.Context, sqlText string, args ...interface{}) ([]map[string]interface{}, error) {
	fixedSqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return nil, err
	}
	rows, err := client.Exer.EnhancedQueryContext(ctx, fixedSqlText, args...)
	if err != nil {
		return nil, getExecutingSqlError(err, sqlText, fixedSqlText, args)
	}
	defer rows.Close()

	results := make([]map[string]interface{}, 0, 5)
	for rows.Next() { // 这个地方不直接 EnhancedQueryRowContext.MapScan 主要是可以在没有行时候省略一次 make map。
		result := make(map[string]interface{})
		err = rows.MapScan(result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

// Row 用于获取单个查询结果行。
// params:
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @row 返回目标行的 EnhanceRow 对象（是对 sql.Row 的增强包装对象）。
//  @err 执行语句时遇到的错误。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) Row(sqlText string, args ...interface{}) (*sqlen.EnhanceRow, error) {
	ctx, _ := client.getExecTimeoutContext()
	return client.RowContext(ctx, sqlText, args...)
}

// RowContext 用于获取单个查询结果行。
// params:
//  @ctx context。
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @row 返回目标行的 EnhanceRow 对象（是对 sql.Row 的增强包装对象）。
//  @err 执行语句时遇到的错误。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) RowContext(ctx context.Context, sqlText string, args ...interface{}) (*sqlen.EnhanceRow, error) {
	fixedSqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return nil, err
	}

	row := client.Exer.EnhancedQueryRowContext(ctx, fixedSqlText, args...)
	if err := row.Err(); err != nil {
		return nil, getExecutingSqlError(err, sqlText, fixedSqlText, args)
	}
	return row, nil
}

// Rows 用于获取查询结果行的游标对象。
// params:
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @row 返回目标行的 EnhanceRows 对象（是对 sql.Rows 的增强包装对象）。
//  @err 执行语句时遇到的错误。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) Rows(sqlText string, args ...interface{}) (*sqlen.EnhanceRows, error) {
	ctx, _ := client.getExecTimeoutContext()
	return client.RowsContext(ctx, sqlText, args...)
}

// RowsContext 用于获取查询结果行的游标对象。
// params:
//  @ctx context。
//  @sqlText SQL 语句，支持 @ 的命名参数占位及 @p1...@pn 这样的索引占位符。
//  @args SQL 语句的参数，支持通过 map[string]interface{} 提供命名参数值 或 通过变长参数提供索引参数值。
// returns:
//  @row 返回目标行的 EnhanceRows 对象（是对 sql.Rows 的增强包装对象）。
//  @err 执行语句时遇到的错误。
// 可以通过 errors.Is 判断的特殊 err：
//  - sqlmer.ErrParseParamFailed: 当 SQL 语句中的参数解析失败时返回该类错误。
//  - sqlmer.ErrExecutingSql: 当 SQL 语句执行时遇到错误，返回该类型错误。
func (client *AbstractDbClient) RowsContext(ctx context.Context, sqlText string, args ...interface{}) (*sqlen.EnhanceRows, error) {
	fixedSqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return nil, err
	}
	rows, err := client.Exer.EnhancedQueryContext(ctx, fixedSqlText, args...)
	if err != nil {
		return nil, getExecutingSqlError(err, sqlText, fixedSqlText, args)
	}
	return rows, nil
}
