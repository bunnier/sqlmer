package sqlmer

import (
	"context"
	"database/sql"

	"github.com/bunnier/sqlmer/sqlen"
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

	// db 可能已经由 option 传入了。
	if config.db == nil {
		if config.driver == "" || config.connectionString == "" {
			return AbstractDbClient{}, errors.Wrap(ErrConnect, "driver or connectionString is empty")
		}

		var err error
		config.db, err = getDb(ctx, config.driver, config.connectionString)
		if err != nil {
			return AbstractDbClient{}, err
		}
	}

	return AbstractDbClient{
		config,
		config.db,
		sqlen.NewDbEnhance(config.db),
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

// getExecTimeoutContext 用于获取数据库语句默认超时 context。
func (client *AbstractDbClient) getExecTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), client.config.execTimeout)
}

// ConnectionString 用于获取当前实例所使用的数据库连接字符串。
func (client *AbstractDbClient) ConnectionString() string {
	return client.config.connectionString
}
