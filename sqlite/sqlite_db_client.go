package sqlite

import (
	"database/sql"
	"reflect"
	"strings"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/named2qm"
	"github.com/bunnier/sqlmer/sqlen"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// DriverName 是 Sqlite 驱动名称。
const DriverName = "sqlite3"

var _ sqlmer.DbClient = (*SqliteDbClient)(nil)

// SqliteDbClient 是针对 Sqlite 的 DbClient 实现。
type SqliteDbClient struct {
	*sqlmer.AbstractDbClient
}

// NewSqliteDbClient 用于创建一个 SqliteDbClient。
func NewSqliteDbClient(dsn string, options ...sqlmer.DbClientOption) (*SqliteDbClient, error) {
	fixedOptions := []sqlmer.DbClientOption{
		sqlmer.WithDsn(DriverName, dsn),
		sqlmer.WithGetScanTypeFunc(getScanTypeFn()),        // 定制 Scan 类型逻辑。
		sqlmer.WithUnifyDataTypeFunc(getUnifyDataTypeFn()), // 定制类型转换逻辑。
		sqlmer.WithBindArgsFunc(bindArgs),                  // 定制参数绑定逻辑。
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

	return &SqliteDbClient{absDbClient}, nil
}

// bindArgs 用于对 SQL 语句和参数进行预处理。
// 第一个参数如果是 map，且仅且只有一个参数的情况下，做命名参数处理，其余情况做位置参数处理。
func bindArgs(sqlText string, args ...any) (string, []any, error) {
	return named2qm.BindQuestionMarkArgs(sqlText, args...)
}

// getScanTypeFn 根据驱动配置返回一个可以正确获取 Scan 类型的函数。
func getScanTypeFn() sqlen.GetScanTypeFunc {
	return func(columnType *sql.ColumnType) reflect.Type {
		// 如果 ScanType 返回 nil，使用 interface{} 或者 sql.RawBytes。
		// 使用 interface{} 可以让驱动决定返回什么类型。
		t := columnType.ScanType()
		if t == nil {
			return reflect.TypeOf(new(any)).Elem()
		}
		return t
	}
}

// getUnifyDataTypeFn 根据驱动配置返回一个统一处理数据类型的函数。
func getUnifyDataTypeFn() sqlen.UnifyDataTypeFn {
	return func(columnType *sql.ColumnType, dest *any) {
		typeName := strings.ToUpper(columnType.DatabaseTypeName())

		// 处理字符串类型。
		if strings.Contains(typeName, "CHAR") || strings.Contains(typeName, "TEXT") || strings.Contains(typeName, "CLOB") {
			switch v := (*dest).(type) {
			case sql.RawBytes:
				if v == nil {
					*dest = nil
					break
				}
				*dest = string(v)
			case []byte: // 某些情况下可能是 []byte
				if v == nil {
					*dest = nil
					break
				}
				*dest = string(v)
			case string:
				// 已经是 string 了，不需要处理
			}
			return
		}

		// 处理时间类型。
		if strings.Contains(typeName, "TIME") || strings.Contains(typeName, "DATE") {
			switch v := (*dest).(type) {
			case string:
				// 尝试解析常见时间格式。
				// 这里假设是标准格式，如果不是，可能需要更多逻辑。
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					*dest = t
				} else if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
					*dest = t
				} else if t, err := time.Parse("2006-01-02 15:04:05.999999999", v); err == nil {
					*dest = t
				} else if t, err := time.Parse("2006-01-02", v); err == nil {
					*dest = t
				} else if t, err := time.Parse("15:04:05.999999999", v); err == nil {
					// Time only
					// Usually we return time.Time with date 0000-01-01 or similar?
					// Use a base date.
					baseDate, _ := time.Parse("2006-01-02", "0001-01-01")
					*dest = baseDate.Add(time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute + time.Duration(t.Second())*time.Second + time.Duration(t.Nanosecond()))
				}
				// 否则保留 string。
			case int64:
				// 假设是 Unix timestamp。
				*dest = time.Unix(v, 0)
			case []byte:
				s := string(v)
				if t, err := time.Parse(time.RFC3339, s); err == nil {
					*dest = t
				} else if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
					*dest = t
				} else if t, err := time.Parse("2006-01-02 15:04:05.999999999", s); err == nil {
					*dest = t
				}
			}
			return
		}

		// 默认处理，处理 sql.RawBytes。
		switch v := (*dest).(type) {
		case sql.RawBytes:
			if v == nil {
				*dest = nil
				break
			}
			*dest = []byte(v)
		case *any:
			if v == nil {
				*dest = nil
				break
			}
			*dest = v
		}
	}
}
