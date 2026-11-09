package sqlen_test

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/bunnier/sqlmer/sqlen"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func sqliteTestGetScanTypeFunc(columnType *sql.ColumnType) reflect.Type {
	tp := columnType.ScanType()
	if tp == nil {
		return reflect.TypeOf(new(any)).Elem()
	}
	return tp
}

func sqliteTestUnifyDataTypeFn(columnType *sql.ColumnType, dest *any) {
	switch v := (*dest).(type) {
	case []byte:
		*dest = string(v)
	case sql.RawBytes:
		if v == nil {
			*dest = nil
			return
		}
		*dest = string(v)
	}
}

func newSqliteEnhanceForTest(t *testing.T) (*sqlen.DbEnhance, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	_, err = db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE go_TypeTest (
		Id INTEGER PRIMARY KEY,
		VarcharTest TEXT NOT NULL,
		DecimalTest REAL NOT NULL
	)`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO t(id, name) VALUES (1, 'row1')`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO go_TypeTest(Id, VarcharTest, DecimalTest) VALUES
		(1, '行1', 1.11),
		(2, '行2', 2.22),
		(3, '行3', 3.33)`)
	if err != nil {
		t.Fatal(err)
	}

	return sqlen.NewDbEnhance(db, sqliteTestGetScanTypeFunc, sqliteTestUnifyDataTypeFn), db
}

func TestEnhanceRows_SliceScan_returns_init_columns_error(t *testing.T) {
	dbEnhance, _ := newSqliteEnhanceForTest(t)

	rows, err := dbEnhance.EnhancedQuery(`SELECT id, name FROM t`)
	if err != nil {
		t.Fatal(err)
	}

	if err = rows.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = rows.SliceScan()
	if err == nil {
		t.Fatal("expected initColumns error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "closed") {
		t.Fatalf("expected closed rows error, got %v", err)
	}
	if err = rows.Err(); err != nil {
		t.Fatalf("metadata error affected Err(): %v", err)
	}
}

func TestEnhanceRows_Err_returns_cached_wrapped_error(t *testing.T) {
	dbEnhance, _ := newSqliteEnhanceForTest(t)

	rows, err := dbEnhance.EnhancedQuery(`SELECT 123 UNION ALL SELECT abs(-9223372036854775808)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("expected first row, got none")
	}

	wrappedErr := errors.New("wrapped iteration error")
	wrapCount := 0
	rows.SetErrWrapper(func(err error) error {
		wrapCount++
		return wrappedErr
	})

	var got int
	if err = rows.Scan(&got); err != nil {
		t.Fatal(err)
	}
	if rows.Next() {
		t.Fatal("expected iteration error")
	}

	err = rows.Err()
	if !errors.Is(err, wrappedErr) {
		t.Fatalf("expected wrapped iteration error, got %v", err)
	}
	if rows.Err() != err || wrapCount != 1 {
		t.Fatalf("expected Err() to reuse cached error, got wrapper count %d", wrapCount)
	}
	if err = rows.Close(); err != nil {
		t.Fatalf("Close() after iteration error = %v, want nil", err)
	}
}

type scanRows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close() error
}

func TestEnhanceRows_scan_errors_match_standard_library(t *testing.T) {
	db, rawDb := newSqliteEnhanceForTest(t)
	queries := []string{
		`SELECT 'bad' UNION ALL SELECT '123' UNION ALL SELECT '456'`,
		`SELECT '123' UNION ALL SELECT 'bad' UNION ALL SELECT '456'`,
		`SELECT '123' UNION ALL SELECT '456' UNION ALL SELECT 'bad'`,
	}
	for _, query := range queries {
		for _, enhanced := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/enhanced=%t", query, enhanced), func(t *testing.T) {
				var rows scanRows
				var err error
				if enhanced {
					rows, err = db.EnhancedQuery(query)
				} else {
					rows, err = rawDb.Query(query)
				}
				if err != nil {
					t.Fatal(err)
				}
				defer rows.Close()
				var numbers []int
				failures := 0
				for rows.Next() {
					var number int
					if err = rows.Scan(&number); err != nil {
						failures++
					} else {
						numbers = append(numbers, number)
					}
					if err = rows.Err(); err != nil {
						t.Fatalf("Err() reported a scan failure: %v", err)
					}
				}
				if failures != 1 || !reflect.DeepEqual(numbers, []int{123, 456}) {
					t.Fatalf("got values %v, failures %d; want [123 456], 1", numbers, failures)
				}
				if err = rows.Err(); err != nil {
					t.Fatalf("Err() after iteration = %v, want nil", err)
				}
				for range 2 {
					if err = rows.Close(); err != nil {
						t.Fatalf("Close() = %v, want nil", err)
					}
				}
			})
		}
	}
}

func TestEnhanceRows_close_after_scan_error_matches_standard_library(t *testing.T) {
	db, rawDb := newSqliteEnhanceForTest(t)
	for _, enhanced := range []bool{false, true} {
		t.Run(fmt.Sprintf("enhanced=%t", enhanced), func(t *testing.T) {
			var rows scanRows
			var err error
			if enhanced {
				rows, err = db.EnhancedQuery(`SELECT 'bad'`)
			} else {
				rows, err = rawDb.Query(`SELECT 'bad'`)
			}
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			if !rows.Next() {
				t.Fatal("expected first row")
			}
			var number int
			if err = rows.Scan(&number); err == nil {
				t.Fatal("expected conversion error")
			}
			for range 2 {
				if err = rows.Close(); err != nil {
					t.Fatalf("Close() after scan error = %v, want nil", err)
				}
			}
			if err = rows.Err(); err != nil {
				t.Fatalf("Err() after Close() = %v, want nil", err)
			}
		})
	}
}

