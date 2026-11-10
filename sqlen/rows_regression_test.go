package sqlen_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/bunnier/sqlmer/sqlen"
)

type scanTestConnector struct {
	rows     *scanTestRows
	queryErr error
}

func (c *scanTestConnector) Connect(context.Context) (driver.Conn, error) {
	return &scanTestConn{connector: c}, nil
}

func (c *scanTestConnector) Driver() driver.Driver {
	return scanTestDriver{}
}

type scanTestDriver struct{}

func (scanTestDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("use connector")
}

type scanTestConn struct {
	connector *scanTestConnector
}

func (c *scanTestConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare unsupported")
}

func (c *scanTestConn) Close() error {
	return nil
}

func (c *scanTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions unsupported")
}

func (c *scanTestConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return c.connector.rows, c.connector.queryErr
}

type scanTestRows struct {
	values     []driver.Value
	offset     int
	nextErr    error
	closeErr   error
	closeCalls int
}

func (r *scanTestRows) Columns() []string {
	return []string{"value"}
}

func (r *scanTestRows) Close() error {
	r.closeCalls++
	return r.closeErr
}

func (r *scanTestRows) Next(dest []driver.Value) error {
	if r.offset < len(r.values) {
		dest[0] = r.values[r.offset]
		r.offset++
		return nil
	}
	if r.nextErr != nil {
		return r.nextErr
	}
	return io.EOF
}

func (r *scanTestRows) ColumnTypeScanType(int) reflect.Type {
	return reflect.TypeOf(int64(0))
}

func newScanTestDb(t *testing.T, rows *scanTestRows, queryErr error) *sqlen.DbEnhance {
	t.Helper()
	db := sql.OpenDB(&scanTestConnector{rows: rows, queryErr: queryErr})
	t.Cleanup(func() { _ = db.Close() })
	return sqlen.NewDbEnhance(db, func(c *sql.ColumnType) reflect.Type {
		return c.ScanType()
	}, func(*sql.ColumnType, *any) {})
}

func TestEnhanceRows_SliceScan_error_allows_retry_and_next(t *testing.T) {
	db := newScanTestDb(t, &scanTestRows{values: []driver.Value{"bad", "123"}}, nil)
	rows, err := db.EnhancedQuery("SELECT value")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatal("expected first row")
	}
	if _, err = rows.SliceScan(); err == nil {
		t.Fatal("expected conversion error")
	}
	if err = rows.Err(); err != nil {
		t.Fatalf("Err() after conversion failure = %v, want nil", err)
	}
	var text string
	if err = rows.Scan(&text); err != nil || text != "bad" {
		t.Fatalf("Scan() retry = %q, %v, want bad, nil", text, err)
	}
	if !rows.Next() {
		t.Fatal("expected second row")
	}
	value, err := rows.SliceScan()
	if err != nil || !reflect.DeepEqual(value, []any{int64(123)}) {
		t.Fatalf("SliceScan() = %v, %v, want [123], nil", value, err)
	}
	if rows.Next() || rows.Err() != nil {
		t.Fatalf("expected clean EOF, got Err() = %v", rows.Err())
	}
	if err = rows.Close(); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}
}

func TestEnhanceRows_MapScan_error_allows_next(t *testing.T) {
	db := newScanTestDb(t, &scanTestRows{values: []driver.Value{"bad", "123"}}, nil)
	rows, err := db.EnhancedQuery("SELECT value")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatal("expected first row")
	}
	if _, err = rows.MapScan(); err == nil {
		t.Fatal("expected conversion error")
	}
	if err = rows.Err(); err != nil {
		t.Fatalf("Err() after conversion failure = %v, want nil", err)
	}
	if !rows.Next() {
		t.Fatal("expected second row")
	}
	value, err := rows.MapScan()
	if err != nil || !reflect.DeepEqual(value, map[string]any{"value": int64(123)}) {
		t.Fatalf("MapScan() = %v, %v, want value=123, nil", value, err)
	}
}

func TestEnhanceRows_metadata_error_does_not_affect_Err(t *testing.T) {
	db := newScanTestDb(t, &scanTestRows{}, nil)
	rows, err := db.EnhancedQuery("SELECT value")
	if err != nil {
		t.Fatal(err)
	}
	if err = rows.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = rows.SliceScan(); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("SliceScan() error = %v, want closed rows error", err)
	}
	if err = rows.Err(); err != nil {
		t.Fatalf("Err() after metadata failure = %v, want nil", err)
	}
}

func TestEnhanceRows_iteration_error_is_preserved(t *testing.T) {
	iterationErr := errors.New("iteration failed")
	db := newScanTestDb(t, &scanTestRows{nextErr: iterationErr}, nil)
	rows, err := db.EnhancedQuery("SELECT value")
	if err != nil {
		t.Fatal(err)
	}
	if rows.Next() || !errors.Is(rows.Err(), iterationErr) {
		t.Fatalf("Err() = %v, want iteration error", rows.Err())
	}
	if err = rows.Close(); err != nil {
		t.Fatalf("Close() after iteration failure = %v, want nil", err)
	}
}

