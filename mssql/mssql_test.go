package mssql

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
)

func Test_bindMsSqlArgs(t *testing.T) {
	testCases := []struct {
		name      string
		oriSql    string
		args      []any
		wantSql   string
		wantParam []any
		wantErr   error
	}{
		{
			"map",
			"SELECT * FROM go_TypeTest WHERE Id=@id",
			[]any{
				map[string]any{
					"id": 1,
				},
			},
			"SELECT * FROM go_TypeTest WHERE Id=@id",
			[]any{sql.Named("id", 1)},
			nil,
		},
		{
			"index",
			"SELECT * FROM go_TypeTest WHERE Id=@p1 OR Id=@p2",
			[]any{1, 2},
			"SELECT * FROM go_TypeTest WHERE Id=@p1 OR Id=@p2",
			[]any{1, 2},
			nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			fixedSql, args, err := bindArgs(tt.oriSql, tt.args...)

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
