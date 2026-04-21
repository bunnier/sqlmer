package sqlite_test

import (
	"database/sql"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/testenv"
	"github.com/bunnier/sqlmer/sqlite"
)

func init() {
	testenv.TryInitConfig("..")
}

func TestMain(m *testing.M) {
	// 删除旧的数据库文件。
	os.Remove("test.db")

	// 创建数据库和 Schema。
	db, err := sql.Open("sqlite3", "test.db")
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

func Test_NewSqliteDbClient_with_sql_parser_cache_capacity(t *testing.T) {
	dbClient, err := sqlite.NewSqliteDbClient("test.db",
		sqlmer.WithConnTimeout(testenv.DefaultTimeout),
		sqlmer.WithExecTimeout(testenv.DefaultTimeout),
		sqlite.WithSqlParserCacheCapacity(8))
	if err != nil {
		t.Fatal(err)
	}

	got, _, err := dbClient.Scalar("SELECT id FROM go_TypeTest WHERE id=@p1", 1)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, int64(1)) {
		t.Errorf("dbClient.Scalar() = %v, want %v", got, int64(1))
	}
}

func Test_WithSqlParserCacheCapacity_invalid_capacity(t *testing.T) {
	_, err := sqlite.NewSqliteDbClient("test.db", sqlite.WithSqlParserCacheCapacity(0))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_SqliteDbClient_internalDbClient_scalar_smoke(t *testing.T) {
	sqliteClient, err := testenv.NewSqliteClient()
	if err != nil {
		t.Fatal(err)
	}

	got, hit, err := sqliteClient.Scalar("SELECT COUNT(1) FROM go_TypeTest WHERE id IN (@p1)", []int{1, 2, 3})
	if err != nil {
		t.Fatalf("Scalar() error = %v", err)
	}
	if !hit {
		t.Fatal("Scalar() hit = false, want true")
	}
	if !reflect.DeepEqual(got, int64(3)) {
		t.Fatalf("Scalar() = %v, want %v", got, int64(3))
	}
}

func Test_SqliteDbClient_internalDbClient_get_smoke(t *testing.T) {
	sqliteClient, err := testenv.NewSqliteClient()
	if err != nil {
		t.Fatal(err)
	}

	got, err := sqliteClient.Get("SELECT varcharTest, dateTest, decimalTest FROM go_TypeTest WHERE id=@p1", 1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !reflect.DeepEqual(got["varcharTest"], "行1") {
		t.Fatalf("Get()[varcharTest] = %v, want %v", got["varcharTest"], "行1")
	}
	if !reflect.DeepEqual(got["dateTest"], time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("Get()[dateTest] = %v", got["dateTest"])
	}

	gotDecimal, ok := got["decimalTest"].(float64)
	if !ok {
		t.Fatalf("Get()[decimalTest] type = %T, want float64", got["decimalTest"])
	}
	if gotDecimal < 1.45678 || gotDecimal > 1.45680 {
		t.Fatalf("Get()[decimalTest] = %v, want about %v", gotDecimal, 1.45678999)
	}
}

func Test_SqliteDbClient_internalDbClient_rows_error_wrapped_smoke(t *testing.T) {
	sqliteClient, err := testenv.NewSqliteClient()
	if err != nil {
		t.Fatal(err)
	}

	rows, err := sqliteClient.Rows("SELECT varcharTest FROM go_TypeTest WHERE id=@p1", 1)
	if err != nil {
		t.Fatalf("Rows() error = %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("Rows().Next() = false, want true")
	}

	var got int
	err = rows.Scan(&got)
	if err == nil {
		t.Fatal("Rows().Scan() error = nil, want wrapped error")
	}
	if !errors.Is(err, sqlmer.ErrExecutingSql) {
		t.Fatalf("Rows().Scan() error = %v, want ErrExecutingSql", err)
	}
	if !errors.Is(rows.Err(), sqlmer.ErrExecutingSql) {
		t.Fatalf("Rows().Err() = %v, want ErrExecutingSql", rows.Err())
	}
}
