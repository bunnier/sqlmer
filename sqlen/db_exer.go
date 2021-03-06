package sqlen

import (
	"context"
	"database/sql"
)

// DbExer 是从 sql.Tx 和 sql.Db 对象上提取的公共数据库操纵接口。
type DbExer interface {

	// QueryRow executes a query that is expected to return at most one row.
	// QueryRow always returns a non-nil value. Errors are deferred until
	// Row's Scan method is called.
	// If the query selects no rows, the *Row's Scan will return ErrNoRows.
	// Otherwise, the *Row's Scan scans the first selected row and discards
	// the rest.
	QueryRow(query string, args ...any) *sql.Row

	// QueryRowContext executes a query that is expected to return at most one row.
	// QueryRowContext always returns a non-nil value. Errors are deferred until
	// Row's Scan method is called.
	// If the query selects no rows, the *Row's Scan will return ErrNoRows.
	// Otherwise, the *Row's Scan scans the first selected row and discards
	// the rest.
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row

	// Query executes a query that returns rows, typically a SELECT.
	// The args are for any placeholder parameters in the query.
	Query(query string, args ...any) (*sql.Rows, error)

	// QueryContext executes a query that returns rows, typically a SELECT.
	// The args are for any placeholder parameters in the query.
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	// Exec executes a query that doesn't return rows.
	// For example: an INSERT and UPDATE.
	Exec(query string, args ...any) (sql.Result, error)

	// ExecContext executes a query without returning any rows.
	// The args are for any placeholder parameters in the query.
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type EnhancedDbExer interface {
	DbExer

	// EnhancedQueryRow executes a query that is expected to return at most one row.
	// 返回增强后的 EnhanceRow 对象，相比原生 sql.Row 提供了更强的数据读取能力。
	EnhancedQueryRow(query string, args ...any) *EnhanceRow

	// EnhancedQueryRowContext executes a query that is expected to return at most one row.
	// 返回增强后的 EnhanceRow 对象，相比原生 sql.Row 提供了更强的数据读取能力。
	EnhancedQueryRowContext(ctx context.Context, query string, args ...any) *EnhanceRow

	// EnhancedQuery executes a query that returns rows.
	// 返回增强后的 EnhanceRows 对象，相比原生 sql.Rows 提供了更强的数据读取能力。
	EnhancedQuery(query string, args ...any) (*EnhanceRows, error)

	// EnhancedQueryContext executes a query that returns rows.
	// 返回增强后的 EnhanceRows 对象，相比原生 sql.Rows 提供了更强的数据读取能力。
	EnhancedQueryContext(ctx context.Context, query string, args ...any) (*EnhanceRows, error)
}
