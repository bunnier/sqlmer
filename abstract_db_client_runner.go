package sqlmer

import (
	"context"
	"database/sql"

	"github.com/bunnier/sqlmer/sqlen"
)

// bindAndExecContext 用于统一处理参数绑定、执行 SQL 与执行错误包装。
func (client *AbstractDbClient) bindAndExecContext(ctx context.Context, rawSql string, args ...any) (sql.Result, string, []any, error) {
	fixedSql, fixedArgs, err := client.config.bindArgsFunc(rawSql, args...)
	if err != nil {
		return nil, "", nil, err
	}

	result, err := client.Exer.ExecContext(ctx, fixedSql, fixedArgs...)
	if err != nil {
		return nil, "", nil, getExecutingSqlError(err, rawSql, fixedSql, fixedArgs)
	}

	return result, fixedSql, fixedArgs, nil
}

// bindAndQueryRowsContext 用于统一处理参数绑定、查询游标与执行错误包装。
func (client *AbstractDbClient) bindAndQueryRowsContext(ctx context.Context, rawSql string, args ...any) (*sqlen.EnhanceRows, string, []any, error) {
	fixedSql, fixedArgs, err := client.config.bindArgsFunc(rawSql, args...)
	if err != nil {
		return nil, "", nil, err
	}

	rows, err := client.Exer.EnhancedQueryContext(ctx, fixedSql, fixedArgs...)
	if err != nil {
		return nil, "", nil, getExecutingSqlError(err, rawSql, fixedSql, fixedArgs)
	}

	return rows, fixedSql, fixedArgs, nil
}

// bindAndQueryRowContext 用于统一处理参数绑定、单行查询对象创建与执行错误包装。
func (client *AbstractDbClient) bindAndQueryRowContext(ctx context.Context, rawSql string, args ...any) (*sqlen.EnhanceRow, string, []any, error) {
	fixedSql, fixedArgs, err := client.config.bindArgsFunc(rawSql, args...)
	if err != nil {
		return nil, "", nil, err
	}

	row := client.Exer.EnhancedQueryRowContext(ctx, fixedSql, fixedArgs...)
	return row, fixedSql, fixedArgs, nil
}
