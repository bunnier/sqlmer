package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/qm_namedsql"
	"github.com/bunnier/sqlmer/sqlen"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

// DriverName 是 MySql 驱动名称。
const DriverName = "mysql"

var _ sqlmer.DbClient = (*MySqlDbClient)(nil)

// MySqlDbClient 是针对 MySql 的 DbClient 实现。
type MySqlDbClient struct {
	*sqlmer.AbstractDbClient
	dsnConfig *mysqlDriver.Config
}

// NewMySqlDbClient 用于创建一个 MySqlDbClient。
func NewMySqlDbClient(dsn string, options ...sqlmer.DbClientOption) (*MySqlDbClient, error) {
	var dsnConfig *mysqlDriver.Config
	var err error

	if dsnConfig, err = mysqlDriver.ParseDSN(dsn); err != nil {
		return nil, err
	}

	// 影响行数的处理，采用和 SQL Server 一样的逻辑，即：
	// UPDATE 时，只要找到行，即使原值和目标值一致，也作为影响到行处理。
	dsnConfig.ClientFoundRows = true
	dsn = dsnConfig.FormatDSN()

	fixedOptions := []sqlmer.DbClientOption{
		sqlmer.WithDsn(DriverName, dsn),
		sqlmer.WithGetScanTypeFunc(getScanTypeFn(dsnConfig)),        // 定制 Scan 类型逻辑。
		sqlmer.WithUnifyDataTypeFunc(getUnifyDataTypeFn(dsnConfig)), // 定制类型转换逻辑。
		sqlmer.WithBindArgsFunc(bindArgs),                           // 定制参数绑定逻辑。
	}
	options = append(fixedOptions, options...) // 用户自定义选项放后面，以覆盖默认。

	config, err := sqlmer.NewDbClientConfig(options...)
	if err != nil {
		return nil, err
	}

	absDbClient, err := sqlmer.NewAbstractDbClient(config)
	if err != nil {
		return nil, err
	}

	// 自动设置连接池的超时时间，
	err = autoSetConnMaxLifetime(absDbClient.Db)
	if err != nil {
		return nil, err
	}

	return &MySqlDbClient{absDbClient, dsnConfig}, nil
}

// autoSetConnMaxLifetime 用于从数据库读取 wait_timeout 设置，并自动设置连接最大生命周期。
// 当数据库的超时时间有容错时间时，连接的使用时间短一些，，以尽量在数据库掐断连接之前主动放弃连接。
func autoSetConnMaxLifetime(db *sqlen.DbEnhance) error {
	timeSettingRows := db.EnhancedQueryRow(`SHOW VARIABLES WHERE Variable_name IN ('wait_timeout')`) // 连接测试。
	if timeSettingRows.Err() != nil {
		return fmt.Errorf("%w: get 'wait_timeout' variable error: %w", sqlmer.ErrConnect, timeSettingRows.Err())
	}

	var waitTimeoutName string
	var waitTimeout int
	err := timeSettingRows.Scan(&waitTimeoutName, &waitTimeout)
	if err != nil {
		return fmt.Errorf("%w: get 'wait_timeout' variable error: %w", sqlmer.ErrConnect, err)
	}

	// 根据 wait_timeout 设置连接最大生命周期。
	// SetConnMaxLifetime 小于 wait_timeout 应该是最佳实践，可以避免数据库强制关闭连接导致的错误（driver: bad connection）。
	// 参考：
	// https://go.dev/doc/database/manage-connections
	// https://github.com/go-sql-driver/mysql?tab=readme-ov-file#important-settings
	switch {
	case waitTimeout == 0:
		return nil

	case waitTimeout > 60:
		waitTimeout = waitTimeout - 5

	case waitTimeout > 1:
		waitTimeout = waitTimeout - 1
	}

	db.SetConnMaxLifetime(time.Duration(waitTimeout) * time.Second)
	return nil
}

// bindArgs 用于对 SQL 语句和参数进行预处理。
// 第一个参数如果是 map，且仅且只有一个参数的情况下，做命名参数处理，其余情况做位置参数处理。
func bindArgs(sqlText string, args ...any) (string, []any, error) {
	return qm_namedsql.BindQuestionMarkArgs(sqlText, args...)
}

// getScanTypeFn 根据驱动配置返回一个可以正确获取 Scan 类型的函数。
func getScanTypeFn(cfg *mysqlDriver.Config) sqlen.GetScanTypeFunc {
	var scanTypeRawBytes = reflect.TypeOf(sql.RawBytes{})
	return func(columnType *sql.ColumnType) reflect.Type {
		if !cfg.ParseTime && isTimeColumn(columnType.DatabaseTypeName()) {
			return scanTypeRawBytes // MySql 的驱动，如果没有开启 ParseTime，只能通过 sql.RawBytes 进行 Scan，否则会失败。
		}
		return columnType.ScanType()
	}
}

// isTimeColumn 判断某个列是否是时间类型。
func isTimeColumn(colTypeName string) bool {
	return colTypeName == "TIMESTAMP" ||
		colTypeName == "TIME" ||
		colTypeName == "DATETIME" ||
		colTypeName == "DATE"
}

// getUnifyDataTypeFn 根据驱动配置返回一个统一处理数据类型的函数。
func getUnifyDataTypeFn(cfg *mysqlDriver.Config) sqlen.UnifyDataTypeFn {
	return func(columnType *sql.ColumnType, dest *any) {
		switch columnType.DatabaseTypeName() {
		case "VARCHAR", "CHAR", "TEXT", "DECIMAL":
			switch v := (*dest).(type) {
			case sql.RawBytes:
				if v == nil {
					*dest = nil
					break
				}
				*dest = string(v)
			}

		case "TIMESTAMP", "TIME", "DATETIME", "DATE":
			if cfg.ParseTime {
				break // 如果驱动开启了 ParseTime，就不需要再进行转换了。
			}
			switch v := (*dest).(type) {
			case sql.RawBytes:
				if v == nil {
					*dest = nil
					break
				}
				// 用 RawBytes 接收的 Time 系列类型，需要转为 time.Time{}，下面这个时间格式是从 MySQL 驱动里拷来的。
				const timeFormat = "2006-01-02 15:04:05.999999"
				timeStr := string(v)
				if time, err := time.ParseInLocation(timeFormat[:len(timeStr)], timeStr, cfg.Loc); err != nil {
					panic(err)
				} else {
					*dest = time
				}
			}

		default: // 将 sql.RawBytes 统一转为 []byte。
			switch v := (*dest).(type) {
			case sql.RawBytes:
				if v == nil {
					*dest = nil
					break
				}
				*dest = []byte(v)
			case *any:
				if v == nil { // 直接 SELECT null 时候会使用 *any 进行 Scan 并在最后返回一个指向 nil 的值，这里将其统一到无类型 nil。
					*dest = nil
					break
				}
				*dest = v
			}

		}
	}
}
