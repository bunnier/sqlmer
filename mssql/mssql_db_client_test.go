package mssql_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/testenv"
	"github.com/bunnier/sqlmer/mssql"
)

func init() {
	testenv.TryInitConfig("..")
}

func getMsSqlClientOrSkip(t *testing.T) sqlmer.DbClient {
	t.Helper()

	dbClient, err := testenv.NewSqlServerClient()
	if err != nil {
		t.Skipf("skip mssql integration smoke test: %v", err)
	}
	if _, _, err = dbClient.Scalar("SELECT 1"); err != nil {
		t.Skipf("skip mssql integration smoke test: %v", err)
	}
	return dbClient
}

func Test_NewMsSqlDbClient(t *testing.T) {
	dbClient := getMsSqlClientOrSkip(t)

	if dbClient.Dsn() != testenv.TestConf.SqlServer {
		t.Errorf("mssqlDbClient.Dsn() connString = %v, wantConnString %v", dbClient.Dsn(), testenv.TestConf.SqlServer)
	}

	if _, err := mssql.NewMsSqlDbClient("test",
		sqlmer.WithConnTimeout(time.Second*15),
		sqlmer.WithExecTimeout(time.Second*15),
		sqlmer.WithPingCheck(true)); err == nil {
		t.Errorf("mssqlDbClient.NewMsSqlDbClient() err = nil, want has a err")
	}
}

func Test_MsSqlDbClient_internalDbClient_scalar_smoke(t *testing.T) {
	mssqlClient := getMsSqlClientOrSkip(t)

	got, hit, err := mssqlClient.Scalar("SELECT Id FROM go_TypeTest WHERE Id=@p1", 1)
	if err != nil {
		t.Fatalf("Scalar() error = %v", err)
	}
	if !hit {
		t.Fatal("Scalar() hit = false, want true")
	}
	if !reflect.DeepEqual(got, int64(1)) {
		t.Fatalf("Scalar() = %v, want %v", got, int64(1))
	}
}

func Test_MsSqlDbClient_internalDbClient_get_smoke(t *testing.T) {
	mssqlClient := getMsSqlClientOrSkip(t)

	got, err := mssqlClient.Get("SELECT DecimalTest, BitTest, DateTime2Test FROM go_TypeTest WHERE Id=@p1", 1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !reflect.DeepEqual(got["DecimalTest"], "1.4567899900") {
		t.Fatalf("Get()[DecimalTest] = %v, want %v", got["DecimalTest"], "1.4567899900")
	}
	if !reflect.DeepEqual(got["BitTest"], true) {
		t.Fatalf("Get()[BitTest] = %v, want %v", got["BitTest"], true)
	}

	gotTime, ok := got["DateTime2Test"].(time.Time)
	if !ok {
		t.Fatalf("Get()[DateTime2Test] type = %T, want time.Time", got["DateTime2Test"])
	}
	if gotTime.UnixNano() != time.Date(2021, 7, 1, 15, 38, 50, 425781300, time.UTC).UnixNano() {
		t.Fatalf("Get()[DateTime2Test] = %v", gotTime)
	}
}

func Test_MsSqlDbClient_internalDbClient_execute_smoke(t *testing.T) {
	mssqlClient := getMsSqlClientOrSkip(t)

	sqlText := `INSERT INTO go_TypeTest (TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest, BinaryTest)
	VALUES (5, 5, 5, 5, N'行5', 'Row5', N'行5', 'Row5', '2021-07-05 15:38:39.583', '2021-07-05 15:38:50.4257813', '2021-07-05', '12:05:01.345', 5.123, 5.12345, 5.45678999, 1);`

	effectedRows, err := mssqlClient.Execute(sqlText)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if effectedRows != 1 {
		t.Fatalf("Execute() effectedRows = %v, want %v", effectedRows, int64(1))
	}

	err = mssqlClient.SizedExecute(2, sqlText)
	if !errors.Is(err, sqlmer.ErrExpectedSizeWrong) {
		t.Fatalf("SizedExecute() error = %v, want ErrExpectedSizeWrong", err)
	}

	_, err = mssqlClient.Execute("DELETE FROM go_TypeTest WHERE TinyIntTest=5")
	if err != nil {
		t.Fatalf("cleanup Execute() error = %v", err)
	}
}
