package sqlmer

import (
	"context"
	"database/sql"
	"reflect"
	"time"

	"github.com/bunnier/sqlmer/sqlen"
)

type DbClientConfig struct {
	context       context.Context // 上下文对象，用于默认的超时、优雅关闭等控制。
	connTimeout   time.Duration   // 数据库连接超时时间。
	execTimeout   time.Duration   // 语句执行超时时间。
	withPingCheck bool            // 用于指定是否在 DbClient 初始化时候进行 ping 操作。

	Driver string  // 数据库驱动名称。
	Dsn    string  // 连接字符串。
	Db     *sql.DB // 数据库对象。

	bindArgsFunc      BindSqlArgsFunc       // 用于处理 sql 语句和所给的参数。
	getScanTypeFunc   sqlen.GetScanTypeFunc // 用于根据列信息获取用于 Scan 的类型。
	unifyDataTypeFunc sqlen.UnifyDataTypeFn // 用于统一不同驱动在 Go 中的映射类型。
}

// NewDbClientConfig 创建一个数据库连接配置。
func NewDbClientConfig(options ...DbClientOption) (*DbClientConfig, error) {
	config := &DbClientConfig{
		context:       context.Background(),
		connTimeout:   time.Second * 30,
		execTimeout:   time.Second * 30,
		withPingCheck: false,
		Driver:        "",
		Dsn:           "",
		Db:            nil,
		bindArgsFunc: func(s string, i ...any) (string, []any, error) {
			return s, i, nil
		},
		getScanTypeFunc: func(columnType *sql.ColumnType) reflect.Type {
			return columnType.ScanType()
		},
		unifyDataTypeFunc: func(columnType *sql.ColumnType, dest *any) {},
	}

	var err error
	for _, option := range options {
		if err = option(config); err != nil {
			return nil, err
		}
	}

	// 为 bindArgsFunc 注入参数合并逻辑。
	oriBindArgsFunc := config.bindArgsFunc
	config.bindArgsFunc = func(s string, i ...any) (string, []any, error) {
		i, err := preHandleArgs(i...) // 进行 结构体/map/索引 等各种参数的合并处理。
		if err != nil {
			return "", nil, err
		}

		return oriBindArgsFunc(s, i...)
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

// WithDb 用于用现有的 sql.DB 初始化 DbClientOption。
func WithDb(db *sql.DB, driver string, dsn string) DbClientOption {
	return func(config *DbClientConfig) error {
		config.Db = db
		config.Dsn = dsn
		config.Driver = driver
		return nil
	}
}

// WithDsn 用于用现有的 sql.DB 初始化 DbClientOption。
func WithDsn(driver string, dsn string) DbClientOption {
	return func(config *DbClientConfig) error {
		config.Dsn = dsn
		config.Driver = driver
		return nil
	}
}

// WithPingCheck 用于选择是否在初始化 DbClient 时候进行 ping 操作（默认为 false）。
func WithPingCheck(withPingCheck bool) DbClientOption {
	return func(config *DbClientConfig) error {
		config.withPingCheck = withPingCheck
		return nil
	}
}

// WithUnifyDataTypeFunc 用于为 DbClient 注入驱动相关的类型转换逻辑。
func WithUnifyDataTypeFunc(unifyDataType sqlen.UnifyDataTypeFn) DbClientOption {
	return func(config *DbClientConfig) error {
		config.unifyDataTypeFunc = unifyDataType
		return nil
	}
}

// BindSqlArgsFunc 定义用于预处理 sql 语句与参数的函数。
type BindSqlArgsFunc func(string, ...any) (string, []any, error)

// WithBindArgsFunc 用于为 DbClientConfig 设置处理参数的函数。
func WithBindArgsFunc(argsFunc BindSqlArgsFunc) DbClientOption {
	return func(config *DbClientConfig) error {
		config.bindArgsFunc = argsFunc
		return nil
	}
}

// WithBindArgsFunc 用于为 DbClientConfig 设置根据列信息获取 Scan 类型的函数。
func WithGetScanTypeFunc(scanFunc sqlen.GetScanTypeFunc) DbClientOption {
	return func(config *DbClientConfig) error {
		config.getScanTypeFunc = scanFunc
		return nil
	}
}
