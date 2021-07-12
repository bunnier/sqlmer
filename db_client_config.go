package sqlmer

import (
	"context"
	"time"
)

type DbClientConfig struct {
	context          context.Context // 上下文对象，用于默认的超时、优雅关闭等控制。
	connTimeout      time.Duration   // 数据库连接超时时间。
	execTimeout      time.Duration   // 语句执行超时时间。
	driver           string          // 数据库驱动名称。
	connectionString string          // 连接字符串。
	bindArgsFunc     BindSqlArgsFunc // 用于处理sql语句和所给的参数。
}

// NewDbClientConfig 创建一个数据库连接配置。
func NewDbClientConfig(driver string, connectionString string, options ...DbClientOption) *DbClientConfig {
	config := &DbClientConfig{
		context.Background(),
		time.Second * 30,
		time.Second * 30,
		driver,
		connectionString,
		func(s string, i ...interface{}) (string, []interface{}, error) {
			return s, i, nil
		},
	}

	for _, option := range options {
		option(config)
	}

	return config
}

// DbClientConfig 的可选配置。
type DbClientOption func(config *DbClientConfig)

// WithExecTimeout 用于为DbClientConfig设置默认的执行超时时间。
func WithExecTimeout(timeout time.Duration) DbClientOption {
	return func(config *DbClientConfig) {
		config.execTimeout = timeout
	}
}

// WithConnTimeout 用于为DbClientConfig设置获取数据库连接的超时时间。
func WithConnTimeout(timeout time.Duration) DbClientOption {
	return func(config *DbClientConfig) {
		config.connTimeout = timeout
	}
}

// BindSqlArgsFunc 定义用于预处理sql语句与参数的函数。
type BindSqlArgsFunc func(string, ...interface{}) (string, []interface{}, error)

// WithBindArgsFunc 用于为DbClientConfig设置处理参数的函数。
func WithBindArgsFunc(argsFunc BindSqlArgsFunc) DbClientOption {
	return func(config *DbClientConfig) {
		config.bindArgsFunc = argsFunc
	}
}
