// DbEnhance 主要是数据库无关的封装，目前测试用例只用 Sql Server 做。
package sqlen_test

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/bunnier/sqlmer/internal/testenv"
	"github.com/bunnier/sqlmer/sqlen"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	testenv.TryInitConfig("..")
}

func mustGetMssqlDb(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", testenv.TestConf.Mysql)
	if err != nil {
		t.Errorf("sql.Open() error = %v, wantErr nil", err)
	}

	if err = db.Ping(); err != nil {
		t.Errorf("db.Ping() error = %v, wantErr nil", err)
	}

	return db
}

func unifyDataTypeFn(columnType *sql.ColumnType, dest *any) {
	switch columnType.DatabaseTypeName() {
	case "VARCHAR", "CHAR", "TEXT", "DECIMAL":
		switch v := (*dest).(type) {
		case sql.RawBytes:
			if v == nil {
				*dest = nil
				break
			}
			*dest = string(v)
		case nil:
			*dest = nil
		}
	}
}

func getScanTypeFunc(columnType *sql.ColumnType) reflect.Type {
	return columnType.ScanType()
}

func TestEnhanceRows_MapScan(t *testing.T) {
	db := sqlen.NewDbEnhance(mustGetMssqlDb(t), getScanTypeFunc, unifyDataTypeFn)

	const testNum int64 = 3
	enhancedRows, err := db.EnhancedQuery("SELECT Id, VarcharTest FROM go_TypeTest WHERE Id<=?", testNum)
	if err != nil {
		t.Fatal(err)
	}
	defer enhancedRows.Close()

	count := int64(0)
	for enhancedRows.Next() {
		count++

		var rowMap map[string]any
		var err error

		if rowMap, err = enhancedRows.MapScan(); err != nil {
			t.Errorf("enhancedRows.MapScan() error = %v, wantErr nil", err)
			return
		}
		if id, ok := rowMap["Id"]; !ok || id.(int64) != int64(count) {
			t.Errorf("enhancedRows.MapScan() Id = %d, wantId %d", id, count)
		}
		if str, ok := rowMap["VarcharTest"]; !ok || str.(string) != fmt.Sprintf("行%d", count) {
			t.Errorf("enhancedRows.MapScan() VarcharTest = %v, wantVarcharTest %v", str.(string), fmt.Sprintf("Row%d", count))
		}
	}
	enhancedRows.Close()

	if enhancedRows.Err() != nil {
		t.Errorf("enhancedRows.Err() error = %v, wantErr nil", err)
		return
	}
	if count != testNum {
		t.Fatalf("enhancedRows.MapScan() total of rows is error, want 3 but %d", count)
	}
}

func TestEnhanceRows_SliceScan(t *testing.T) {
	db := sqlen.NewDbEnhance(mustGetMssqlDb(t), getScanTypeFunc, unifyDataTypeFn)

	const testNum int64 = 3
	enhancedRows, err := db.EnhancedQuery("SELECT Id, VarcharTest, DecimalTest FROM go_TypeTest WHERE Id<=?", testNum)
	if err != nil {
		t.Fatal(err)
	}
	defer enhancedRows.Close()

	count := int64(0)
	for enhancedRows.Next() {
		count++
		sliceRow, err := enhancedRows.SliceScan()
		if err != nil {
			t.Errorf("enhancedRows.SliceScan() error = %v, wantErr nil", err)
			return
		}
		if id, ok := sliceRow[0].(int64); !ok || id != int64(count) {
			t.Errorf("enhancedRows.SliceScan() Id = %d, wantId %d", id, count)
		}

		if str, ok := sliceRow[1].(string); !ok || str != fmt.Sprintf("行%d", count) {
			t.Errorf("enhancedRows.SliceScan() VarcharTest = %v, wantVarcharTest %v", str, fmt.Sprintf("Row%d", count))
		}
	}
	enhancedRows.Close()

	if enhancedRows.Err() != nil {
		t.Errorf("enhancedRows.Err() error = %v, wantErr nil", err)
		return
	}
	if count != testNum {
		t.Fatalf("enhancedRows.SliceScan() total of rows is error, want 3 but %d", count)
	}
}

func TestEnhanceRow_MapScan(t *testing.T) {
	db := sqlen.NewDbEnhance(mustGetMssqlDb(t), getScanTypeFunc, unifyDataTypeFn)
	enhancedRow := db.EnhancedQueryRow("SELECT Id, VarcharTest, DecimalTest FROM go_TypeTest WHERE Id=1")

	var rowMap map[string]any
	var err error
	if rowMap, err = enhancedRow.MapScan(); err != nil {
		t.Errorf("enhancedRow.MapScan() error = %v, wantErr nil", err)
		return
	}
	if id, ok := rowMap["Id"]; !ok || id.(int64) != 1 {
		t.Errorf("enhancedRow.MapScan() Id = %d, wantId 1", id)
	}
	if str, ok := rowMap["VarcharTest"]; !ok || str.(string) != "行1" {
		t.Errorf("enhancedRow.MapScan() VarcharTest = %v, wantVarcharTest Row1", str)
	}
}

