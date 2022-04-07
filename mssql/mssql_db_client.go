package mssql

import (
	"database/sql"
	"reflect"

	"github.com/bunnier/sqlmer"

	_ "github.com/denisenkom/go-mssqldb"
)

// DriverName 是 SqlServer 驱动名称。
const DriverName = "sqlserver"

var _ sqlmer.DbClient = (*MsSqlDbClient)(nil)

// MsSqlDbClient 是针对 SqlServer 的 DbClient 实现。
type MsSqlDbClient struct {
	sqlmer.AbstractDbClient
}

// NewMsSqlDbClient 用于创建一个 MsSqlDbClient。
func NewMsSqlDbClient(connectionString string, options ...sqlmer.DbClientOption) (*MsSqlDbClient, error) {
	options = append(options,
		sqlmer.WithConnectionString(DriverName, connectionString),
		sqlmer.WithUnifyDataTypeFunc(unifyDataType),
		sqlmer.WithBindArgsFunc(bindMsSqlArgs)) // SqlServer 要支持命名参数，需要定制一个参数解析函数。
	config, err := sqlmer.NewDbClientConfig(options...)
	if err != nil {
		return nil, err
	}

	internalDbClient, err := sqlmer.NewInternalDbClient(config)
	if err != nil {
		return nil, err
	}

	return &MsSqlDbClient{
		internalDbClient,
	}, nil
}

// unifyDataType 用于统一数据类型。
func unifyDataType(columnType *sql.ColumnType, dest *interface{}) {
	switch columnType.DatabaseTypeName() {
	case "DECIMAL", "SMALLMONEY", "MONEY":
		switch v := (*dest).(type) {
		case []byte:
			if v == nil {
				*dest = nil
			}
			*dest = string(v)
		case *string:
			*dest = v
		case nil:
			*dest = nil
		}
	}
}

// bindMsSqlArgs 用于对 sql 语句和参数进行预处理。
// 第一个参数如果是 map，且仅且只有一个参数的情况下，做命名参数处理；其余情况做位置参数处理。
func bindMsSqlArgs(sqlText string, args ...interface{}) (string, []interface{}, error) {
	if len(args) != 1 || reflect.ValueOf(args[0]).Kind() != reflect.Map {
		return sqlText, args, nil
	}

	mapArgs := args[0].(map[string]interface{})
	namedArgs := make([]interface{}, 0, len(mapArgs))
	for name, value := range mapArgs {
		namedArgs = append(namedArgs, sql.Named(name, value))
	}
	return sqlText, namedArgs, nil
}