func TestEnhanceRows_scan_retry_matches_standard_library(t *testing.T) {
	db, rawDb := newSqliteEnhanceForTest(t)
	for _, enhanced := range []bool{false, true} {
		t.Run(fmt.Sprintf("enhanced=%t", enhanced), func(t *testing.T) {
			var rows scanRows
			var err error
			if enhanced {
				rows, err = db.EnhancedQuery(`SELECT 'bad'`)
			} else {
				rows, err = rawDb.Query(`SELECT 'bad'`)
			}
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			if !rows.Next() {
				t.Fatal("expected first row")
			}
			var number int
			if err = rows.Scan(&number); err == nil {
				t.Fatal("expected conversion error")
			}
			var text string
			if err = rows.Scan(&text); err != nil || text != "bad" {
				t.Fatalf("second Scan() = %q, %v, want bad, nil", text, err)
			}
			if err = rows.Err(); err != nil {
				t.Fatalf("Err() after second Scan() = %v, want nil", err)
			}
		})
	}
}

func TestEnhanceRows_scan_extensions_continue_after_conversion_error(t *testing.T) {
	_, rawDb := newSqliteEnhanceForTest(t)
	db := sqlen.NewDbEnhance(rawDb, func(*sql.ColumnType) reflect.Type {
		return reflect.TypeOf(int64(0))
	}, func(*sql.ColumnType, *any) {})
	for _, method := range []string{"SliceScan", "MapScan"} {
		t.Run(method, func(t *testing.T) {
			rows, err := db.EnhancedQuery(`SELECT 'bad' AS value UNION ALL SELECT '123' AS value`)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			scan := func() (any, error) {
				if method == "SliceScan" {
					return rows.SliceScan()
				}
				return rows.MapScan()
			}
			if !rows.Next() {
				t.Fatal("expected first row")
			}
			if _, err = scan(); err == nil {
				t.Fatal("expected conversion error")
			}
			if err = rows.Err(); err != nil {
				t.Fatalf("Err() after scan failure = %v, want nil", err)
			}
			if !rows.Next() {
				t.Fatal("expected second row")
			}
			value, err := scan()
			if err != nil {
				t.Fatal(err)
			}
			var want any = []any{int64(123)}
			if method == "MapScan" {
				want = map[string]any{"value": int64(123)}
			}
			if !reflect.DeepEqual(value, want) {
				t.Fatalf("got %v, want %v", value, want)
			}

			row := db.EnhancedQueryRow(`SELECT 'bad' AS value`)
			if method == "SliceScan" {
				_, err = row.SliceScan()
			} else {
				_, err = row.MapScan()
			}
			if err == nil {
				t.Fatal("expected single-row conversion error")
			}
			if err = row.Err(); err != nil {
				t.Fatalf("Row.Err() after conversion failure = %v, want nil", err)
			}
		})
	}
}

func TestEnhanceRow_errors_match_standard_library(t *testing.T) {
	db, rawDb := newSqliteEnhanceForTest(t)
	for _, query := range []string{`SELECT name FROM t`, `SELECT name FROM t WHERE id=0`, `SELECT missing FROM t`, `SELECT 123`} {
		for _, enhanced := range []bool{false, true} {
			for _, rawBytes := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/enhanced=%t/rawBytes=%t", query, enhanced, rawBytes), func(t *testing.T) {
					var row interface {
						Scan(...any) error
						Err() error
					}
					if enhanced {
						row = db.EnhancedQueryRow(query)
					} else {
						row = rawDb.QueryRow(query)
					}
					queryErr := row.Err()
					if (queryErr != nil) != (query == `SELECT missing FROM t`) {
						t.Fatalf("unexpected query error: %v", queryErr)
					}
					var dest any = new(int)
					if rawBytes {
						dest = new(sql.RawBytes)
					}
					err := row.Scan(dest)
					switch {
					case queryErr != nil:
						if err != queryErr {
							t.Fatalf("Scan() = %v, want query error %v", err, queryErr)
						}
					case rawBytes:
						if err == nil || !strings.Contains(err.Error(), "RawBytes isn't allowed") {
							t.Fatalf("Scan() = %v, want RawBytes error", err)
						}
					case query == `SELECT name FROM t WHERE id=0`:
						if !errors.Is(err, sql.ErrNoRows) {
							t.Fatalf("Scan() = %v, want ErrNoRows", err)
						}
					case query == `SELECT 123`:
						if err != nil {
							t.Fatal(err)
						}
					default:
						if err == nil {
							t.Fatal("expected conversion error")
						}
					}
					if row.Err() != queryErr {
						t.Fatalf("Row.Err() changed after Scan(): %v", row.Err())
					}
					var text string
					err = row.Scan(&text)
					if queryErr != nil {
						if err != queryErr {
							t.Fatalf("second Scan() = %v, want query error %v", err, queryErr)
						}
					} else if !errors.Is(err, sql.ErrNoRows) {
						t.Fatalf("second Scan() = %v, want ErrNoRows from closed row", err)
					}
				})
			}
		}
	}
}
