package sqlen

import (
	"context"
	"database/sql"
)

var _ EnhancedDbExer = (*DbEnhance)(nil)

// DbEnhance 是对原生 sql.DB 的包装，除了原本的方法外，另外实现了 EnhancedDbExer 接口定义的额外方法。
type DbEnhance struct {
	*sql.DB
}

func NewDbEnhance(db *sql.DB) *DbEnhance {
	return &DbEnhance{db}
}

// EnhancedQueryRow executes a query that is expected to return at most one row.
// 返回增强后的 EnhanceRow 对象，相比原生 sql.Row 提供了更强的数据读取能力。
func (db *DbEnhance) EnhancedQueryRow(query string, args ...interface{}) *EnhanceRow {
	return db.EnhancedQueryRowContext(context.Background(), query, args...)
}

// EnhancedQueryRowContext executes a query that is expected to return at most one row.
// 返回增强后的 EnhanceRow 对象，相比原生 sql.Row 提供了更强的数据读取能力。
func (db *DbEnhance) EnhancedQueryRowContext(ctx context.Context, query string, args ...interface{}) *EnhanceRow {
	rows, err := db.EnhancedQueryContext(ctx, query, args...)
	return &EnhanceRow{rows: rows, err: err}
}

// EnhancedQuery executes a query that returns rows.
// 返回增强后的 EnhanceRows 对象，相比原生 sql.Rows 提供了更强的数据读取能力。
func (db *DbEnhance) EnhancedQuery(query string, args ...interface{}) (*EnhanceRows, error) {
	return db.EnhancedQueryContext(context.Background(), query, args...)
}

// EnhancedQueryContext executes a query that returns rows.
// 返回增强后的 EnhanceRows 对象，相比原生 sql.Rows 提供了更强的数据读取能力。
func (db *DbEnhance) EnhancedQueryContext(ctx context.Context, query string, args ...interface{}) (*EnhanceRows, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &EnhanceRows{rows, nil, nil, nil}, nil
}
