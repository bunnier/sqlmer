package sqlen

import (
	"context"
	"database/sql"
)

var _ EnhancedDbExer = (*TxEnhance)(nil)

// TxEnhance 是对原生 sql.Tx 的包装，除了原本的方法外，另外实现了 EnhancedDbExer 接口定义的额外方法。
type TxEnhance struct {
	*sql.Tx
	dbEnhance *DbEnhance // 用于统一不同驱动在 Go 中的映射类型。
}

func NewTxEnhance(tx *sql.Tx, dbEnhance *DbEnhance) *TxEnhance {
	return &TxEnhance{tx, dbEnhance}
}

// EnhancedQueryRow executes a query that is expected to return at most one row.
// 返回增强后的 EnhanceRow 对象，相比原生 sql.Row 提供了更强的数据读取能力。
func (tx *TxEnhance) EnhancedQueryRow(query string, args ...any) *EnhanceRow {
	return tx.EnhancedQueryRowContext(context.Background(), query, args...)
}

// EnhancedQueryRowContext executes a query that is expected to return at most one row.
// 返回增强后的EnhanceRow对象，相比原生sql.Row提供了更强的数据读取能力。
func (tx *TxEnhance) EnhancedQueryRowContext(ctx context.Context, query string, args ...any) *EnhanceRow {
	rows, err := tx.EnhancedQueryContext(ctx, query, args...)
	return &EnhanceRow{rows: rows, err: err}
}

// EnhancedQuery executes a query that returns rows.
// 返回增强后的 EnhanceRows 对象，相比原生 sql.Rows 提供了更强的数据读取能力。
func (tx *TxEnhance) EnhancedQuery(query string, args ...any) (*EnhanceRows, error) {
	return tx.EnhancedQueryContext(context.Background(), query, args...)
}

// EnhancedQueryContext executes a query that returns rows.
// 返回增强后的 EnhanceRows 对象，相比原生 sql.Rows 提供了更强的数据读取能力。
func (tx *TxEnhance) EnhancedQueryContext(ctx context.Context, query string, args ...any) (*EnhanceRows, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &EnhanceRows{
		Rows:          rows,
		getScanTypeFn: tx.dbEnhance.getScanTypeFn,
		unifyDataType: tx.dbEnhance.unifyDataType,
	}, nil
}
