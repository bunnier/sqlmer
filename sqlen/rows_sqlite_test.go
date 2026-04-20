package sqlen_test

import (
	"database/sql"
	"errors"
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
}

func TestEnhanceRows_Err_returns_cached_wrapped_error(t *testing.T) {
	dbEnhance, _ := newSqliteEnhanceForTest(t)

	rows, err := dbEnhance.EnhancedQuery(`SELECT name FROM t`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("expected first row, got none")
	}

	wrappedErr := errors.New("wrapped scan error")
	wrapCount := 0
	rows.SetErrWrapper(func(err error) error {
		wrapCount++
		return wrappedErr
	})

	var got int
	err = rows.Scan(&got)
	if !errors.Is(err, wrappedErr) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
	if wrapCount != 1 {
		t.Fatalf("expected wrapper called once, got %d", wrapCount)
	}

	err = rows.Err()
	if !errors.Is(err, wrappedErr) {
		t.Fatalf("expected cached wrapped error from Err(), got %v", err)
	}
	if wrapCount != 1 {
		t.Fatalf("expected Err() to reuse cached error, got wrapper count %d", wrapCount)
	}
}