func TestEnhanceRow_scan_error_does_not_affect_Err(t *testing.T) {
	fixture := &scanTestRows{values: []driver.Value{"bad"}}
	row := newScanTestDb(t, fixture, nil).EnhancedQueryRow("SELECT value")
	var number int
	if err := row.Scan(&number); err == nil {
		t.Fatal("expected conversion error")
	}
	if err := row.Err(); err != nil {
		t.Fatalf("Err() after conversion failure = %v, want nil", err)
	}
	var text string
	if err := row.Scan(&text); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("second Scan() error = %v, want ErrNoRows", err)
	}
	if fixture.closeCalls != 1 {
		t.Fatalf("driver Close() calls = %d, want 1", fixture.closeCalls)
	}
}

func TestEnhanceRow_SliceScan_error_does_not_affect_Err(t *testing.T) {
	row := newScanTestDb(t, &scanTestRows{values: []driver.Value{"bad"}}, nil).EnhancedQueryRow("SELECT value")
	if _, err := row.SliceScan(); err == nil {
		t.Fatal("expected conversion error")
	}
	if err := row.Err(); err != nil {
		t.Fatalf("Err() after conversion failure = %v, want nil", err)
	}
}

func TestEnhanceRow_MapScan_error_does_not_affect_Err(t *testing.T) {
	row := newScanTestDb(t, &scanTestRows{values: []driver.Value{"bad"}}, nil).EnhancedQueryRow("SELECT value")
	if _, err := row.MapScan(); err == nil {
		t.Fatal("expected conversion error")
	}
	if err := row.Err(); err != nil {
		t.Fatalf("Err() after conversion failure = %v, want nil", err)
	}
}

func TestEnhanceRow_RawBytes_is_rejected_and_closed(t *testing.T) {
	fixture := &scanTestRows{values: []driver.Value{"123"}}
	row := newScanTestDb(t, fixture, nil).EnhancedQueryRow("SELECT value")
	var raw sql.RawBytes
	if err := row.Scan(&raw); err == nil || !strings.Contains(err.Error(), "RawBytes isn't allowed") {
		t.Fatalf("Scan() error = %v, want RawBytes rejection", err)
	}
	if err := row.Err(); err != nil {
		t.Fatalf("Err() after RawBytes rejection = %v, want nil", err)
	}
	if fixture.closeCalls != 1 {
		t.Fatalf("driver Close() calls = %d, want 1", fixture.closeCalls)
	}
}

func TestEnhanceRow_iteration_error_does_not_affect_Err(t *testing.T) {
	iterationErr := errors.New("iteration failed")
	row := newScanTestDb(t, &scanTestRows{nextErr: iterationErr}, nil).EnhancedQueryRow("SELECT value")
	var number int
	if err := row.Scan(&number); !errors.Is(err, iterationErr) {
		t.Fatalf("Scan() error = %v, want iteration error", err)
	}
	if err := row.Err(); err != nil {
		t.Fatalf("Err() after iteration failure = %v, want nil", err)
	}
}

func TestEnhanceRow_query_error_is_preserved(t *testing.T) {
	queryErr := errors.New("query failed")
	row := newScanTestDb(t, nil, queryErr).EnhancedQueryRow("SELECT value")
	if row.Err() != queryErr {
		t.Fatalf("Err() = %v, want query error", row.Err())
	}
	if err := row.Scan(new(int)); err != queryErr {
		t.Fatalf("Scan() error = %v, want query error", err)
	}
	if _, err := row.SliceScan(); err != queryErr {
		t.Fatalf("SliceScan() error = %v, want query error", err)
	}
	if _, err := row.MapScan(); err != queryErr {
		t.Fatalf("MapScan() error = %v, want query error", err)
	}
}

func TestEnhanceRow_close_error_does_not_affect_Err(t *testing.T) {
	closeErr := errors.New("close failed")
	fixture := &scanTestRows{values: []driver.Value{int64(123)}, closeErr: closeErr}
	row := newScanTestDb(t, fixture, nil).EnhancedQueryRow("SELECT value")
	var number int
	if err := row.Scan(&number); !errors.Is(err, closeErr) || number != 123 {
		t.Fatalf("Scan() = %d, %v, want 123, close error", number, err)
	}
	if err := row.Err(); err != nil {
		t.Fatalf("Err() after close failure = %v, want nil", err)
	}
	if fixture.closeCalls != 1 {
		t.Fatalf("driver Close() calls = %d, want 1", fixture.closeCalls)
	}
}
