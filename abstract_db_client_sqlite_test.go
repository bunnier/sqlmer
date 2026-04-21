package sqlmer_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/testenv"
	"github.com/bunnier/sqlmer/sqlite"
)

func newSqliteDbClientForAbstractDbTest(t *testing.T) sqlmer.DbClient {
	t.Helper()

	dsn := filepath.Join(t.TempDir(), "abstract.db")
	db, err := sql.Open(sqlite.DriverName, dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	if err = testenv.CreateSqliteSchema(db); err != nil {
		t.Fatalf("CreateSqliteSchema() error = %v", err)
	}

	dbClient, err := sqlite.NewSqliteDbClient(
		dsn,
		sqlmer.WithConnTimeout(testenv.DefaultTimeout),
		sqlmer.WithExecTimeout(testenv.DefaultTimeout),
	)
	if err != nil {
		t.Fatalf("NewSqliteDbClient() error = %v", err)
	}
	return dbClient
}

func assertApproxFloat64(t *testing.T, got float64, want float64) {
	t.Helper()

	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff > 0.00001 {
		t.Fatalf("float mismatch: got %v, want %v", got, want)
	}
}

func Test_AbstractDbClient_Execute(t *testing.T) {
	dbClient := newSqliteDbClientForAbstractDbTest(t)

	sqlText := `INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest)
	VALUES (5, 5, 5, 5, 5, '行5', '行5char', '行5text','2021-07-05','2021-07-05 15:38:50.425','2021-07-05 15:38:50.425', 5.456, 5.15678, 5.45678999, 1);`

	effectedRows, err := dbClient.Execute(sqlText)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if effectedRows != 1 {
		t.Fatalf("Execute() effectedRows = %v, want %v", effectedRows, int64(1))
	}

	err = dbClient.SizedExecute(1, sqlText)
	if err != nil {
		t.Fatalf("SizedExecute() error = %v", err)
	}

	err = dbClient.SizedExecute(2, sqlText)
	if !errors.Is(err, sqlmer.ErrExpectedSizeWrong) {
		t.Fatalf("SizedExecute() error = %v, want ErrExpectedSizeWrong", err)
	}
}

func Test_AbstractDbClient_Exists(t *testing.T) {
	dbClient := newSqliteDbClientForAbstractDbTest(t)

	got, err := dbClient.Exists("SELECT varcharTest FROM go_TypeTest WHERE id=1")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !got {
		t.Fatalf("Exists() = %v, want %v", got, true)
	}

	got, err = dbClient.Exists("SELECT varcharTest FROM go_TypeTest WHERE id=10000")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if got {
		t.Fatalf("Exists() = %v, want %v", got, false)
	}
}

func Test_AbstractDbClient_Scalar(t *testing.T) {
	dbClient := newSqliteDbClientForAbstractDbTest(t)

	got, hit, err := dbClient.Scalar("SELECT COUNT(1) FROM go_TypeTest WHERE id IN (@p1)", []int{1, 2, 3})
	if err != nil {
		t.Fatalf("Scalar() error = %v", err)
	}
	if !hit {
		t.Fatal("Scalar() hit = false, want true")
	}
	if !reflect.DeepEqual(got, int64(3)) {
		t.Fatalf("Scalar() = %v, want %v", got, int64(3))
	}

	got, hit, err = dbClient.Scalar(
		"SELECT COUNT(1) FROM go_TypeTest WHERE varcharTest=@varcharTest OR id=@p1 OR id=@p2",
		3,
		map[string]any{"varcharTest": "行1"},
		2,
	)
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

func Test_AbstractDbClient_Get(t *testing.T) {
	dbClient := newSqliteDbClientForAbstractDbTest(t)

	got, err := dbClient.Get(`SELECT varcharTest, dateTest, dateTimeTest, decimalTest, nullVarcharTest
		FROM go_TypeTest WHERE id=1`)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !reflect.DeepEqual(got["varcharTest"], "行1") {
		t.Fatalf("Get()[varcharTest] = %v, want %v", got["varcharTest"], "行1")
	}
	if !reflect.DeepEqual(got["dateTest"], time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("Get()[dateTest] = %v", got["dateTest"])
	}

	gotDateTime, ok := got["dateTimeTest"].(time.Time)
	if !ok {
		t.Fatalf("Get()[dateTimeTest] type = %T, want time.Time", got["dateTimeTest"])
	}
	if gotDateTime.UnixNano() != time.Date(2021, 7, 1, 15, 38, 50, 425000000, time.UTC).UnixNano() {
		t.Fatalf("Get()[dateTimeTest] = %v", gotDateTime)
	}

	gotDecimal, ok := got["decimalTest"].(float64)
	if !ok {
		t.Fatalf("Get()[decimalTest] type = %T, want float64", got["decimalTest"])
	}
	assertApproxFloat64(t, gotDecimal, 1.45678999)

	if got["nullVarcharTest"] != nil {
		t.Fatalf("Get()[nullVarcharTest] = %v, want nil", got["nullVarcharTest"])
	}

	emptyRow, err := dbClient.Get("SELECT varcharTest FROM go_TypeTest WHERE id=10000")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if emptyRow != nil {
		t.Fatalf("Get() = %v, want nil", emptyRow)
	}
}

func Test_AbstractDbClient_SliceGet(t *testing.T) {
	dbClient := newSqliteDbClientForAbstractDbTest(t)

	type idsArg struct {
		Ids []int
	}

	got, err := dbClient.SliceGet(
		"SELECT id, varcharTest FROM go_TypeTest WHERE id IN (@Ids) OR id IN (@p1, @p2) ORDER BY id",
		map[string]any{"Ids": []int{1}},
		2,
		4,
		idsArg{Ids: []int{3}},
	)
	if err != nil {
		t.Fatalf("SliceGet() error = %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("SliceGet() len = %v, want %v", len(got), 3)
	}
	if !reflect.DeepEqual(got[0]["id"], int64(2)) {
		t.Fatalf("SliceGet()[0][id] = %v, want %v", got[0]["id"], int64(2))
	}
	if !reflect.DeepEqual(got[1]["id"], int64(3)) {
		t.Fatalf("SliceGet()[1][id] = %v, want %v", got[1]["id"], int64(3))
	}
	if !reflect.DeepEqual(got[2]["id"], int64(4)) {
		t.Fatalf("SliceGet()[2][id] = %v, want %v", got[2]["id"], int64(4))
	}
}

func Test_AbstractDbClient_Rows(t *testing.T) {
	dbClient := newSqliteDbClientForAbstractDbTest(t)

	rows, err := dbClient.Rows(
		"SELECT varcharTest, dateTest, decimalTest FROM go_TypeTest WHERE id IN (@Ids) AND id!=@p1 ORDER BY id",
		struct{ Ids []int }{Ids: []int{1, 2}},
		1,
	)
	if err != nil {
		t.Fatalf("Rows() error = %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("Rows().Next() = false, want true")
	}

	got, err := rows.MapScan()
	if err != nil {
		t.Fatalf("Rows().MapScan() error = %v", err)
	}
	if !reflect.DeepEqual(got["varcharTest"], "行2") {
		t.Fatalf("Rows().MapScan()[varcharTest] = %v, want %v", got["varcharTest"], "行2")
	}
	if !reflect.DeepEqual(got["dateTest"], time.Date(2021, 7, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("Rows().MapScan()[dateTest] = %v", got["dateTest"])
	}

	gotDecimal, ok := got["decimalTest"].(float64)
	if !ok {
		t.Fatalf("Rows().MapScan()[decimalTest] type = %T, want float64", got["decimalTest"])
	}
	assertApproxFloat64(t, gotDecimal, 2.45678999)

	if rows.Next() {
		t.Fatal("Rows().Next() = true, want false")
	}
}

func Test_AbstractDbClient_Rows_error_wrapped(t *testing.T) {
	dbClient := newSqliteDbClientForAbstractDbTest(t)

	rows, err := dbClient.Rows("SELECT varcharTest FROM go_TypeTest WHERE id=@p1", 1)
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
	if !strings.Contains(err.Error(), "input sql=SELECT varcharTest FROM go_TypeTest WHERE id=@p1") {
		t.Fatalf("Rows().Scan() error = %v, want input sql", err)
	}
	if !strings.Contains(err.Error(), "@p1=1") {
		t.Fatalf("Rows().Scan() error = %v, want params", err)
	}
	if !errors.Is(rows.Err(), sqlmer.ErrExecutingSql) {
		t.Fatalf("Rows().Err() = %v, want ErrExecutingSql", rows.Err())
	}
}

func Test_AbstractDbClient_Row_error_wrapped(t *testing.T) {
	dbClient := newSqliteDbClientForAbstractDbTest(t)

	row, err := dbClient.Row("SELECT varcharTest FROM go_TypeTest WHERE id=@p1", 1)
	if err != nil {
		t.Fatalf("Row() error = %v", err)
	}

	var got int
	err = row.Scan(&got)
	if err == nil {
		t.Fatal("Row().Scan() error = nil, want wrapped error")
	}
	if !errors.Is(err, sqlmer.ErrExecutingSql) {
		t.Fatalf("Row().Scan() error = %v, want ErrExecutingSql", err)
	}
	if !strings.Contains(err.Error(), "input sql=SELECT varcharTest FROM go_TypeTest WHERE id=@p1") {
		t.Fatalf("Row().Scan() error = %v, want input sql", err)
	}
	if !strings.Contains(err.Error(), "@p1=1") {
		t.Fatalf("Row().Scan() error = %v, want params", err)
	}
	if !errors.Is(row.Err(), sqlmer.ErrExecutingSql) {
		t.Fatalf("Row().Err() = %v, want ErrExecutingSql", row.Err())
	}
}

func Test_AbstractDbClient_Row_no_rows_not_wrapped(t *testing.T) {
	dbClient := newSqliteDbClientForAbstractDbTest(t)

	row, err := dbClient.Row("SELECT varcharTest FROM go_TypeTest WHERE id=@p1", 10000)
	if err != nil {
		t.Fatalf("Row() error = %v", err)
	}

	var got string
	err = row.Scan(&got)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Row().Scan() error = %v, want sql.ErrNoRows", err)
	}
	if errors.Is(err, sqlmer.ErrExecutingSql) {
		t.Fatalf("Row().Scan() error = %v, do not want ErrExecutingSql", err)
	}
}
