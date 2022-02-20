package sqlmer

import (
	"context"

	"github.com/bunnier/sqlmer/sqlen"
	"github.com/pkg/errors"
)

var _ errorDbClient = (*AbstractDbClient)(nil)

// CreateTransaction 用于开始一个事务。
func (client *AbstractDbClient) CreateTransaction() (TransactionKeeper, error) {
	tx, err := client.SqlDB.Begin()
	if err != nil {
		return nil, err
	}

	txDbClient := &AbstractDbClient{
		client.config,
		client.SqlDB,
		sqlen.NewTxEnhance(tx), // 新的client中的实际执行对象使用开启的事务。
	}

	return &abstractTransactionKeeper{
		txDbClient, tx, false, 0,
	}, nil
}

// Execute 用于执行非查询SQL语句，并返回所影响的行数。
func (client *AbstractDbClient) Execute(sqlText string, args ...interface{}) (int64, error) {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.ExecuteContext(ctx, sqlText, args...)
}

// ExecuteContext 用于执行非查询 sql 语句，并返回所影响的行数。
func (client *AbstractDbClient) ExecuteContext(ctx context.Context, sqlText string, args ...interface{}) (int64, error) {
	sqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return 0, err
	}
	sqlResult, err := client.Exer.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return 0, err
	}
	return sqlResult.RowsAffected()
}

// SizedExecute 用于执行非查询 sql 语句，并断言所影响的行数。若影响的函数不正确，抛出异常。
func (client *AbstractDbClient) SizedExecute(expectedSize int64, sqlText string, args ...interface{}) error {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.SizedExecuteContext(ctx, expectedSize, sqlText, args...)
}

// SizedExecuteContext 用于执行非查询 sql 语句，并断言所影响的行数。若影响的函数不正确，抛出异常。
func (client *AbstractDbClient) SizedExecuteContext(ctx context.Context, expectedSize int64, sqlText string, args ...interface{}) error {
	affectedRow, err := client.ExecuteContext(ctx, sqlText, args...)
	if err != nil {
		return err
	}
	if affectedRow != expectedSize {
		return errors.WithMessagef(ErrSql, "affected rows expected: %d, acttually: %d, sql=%s", expectedSize, affectedRow, sqlText)
	}
	return nil
}

// Exists 用于判断给定的查询的结果是否至少包含 1 行。
func (client *AbstractDbClient) Exists(sqlText string, args ...interface{}) (bool, error) {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.ExistsContext(ctx, sqlText, args...)
}

// ExistsContext 用于判断给定的查询的结果是否至少包含 1 行。
func (client *AbstractDbClient) ExistsContext(ctx context.Context, sqlText string, args ...interface{}) (bool, error) {
	sqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return false, err
	}

	rows, err := client.Exer.EnhancedQueryContext(ctx, sqlText, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	hasData := rows.Next()
	if err = rows.Err(); err != nil {
		return false, err
	}
	return hasData, nil
}

// Scalar 用于获取查询的第一行第一列的值。
func (client *AbstractDbClient) Scalar(sqlText string, args ...interface{}) (interface{}, error) {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.ScalarContext(ctx, sqlText, args...)
}

// ScalarContext 用于获取查询的第一行第一列的值。
func (client *AbstractDbClient) ScalarContext(ctx context.Context, sqlText string, args ...interface{}) (interface{}, error) {
	sqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return nil, err
	}

	if result, err := client.Exer.EnhancedQueryRowContext(ctx, sqlText, args...).SliceScan(); err != nil {
		return nil, err
	} else {
		return result[0], nil // 只要没有 error，至少有 1 列的。
	}
}

// Get 用于获取查询结果的第一行记录。
func (client *AbstractDbClient) Get(sqlText string, args ...interface{}) (map[string]interface{}, error) {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.GetContext(ctx, sqlText, args...)
}

// GetContext 用于获取查询结果的第一行记录。
func (client *AbstractDbClient) GetContext(ctx context.Context, sqlText string, args ...interface{}) (map[string]interface{}, error) {
	sqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return nil, err
	}
	rows, err := client.Exer.EnhancedQueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() { // 这个地方不直接 EnhancedQueryRowContext.MapScan 主要是可以在没有行时候省略一次 make map。
		result := make(map[string]interface{})
		err = rows.MapScan(result)
		return result, err
	}
	return nil, rows.Err()
}

// SliceGet 用于获取查询结果得行序列。
func (client *AbstractDbClient) SliceGet(sqlText string, args ...interface{}) ([]map[string]interface{}, error) {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.SliceGetContext(ctx, sqlText, args...)
}

// SliceGetContext 用于获取查询结果得行序列。
func (client *AbstractDbClient) SliceGetContext(ctx context.Context, sqlText string, args ...interface{}) ([]map[string]interface{}, error) {
	sqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return nil, err
	}
	rows, err := client.Exer.EnhancedQueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
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
func (client *AbstractDbClient) Row(sqlText string, args ...interface{}) (*sqlen.EnhanceRow, error) {
	ctx, _ := client.getExecTimeoutContext()
	return client.RowContext(ctx, sqlText, args...)
}

// RowContext 用于获取单个查询结果行。
func (client *AbstractDbClient) RowContext(ctx context.Context, sqlText string, args ...interface{}) (*sqlen.EnhanceRow, error) {
	sqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return nil, err
	}

	row := client.Exer.EnhancedQueryRowContext(ctx, sqlText, args...)
	return row, row.Err()
}

// Rows 用于获取读取数据的游标 sql.Rows。
func (client *AbstractDbClient) Rows(sqlText string, args ...interface{}) (*sqlen.EnhanceRows, error) {
	ctx, _ := client.getExecTimeoutContext()
	return client.RowsContext(ctx, sqlText, args...)
}

// RowsContext 用于获取读取数据的游标 sql.Rows。
func (client *AbstractDbClient) RowsContext(ctx context.Context, sqlText string, args ...interface{}) (*sqlen.EnhanceRows, error) {
	sqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return nil, err
	}
	rows, err := client.Exer.EnhancedQueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
