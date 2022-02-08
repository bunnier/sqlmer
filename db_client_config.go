package sqlmer

import (
	"context"
	"database/sql"
	"time"
)

type DbClientConfig struct {
	context     context.Context // 上下文对象，用于默认的超时、优雅关闭等控制。
	connTimeout time.Duration   // 数据库连接超时时间。
	execTimeout time.Duration   // 语句执行超时时间。

	driver           string  // 数据库驱动名称。
	connectionString string  // 连接字符串。
	db               *sql.DB // 数据库对象。

	bindArgsFunc BindSqlArgsFunc // 用于处理 sql 语句和所给的参数。
}

// NewDbClientConfig 创建一个数据库连接配置。
func NewDbClientConfig(options ...DbClientOption) (*DbClientConfig, error) {
	config := &DbClientConfig{
		context.Background(),
		time.Second * 30,
		time.Second * 30,
		"",
		"",
		nil,
		func(s string, i ...interface{}) (string, []interface{}, error) {
			return s, i, nil
		},
	}

	var err error
	for _, option := range options {
		if err = option(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// DbClientOption 是 DbClientConfig 的可选配置。
type DbClientOption func(config *DbClientConfig) error

// WithExecTimeout 用于为 DbClientConfig 设置默认的执行超时时间。
func WithExecTimeout(timeout time.Duration) DbClientOption {
	return func(config *DbClientConfig) error {
		config.execTimeout = timeout
		return nil
	}
}

// WithConnTimeout 用于为 DbClientConfig 设置获取数据库连接的超时时间。
func WithConnTimeout(timeout time.Duration) DbClientOption {
	return func(config *DbClientConfig) error {
		config.connTimeout = timeout
		return nil
	}
}

// WithConnTimeout 用于用现有的 sql.DB 初始化 DbClientOption。
func WithDbFunc(db *sql.DB) DbClientOption {
	return func(config *DbClientConfig) error {
		config.db = db
		return nil
	}
}

// WithConnTimeout 用于用现有的 sql.DB 初始化 DbClientOption。
func WithConnectionStringFunc(driver string, connectionString string) DbClientOption {
	return func(config *DbClientConfig) error {
		config.connectionString = connectionString
		config.driver = driver
		return nil
	}
}

// BindSqlArgsFunc 定义用于预处理 sql 语句与参数的函数。
type BindSqlArgsFunc func(string, ...interface{}) (string, []interface{}, error)

// WithBindArgsFunc 用于为 DbClientConfig 设置处理参数的函数。
func WithBindArgsFunc(argsFunc BindSqlArgsFunc) DbClientOption {
	return func(config *DbClientConfig) error {
		config.bindArgsFunc = argsFunc
		return nil
	}
}
