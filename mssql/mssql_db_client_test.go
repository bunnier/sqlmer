package mssql

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/testenv"
	"github.com/pkg/errors"
)

// 初始化测试配置。
var testConf testenv.TestConf = testenv.MustLoadTestConfig("../test_conf.yml")

// 用于获取一个 SqlServer 数据库的 DbClient 对象。
func getMsSqlDbClient() (sqlmer.DbClient, error) {
	return NewMsSqlDbClient(
		testConf.SqlServer,
		sqlmer.WithConnTimeout(time.Second*15),
		sqlmer.WithExecTimeout(time.Second*15),
	)
}

func Test_NewMsSqlDbClient(t *testing.T) {
	dbClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}

	if dbClient.ConnectionString() != testConf.SqlServer {
		t.Errorf("mssqlDbClient.ConnectionString() connString = %v, wantConnString %v", dbClient.ConnectionString(), testConf.SqlServer)
	}

	_, err = NewMsSqlDbClient("test",
		sqlmer.WithConnTimeout(time.Second*15),
		sqlmer.WithExecTimeout(time.Second*15))
	if err == nil {
		t.Errorf("mssqlDbClient.NewMsSqlDbClient() err = nil, want has a err")
	}
}

func Test_bindMsSqlArgs(t *testing.T) {
	testCases := []struct {
		name      string
		oriSql    string
		args      []interface{}
		wantSql   string
		wantParam []interface{}
		wantErr   error
	}{
		{
			"map",
			"SELECT * FROM go_TypeTest WHERE Id=@id",
			[]interface{}{
				map[string]interface{}{
					"id": 1,
				},
			},
			"SELECT * FROM go_TypeTest WHERE Id=@id",
			[]interface{}{sql.Named("id", 1)},
			nil,
		},
		{
			"index",
			"SELECT * FROM go_TypeTest WHERE Id=@p1 OR Id=@p2",
			[]interface{}{1, 2},
			"SELECT * FROM go_TypeTest WHERE Id=@p1 OR Id=@p2",
			[]interface{}{1, 2},
			nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			fixedSql, args, err := bindMsSqlArgs(tt.oriSql, tt.args...)

			if tt.wantErr != nil {
				if !errors.As(err, &tt.wantErr) {
					t.Errorf("mssqlDbClient.bindMsSqlArgs() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else {
				if err != nil {
					t.Error(err)
					return
				}
				if fixedSql != tt.wantSql {
					t.Errorf("mssqlDbClient.bindMsSqlArgs() sql = %v, wantSql %v", fixedSql, tt.wantSql)
				}

				if !reflect.DeepEqual(args, tt.wantParam) {
					t.Errorf("mssqlDbClient.bindMsSqlArgs() args = %v, wantParam %v", args, tt.wantParam)
				}
			}
		})
	}
}
