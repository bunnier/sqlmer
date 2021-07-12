package sqlen

import (
	"context"
	"database/sql"
)

var _ EnhancedDbExer = (*sqlen)(nil)
var _ EnhancedDbExer = (*TxEnhance)(nil)

// sqlen是对原生sql.DB的包装，除了原本的方法外，另外实现了EnhancedDbExer接口定义的额外方法。
type sqlen struct {
	*sql.DB
}

// sqlen是对原生sql.Tx的包装，除了原本的方法外，另外实现了EnhancedDbExer接口定义的额外方法。
type TxEnhance struct {
	*sql.Tx
}

func Newsqlen(db *sql.DB) *sqlen {
	return &sqlen{db}
}

func NewTxEnhance(tx *sql.Tx) *TxEnhance {
	return &TxEnhance{tx}
}

// EnhancedQueryRow executes a query that is expected to return at most one row.
// 返回增强后的EnhanceRow对象，相比原生sql.Row提供了更强的数据读取能力。
func (db *sqlen) EnhancedQueryRow(query string, args ...interface{}) *EnhanceRow {
	return db.EnhancedQueryRowContext(context.Background(), query, args...)
}

// EnhancedQueryRowContext executes a query that is expected to return at most one row.
// 返回增强后的EnhanceRow对象，相比原生sql.Row提供了更强的数据读取能力。
func (db *sqlen) EnhancedQueryRowContext(ctx context.Context, query string, args ...interface{}) *EnhanceRow {
	rows, err := db.EnhancedQueryContext(ctx, query, args...)
	return &EnhanceRow{rows: rows, err: err}
}

// EnhancedQuery executes a query that returns rows.
// 返回增强后的EnhanceRows对象，相比原生sql.Rows提供了更强的数据读取能力。
func (db *sqlen) EnhancedQuery(query string, args ...interface{}) (*EnhanceRows, error) {
	return db.EnhancedQueryContext(context.Background(), query, args...)
}

// EnhancedQueryContext executes a query that returns rows.
// 返回增强后的EnhanceRows对象，相比原生sql.Rows提供了更强的数据读取能力。
func (db *sqlen) EnhancedQueryContext(ctx context.Context, query string, args ...interface{}) (*EnhanceRows, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &EnhanceRows{rows, nil, nil, nil}, nil
}

// EnhancedQueryRow executes a query that is expected to return at most one row.
// 返回增强后的EnhanceRow对象，相比原生sql.Row提供了更强的数据读取能力。
func (tx *TxEnhance) EnhancedQueryRow(query string, args ...interface{}) *EnhanceRow {
	return tx.EnhancedQueryRowContext(context.Background(), query, args...)
}

// EnhancedQueryRowContext executes a query that is expected to return at most one row.
// 返回增强后的EnhanceRow对象，相比原生sql.Row提供了更强的数据读取能力。
func (tx *TxEnhance) EnhancedQueryRowContext(ctx context.Context, query string, args ...interface{}) *EnhanceRow {
	rows, err := tx.EnhancedQueryContext(ctx, query, args...)
	return &EnhanceRow{rows: rows, err: err}
}

// EnhancedQuery executes a query that returns rows.
// 返回增强后的EnhanceRows对象，相比原生sql.Rows提供了更强的数据读取能力。
func (tx *TxEnhance) EnhancedQuery(query string, args ...interface{}) (*EnhanceRows, error) {
	return tx.EnhancedQueryContext(context.Background(), query, args...)
}

// EnhancedQueryContext executes a query that returns rows.
// 返回增强后的EnhanceRows对象，相比原生sql.Rows提供了更强的数据读取能力。
func (tx *TxEnhance) EnhancedQueryContext(ctx context.Context, query string, args ...interface{}) (*EnhanceRows, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &EnhanceRows{rows, nil, nil, nil}, nil
}
