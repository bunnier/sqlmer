package sqlmer

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bunnier/sqlmer/internal/sqlen"
)

var _ DbClient = (*internalDbClient)(nil)

// internalDbClient 是通过sqlx实现的DbClient结构。
type internalDbClient struct {
	config *DbClientConfig      // 存储数据库连接配置。
	SqlDB  *sql.DB              // 通过sqlx包装的数据库连接池。
	Exer   sqlen.EnhancedDbExer // 获取方法实际使用的执行对象。
}

// 获取一个sqlxDbClient对象。
func newInternalDbClient(config *DbClientConfig) (internalDbClient, error) {
	// 控制连接超时的context。
	ctx, cancelFunc := context.WithTimeout(context.Background(), config.connTimeout)
	defer cancelFunc()

	db, err := getDb(ctx, config.driver, config.connectionString)
	if err != nil {
		return internalDbClient{}, err
	}

	return internalDbClient{
		config,
		db,
		sqlen.Newsqlen(db),
	}, nil
}

// CreateTransaction 用于开始一个事务。
func (client *internalDbClient) CreateTransaction() (TransactionKeeper, error) {
	tx, err := client.SqlDB.Begin()
	if err != nil {
		return nil, err
	}

	txDbClient := &internalDbClient{
		client.config,
		client.SqlDB,
		sqlen.NewTxEnhance(tx), // 新的client中的实际执行对象使用开启的事务。
	}

	return &sqlxDbTransactionKeeper{
		txDbClient, tx, false, 0,
	}, nil
}

// ConnectionString 用于获取当前实例所使用的数据库连接字符串。
func (client *internalDbClient) ConnectionString() string {
	return client.config.connectionString
}

// getExecTimeoutContext 用于获取数据库语句默认超时context。
func (client *internalDbClient) getExecTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), client.config.execTimeout)
}

// Scalar 用于获取查询的第一行第一列的值。
func (client *internalDbClient) Scalar(sqlText string, args ...interface{}) (interface{}, error) {
	context, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.ScalarContext(context, sqlText, args...)
}

// Scalar 用于获取查询的第一行第一列的值。
func (client *internalDbClient) ScalarContext(ctx context.Context, sqlText string, args ...interface{}) (interface{}, error) {
	sqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return nil, err
	}
	rows, err := client.Exer.EnhancedQueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		result, err := rows.SliceScan()
		if err != nil {
			return nil, err
		}
		return result[0], err // 只要执行成功，至少有1列的。
	}
	return nil, rows.Err()
}

// Execute 用于执行非查询SQL语句，并返回所影响的行数。
func (client *internalDbClient) Execute(sqlText string, args ...interface{}) (int64, error) {
	context, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.ExecuteContext(context, sqlText, args...)
}

// Execute 用于执行非查询SQL语句，并返回所影响的行数。
func (client *internalDbClient) ExecuteContext(ctx context.Context, sqlText string, args ...interface{}) (int64, error) {
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

// SizedExecute 用于执行非查询SQL语句，并断言所影响的行数。若影响的函数不正确，抛出异常。
func (client *internalDbClient) SizedExecute(expectedSize int64, sqlText string, args ...interface{}) error {
	context, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.SizedExecuteContext(context, expectedSize, sqlText, args...)
}

// SizedExecute 用于执行非查询SQL语句，并断言所影响的行数。若影响的函数不正确，抛出异常。
func (client *internalDbClient) SizedExecuteContext(ctx context.Context, expectedSize int64, sqlText string, args ...interface{}) error {
	affectedRow, err := client.ExecuteContext(ctx, sqlText, args...)
	if err != nil {
		return err
	}
	if affectedRow != expectedSize {
		return NewDbSqlError(fmt.Sprintf("affected rows expected: %d, acttually: %d. ", expectedSize, affectedRow), sqlText)
	}
	return nil
}

// Exists 用于判断给定的查询的结果是否至少包含1行。
func (client *internalDbClient) Exists(sqlText string, args ...interface{}) (bool, error) {
	context, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.ExistsContext(context, sqlText, args...)
}

// Exists 用于判断给定的查询的结果是否至少包含1行。
func (client *internalDbClient) ExistsContext(ctx context.Context, sqlText string, args ...interface{}) (bool, error) {
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
	return hasData, nil
}

// Get 用于获取查询结果的第一行记录。
func (client *internalDbClient) Get(sqlText string, args ...interface{}) (map[string]interface{}, error) {
	context, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.GetContext(context, sqlText, args...)
}

// GetContext 用于获取查询结果的第一行记录。
func (client *internalDbClient) GetContext(ctx context.Context, sqlText string, args ...interface{}) (map[string]interface{}, error) {
	sqlText, args, err := client.config.bindArgsFunc(sqlText, args...)
	if err != nil {
		return nil, err
	}
	rows, err := client.Exer.EnhancedQueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		result := make(map[string]interface{})
		err = rows.MapScan(result)
		return result, err
	}
	return nil, rows.Err()
}

// SliceGet 用于获取查询结果得行序列。
func (client *internalDbClient) SliceGet(sqlText string, args ...interface{}) ([]map[string]interface{}, error) {
	context, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.SliceGetContext(context, sqlText, args...)
}

// SliceGet 用于获取查询结果得行序列。
func (client *internalDbClient) SliceGetContext(ctx context.Context, sqlText string, args ...interface{}) ([]map[string]interface{}, error) {
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
	for rows.Next() {
		result := make(map[string]interface{})
		err = rows.MapScan(result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

// Rows 用于获取读取数据的游标sql.Rows。
func (client *internalDbClient) Rows(sqlText string, args ...interface{}) (*sqlen.EnhanceRows, error) {
	context, _ := client.getExecTimeoutContext()
	return client.RowsContext(context, sqlText, args...)
}

// RowsContext 用于获取读取数据的游标sql.Rows。
func (client *internalDbClient) RowsContext(ctx context.Context, sqlText string, args ...interface{}) (*sqlen.EnhanceRows, error) {
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