func TestEnhanceRow_SliceScan(t *testing.T) {
	db := sqlen.NewDbEnhance(mustGetMssqlDb(t), getScanTypeFunc, unifyDataTypeFn)
	enhancedRow := db.EnhancedQueryRow("SELECT Id, VarcharTest FROM go_TypeTest WHERE Id=1")

	sliceRow, err := enhancedRow.SliceScan()
	if err != nil {
		t.Errorf("enhancedRow.SliceScan() error = %v, wantErr nil", err)
		return
	}

	if id, ok := sliceRow[0].(int64); !ok || id != 1 {
		t.Errorf("enhancedRow.SliceScan() Id = %d, wantId 1", id)
	}
	if str, ok := sliceRow[1].(string); !ok || str != "行1" {
		t.Errorf("enhancedRow.SliceScan() VarcharTest = %v, wantVarcharTest Row1", str)
	}
}

func TestEnhanceRow_Scan(t *testing.T) {
	db := sqlen.NewDbEnhance(mustGetMssqlDb(t), getScanTypeFunc, unifyDataTypeFn)
	enhancedRow := db.EnhancedQueryRow("SELECT Id, VarcharTest FROM go_TypeTest WHERE Id=1")

	var id int64
	var str string
	err := enhancedRow.Scan(&id, &str)
	if err != nil {
		t.Errorf("enhancedRow.Scan() error = %v, wantErr nil", err)
		return
	}

	if id != 1 {
		t.Errorf("enhancedRow.Scan() Id = %d, wantId 1", id)
	}
	if str != "行1" {
		t.Errorf("enhancedRow.Scan() VarcharTest = %v, wantVarcharTest Row1", str)
	}
}

func TestEnhanceRow_Err(t *testing.T) {
	db := sqlen.NewDbEnhance(mustGetMssqlDb(t), getScanTypeFunc, unifyDataTypeFn)
	sqlText := "SELECT Id, VarcharTest FROM go_TypeTest WHERE Id=10000" // 没数据。

	t.Run("SliceScan", func(t *testing.T) {
		enhancedRow := db.EnhancedQueryRow(sqlText)
		_, err := enhancedRow.SliceScan()
		if err != sql.ErrNoRows || err != enhancedRow.Err() || err != sql.ErrNoRows {
			t.Errorf("enhancedRow.EnhancedQueryRow() error = %v, wantErr ErrNoRows", err)
		}
	})

	t.Run("EmptySliceScan", func(t *testing.T) {
		enhancedRow := &sqlen.EnhanceRow{}
		_, err := enhancedRow.SliceScan()
		if err != sql.ErrNoRows || err != enhancedRow.Err() || err != sql.ErrNoRows {
			t.Errorf("enhancedRow.SliceScan() error = %v, wantErr ErrNoRows", err)
		}
	})

	t.Run("MapScan", func(t *testing.T) {
		enhancedRow := db.EnhancedQueryRow(sqlText)
		_, err := enhancedRow.MapScan()
		if err != sql.ErrNoRows || err != enhancedRow.Err() || err != sql.ErrNoRows {
			t.Errorf("enhancedRow.EnhancedQueryRow() error = %v, wantErr ErrNoRows", err)
		}
	})

	t.Run("EmptyMapScan", func(t *testing.T) {
		enhancedRow := &sqlen.EnhanceRow{}
		_, err := enhancedRow.MapScan()
		if err != sql.ErrNoRows || err != enhancedRow.Err() || err != sql.ErrNoRows {
			t.Errorf("enhancedRow.MapScan() error = %v, wantErr ErrNoRows", err)
		}
	})

	t.Run("Scan", func(t *testing.T) {
		enhancedRow := db.EnhancedQueryRow(sqlText)
		err := enhancedRow.Scan()
		if err != sql.ErrNoRows || err != enhancedRow.Err() || err != sql.ErrNoRows {
			t.Errorf("enhancedRow.Scan() error = %v, wantErr ErrNoRows", err)
		}
	})

	t.Run("EmptyRowScan", func(t *testing.T) {
		enhancedRow := &sqlen.EnhanceRow{}
		err := enhancedRow.Scan()
		if err != sql.ErrNoRows || err != enhancedRow.Err() || err != sql.ErrNoRows {
			t.Errorf("enhancedRow.Scan() error = %v, wantErr ErrNoRows", err)
		}
	})
}
