package sqlmer

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
)

// getMsSqlDbClient 用于 获取一个 SqlServer 数据库的DbClient对象。
func getMsSqlDbClient() (DbClient, error) {
	return NewMsSqlDbClient(
		testConf.SqlServer,
		WithConnTimeout(time.Second*15),
		WithExecTimeout(time.Second*15),
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
		WithConnTimeout(time.Second*15),
		WithExecTimeout(time.Second*15))
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
