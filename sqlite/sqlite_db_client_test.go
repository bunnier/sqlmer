package sqlite_test

import (
	"database/sql"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/testenv"
	"github.com/bunnier/sqlmer/sqlen"
	"github.com/bunnier/sqlmer/sqlite"
)

func init() {
	testenv.TryInitConfig("..")
}

func TestMain(m *testing.M) {
	// 删除旧的数据库文件。
	os.Remove("test.db")

	// 创建数据库和 Schema。
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		panic(err)
	}

	if err := testenv.CreateSqliteSchema(db); err != nil {
		panic(err)
	}
	db.Close()

	code := m.Run()

	// 清理。
	os.Remove("test.db")
	os.Exit(code)
}

func Test_NewSqliteDbClient(t *testing.T) {
	dbClient, err := testenv.NewSqliteClient()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(dbClient.Dsn(), testenv.TestConf.Sqlite) {
		t.Errorf("sqliteDbClient.Dsn() connString = %v, want contains  %v", dbClient.Dsn(), testenv.TestConf.Sqlite)
	}

	_, err = sqlite.NewSqliteDbClient("test.db",
		sqlmer.WithConnTimeout(testenv.DefaultTimeout),
		sqlmer.WithExecTimeout(testenv.DefaultTimeout),
		sqlmer.WithPingCheck(true))
	if err != nil {
		panic(err)
	}
}

func Test_internalDbClient_Scalar(t *testing.T) {
	sqliteClient, err := testenv.NewSqliteClient()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("sqlite1", func(t *testing.T) {
		got, _, err := sqliteClient.Scalar("SELECT id FROM go_TypeTest WHERE id=@p1", 1)
		if err != nil {
			t.Errorf("internalDbClient.Scalar() error = %v, wantErr %v", err, false)
			return
		}
		if !reflect.DeepEqual(got, int64(1)) {
			t.Errorf("internalDbClient.Scalar() = %v, want %v", got, int64(1))
		}
	})

	t.Run("sqlite2", func(t *testing.T) {
		got, _, err := sqliteClient.Scalar("SELECT id FROM go_TypeTest WHERE id in (@p1)", []int{1})
		if err != nil {
			t.Errorf("internalDbClient.Scalar() error = %v, wantErr %v", err, false)
			return
		}
		if !reflect.DeepEqual(got, int64(1)) {
			t.Errorf("internalDbClient.Scalar() = %v, want %v", got, int64(1))
		}
	})
}

func Test_internalDbClient_Get(t *testing.T) {
	sqliteClient, err := testenv.NewSqliteClient()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("sqlite_nullable_null", func(t *testing.T) {
		got, err := sqliteClient.Get(`SELECT intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest,
				nullIntTest, nullTinyintTest, nullSmallIntTest, nullBigIntTest, nullUnsignedTest, nullVarcharTest, nullCharTest, nullTextTest, nullDateTest, nullDateTimeTest, nullTimestampTest, nullFloatTest, nullDoubleTest, nullDecimalTest, nullBitTest,
				null colNull 
				FROM go_TypeTest WHERE id=1`)
		if err != nil {
			t.Errorf("internalDbClient.Get() error = %v, wantErr %v", err, false)
			return
		}

		want := map[string]any{
			"intTest":           int64(1),
			"tinyintTest":       int64(1),
			"smallIntTest":      int64(1),
			"bigIntTest":        int64(1),
			"unsignedTest":      int64(1),
			"varcharTest":       "行1",
			"charTest":          "行1char",
			"charTextTest":      "行1text",
			"dateTest":          time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
			"dateTimeTest":      time.Date(2021, 7, 1, 15, 38, 50, 425000000, time.UTC), // 精度可能会丢失或不同。
			"timestampTest":     time.Date(2021, 7, 1, 15, 38, 50, 425000000, time.UTC),
			"floatTest":         float64(1.456),
			"doubleTest":        float64(1.15678),
			"decimalTest":       1.45678999, // SQLite 可能会标准化或返回 float。
			"bitTest":           int64(1),
			"nullIntTest":       nil,
			"nullTinyintTest":   nil,
			"nullSmallIntTest":  nil,
			"nullBigIntTest":    nil,
			"nullUnsignedTest":  nil,
			"nullVarcharTest":   nil,
			"nullCharTest":      nil,
			"nullTextTest":      nil,
			"nullDateTest":      nil,
			"nullDateTimeTest":  nil,
			"nullTimestampTest": nil,
			"nullFloatTest":     nil,
			"nullDoubleTest":    nil,
			"nullDecimalTest":   nil,
			"nullBitTest":       nil,
			"colNull":           nil,
		}

		for k, v := range got {
			wantV := want[k]
			// 处理带有容差的时间比较或字符串格式问题。
			if k == "dateTimeTest" || k == "timestampTest" {
				// SQLite 存储为字符串，然后解析回来。精度可能会有所不同。
				// format 时候用的是 .999999999，所以它应该能在毫秒级别也准确。
				if gotTime, ok := v.(time.Time); ok {
					if wantT, ok2 := wantV.(time.Time); ok2 {
						if !gotTime.Equal(wantT) {
							// 尝试对时区宽松处理。
							if gotTime.UnixNano() != wantT.UnixNano() {
								t.Errorf("fieldname = %s, got %v, want %v", k, v, wantV)
							}
						}
						continue
					}
				}
			}

			if !reflect.DeepEqual(v, wantV) {
				if wantFloat, ok := wantV.(float64); ok {
					if gotFloat, ok2 := v.(float64); ok2 {
						if wantFloat-gotFloat < 0.00001 && gotFloat-wantFloat < 0.00001 {
							continue
						}
					}
				}
				t.Errorf("fieldname = %s, internalDbClient.Get() = %v, want %v", k, v, wantV)
			}
		}
	})
}

func Test_internalDbClient_Rows(t *testing.T) {
	row1 := map[string]any{
		"varcharTest": "行1",
		"dateTest":    time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
		"decimalTest": 1.45678999,
	}
	row2 := map[string]any{
		"varcharTest": "行2",
		"dateTest":    time.Date(2021, 7, 2, 0, 0, 0, 0, time.UTC),
		"decimalTest": 2.45678999,
	}

	sqliteClient, err := testenv.NewSqliteClient()
	if err != nil {
		t.Fatal(err)
	}

	rowAssert := func(rows *sqlen.EnhanceRows, want []map[string]any) {
		index := 0
		for rows.Next() {
			got, err := rows.MapScan()
			if err != nil {
				t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
				return
			}
			// 仅检查特定字段。
			for k, wantV := range want[index] {
				v := got[k]
				if !reflect.DeepEqual(v, wantV) {
					t.Errorf("fieldname = %s, got = %v, want %v", k, v, wantV)
					return
				}
			}
			index++
		}
	}

	t.Run("arrayParams", func(t *testing.T) {
		rows, err := sqliteClient.Rows("SELECT varcharTest,dateTest,decimalTest FROM go_TypeTest WHERE id IN (@p1)", []int{1, 2})
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row1, row2})
	})
}
