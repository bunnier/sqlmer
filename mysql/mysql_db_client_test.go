package mysql_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/testenv"
	"github.com/bunnier/sqlmer/mysql"
)

func init() {
	testenv.TryInitConfig("..")
}

func Test_NewMySqlDbClient(t *testing.T) {
	dbClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(dbClient.Dsn(), testenv.TestConf.Mysql) {
		t.Errorf("mysqlDbClient.Dsn() connString = %v, want contains  %v", dbClient.Dsn(), testenv.TestConf.Mysql)
	}

	_, err = mysql.NewMySqlDbClient("test",
		sqlmer.WithConnTimeout(testenv.DefaultTimeout),
		sqlmer.WithExecTimeout(testenv.DefaultTimeout),
		sqlmer.WithPingCheck(true))
	if err == nil {
		t.Errorf("mysqlDbClient.NewMsSqlDbClient() err = nil, want has a err")
	}
}

func getMysqlClientOrSkip(t *testing.T) sqlmer.DbClient {
	t.Helper()

	mysqlClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Skipf("skip mysql integration smoke test: %v", err)
	}
	return mysqlClient
}

func Test_MySqlDbClient_internalDbClient_scalar_smoke(t *testing.T) {
	mysqlClient := getMysqlClientOrSkip(t)

	got, hit, err := mysqlClient.Scalar("SELECT COUNT(1) FROM go_TypeTest WHERE id IN (@p1)", []int{1, 2, 3})
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

func Test_MySqlDbClient_internalDbClient_get_smoke(t *testing.T) {
	mysqlClient := getMysqlClientOrSkip(t)

	got, err := mysqlClient.Get("SELECT varcharTest, decimalTest, bitTest FROM go_TypeTest WHERE id=@p1", 1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !reflect.DeepEqual(got["varcharTest"], "行1") {
		t.Fatalf("Get()[varcharTest] = %v, want %v", got["varcharTest"], "行1")
	}
	if !reflect.DeepEqual(got["decimalTest"], "1.4567899900") {
		t.Fatalf("Get()[decimalTest] = %v, want %v", got["decimalTest"], "1.4567899900")
	}
	if !reflect.DeepEqual(got["bitTest"], []byte{1}) {
		t.Fatalf("Get()[bitTest] = %v, want %v", got["bitTest"], []byte{1})
	}
}

func Test_MySqlDbClient_internalDbClient_rows_error_wrapped_smoke(t *testing.T) {
	mysqlClient := getMysqlClientOrSkip(t)

	rows, err := mysqlClient.Rows("SELECT varcharTest FROM go_TypeTest WHERE id=@p1", 1)
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
}
