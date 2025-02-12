package mssql

import (
	"database/sql"
	"reflect"
	"testing"
)

func Test_bindMsSqlArgs(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE Id=@id"
		args := []any{
			map[string]any{
				"id": 1,
			},
		}
		wantSql := "SELECT * FROM go_TypeTest WHERE Id=@id"
		wantParam := []any{sql.Named("id", 1)}

		fixedSql, gotArgs, err := bindArgs(oriSql, args...)
		if err != nil {
			t.Error(err)
			return
		}

		if fixedSql != wantSql {
			t.Errorf("bindArgs() sql = %v, wantSql %v", fixedSql, wantSql)
		}

		if !reflect.DeepEqual(gotArgs, wantParam) {
			t.Errorf("bindArgs() args = %v, wantParam %v", gotArgs, wantParam)
		}
	})

	t.Run("index", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE Id=@p1 OR Id=@p2"
		args := []any{1, 2}
		wantSql := "SELECT * FROM go_TypeTest WHERE Id=@p1 OR Id=@p2"
		wantParam := []any{1, 2}

		fixedSql, gotArgs, err := bindArgs(oriSql, args...)
		if err != nil {
			t.Error(err)
			return
		}

		if fixedSql != wantSql {
			t.Errorf("bindArgs() sql = %v, wantSql %v", fixedSql, wantSql)
		}

		if !reflect.DeepEqual(gotArgs, wantParam) {
			t.Errorf("bindArgs() args = %v, wantParam %v", gotArgs, wantParam)
		}
	})
}
