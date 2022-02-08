package sqlmer

import (
	"context"
	"database/sql"

	"github.com/bunnier/sqlmer/internal/sqlen"
	"github.com/pkg/errors"
)

var _ DbClient = (*AbstractDbClient)(nil)

// AbstractDbClient 是一个 DbClient 的抽象实现。
type AbstractDbClient struct {
	config *DbClientConfig      // 存储数据库连接配置。
	SqlDB  *sql.DB              // 内部依赖的连接池。
	Exer   sqlen.EnhancedDbExer // 获取方法实际使用的执行对象。
}

// NewInternalDbClient 用于获取一个 internalDbClient 对象。
func NewInternalDbClient(config *DbClientConfig) (AbstractDbClient, error) {
	// 控制连接超时的 context。
	ctx, cancelFunc := context.WithTimeout(context.Background(), config.connTimeout)
	defer cancelFunc()

	db, err := getDb(ctx, config.driver, config.connectionString)
	if err != nil {
		return AbstractDbClient{}, err
	}

	return AbstractDbClient{
		config,
		db,
		sqlen.NewDbEnhance(db),
	}, nil
}

// 用于获取数据库连接池对象。
func getDb(ctx context.Context, driverName string, connectionString string) (*sql.DB, error) {
	db, err := sql.Open(driverName, connectionString) // 获取连接池。

	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil { // Open 操作并不会实际建立链接，需要 ping 一下，确保连接可用。
		return nil, err
	}

	return db, nil
}

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

	return &sqlxDbTransactionKeeper{
		txDbClient, tx, false, 0,
	}, nil
}

// ConnectionString 用于获取当前实例所使用的数据库连接字符串。
func (client *AbstractDbClient) ConnectionString() string {
	return client.config.connectionString
}

// getExecTimeoutContext 用于获取数据库语句默认超时 context。
func (client *AbstractDbClient) getExecTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), client.config.execTimeout)
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

// Exists 用于判断给定的查询的结果是否至少包含1行。
func (client *AbstractDbClient) Exists(sqlText string, args ...interface{}) (bool, error) {
	ctx, cancelFunc := client.getExecTimeoutContext()
	defer cancelFunc()
	return client.ExistsContext(ctx, sqlText, args...)
}

// ExistsContext 用于判断给定的查询的结果是否至少包含1行。
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

	if rows.Next() {
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
