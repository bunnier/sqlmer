package sqlmer

import (
	"database/sql"
	"reflect"
)

var _ DbClient = (*MsSqlDbClient)(nil)

// MsSqlDbClient 是针对SqlServer的DbClient实现。
type MsSqlDbClient struct {
	internalDbClient
}

// NewMsSqlDbClient 用于创建一个MsSqlDbClient。
func NewMsSqlDbClient(connectionString string, options ...DbClientOption) (*MsSqlDbClient, error) {
	options = append(options, WithBindArgsFunc(bindMsSqlArgs)) // SqlServer要支持命名参数，需要定制一个参数解析函数。
	config := NewDbClientConfig(SqlServeDriver, connectionString, options...)
	internalDbClient, err := newInternalDbClient(config)

	if err != nil {
		return nil, err
	}

	return &MsSqlDbClient{
		internalDbClient,
	}, nil
}

// bindMsSqlArgs 用于对sql语句和参数进行预处理。
// 第一个参数如果是map，且仅且只有一个参数的情况下，做命名参数处理；其余情况做位置参数处理。
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
