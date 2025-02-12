package mysql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/bunnier/sqlmer"
	"golang.org/x/sync/errgroup"
)

func Test_parseMySqlNamedSql(t *testing.T) {
	assertNamedSql := func(t *testing.T, inputSql, expectedSql, expectedParams string) {
		namedParsedResult := parseMySqlNamedSql(inputSql)
		if namedParsedResult.Sql != expectedSql || strings.Join(namedParsedResult.Names, ",") != expectedParams {
			t.Errorf("expected sql=%s, param=%s\nActual sql=%s, param=%s",
				expectedSql, expectedParams,
				namedParsedResult.Sql, strings.Join(namedParsedResult.Names, ","))
		}
	}

	t.Run("escape_at_symbol", func(t *testing.T) {
		inputSql := "SELECT * FROM go_TypeTest WHERE @@id=1"
		expectedSql := "SELECT * FROM go_TypeTest WHERE @id=1"
		expectedParams := ""
		assertNamedSql(t, inputSql, expectedSql, expectedParams)
	})

	t.Run("single_param", func(t *testing.T) {
		inputSql := "SELECT * FROM go_TypeTest WHERE id=@id"
		expectedSql := "SELECT * FROM go_TypeTest WHERE id=?"
		expectedParams := "id"
		assertNamedSql(t, inputSql, expectedSql, expectedParams)
	})

	t.Run("duplicate_param", func(t *testing.T) {
		inputSql := "SELECT * FROM go_TypeTest WHERE id=@id AND id=@id"
		expectedSql := "SELECT * FROM go_TypeTest WHERE id=? AND id=?"
		expectedParams := "id,id"
		assertNamedSql(t, inputSql, expectedSql, expectedParams)
	})

	t.Run("string_literal_param", func(t *testing.T) {
		inputSql := "SELECT * FROM go_TypeTest WHERE id=@id AND varcharTest='@varcharTest'"
		expectedSql := "SELECT * FROM go_TypeTest WHERE id=? AND varcharTest='@varcharTest'"
		expectedParams := "id"
		assertNamedSql(t, inputSql, expectedSql, expectedParams)
	})

	t.Run("multiple_params", func(t *testing.T) {
		inputSql := "SELECT * FROM go_TypeTest WHERE id=@id AND varcharTest=@varcharTest"
		expectedSql := "SELECT * FROM go_TypeTest WHERE id=? AND varcharTest=?"
		expectedParams := "id,varcharTest"
		assertNamedSql(t, inputSql, expectedSql, expectedParams)
	})

	t.Run("different_order_params", func(t *testing.T) {
		inputSql := "SELECT * FROM go_TypeTest WHERE varcharTest=@varcharTest AND id=@id"
		expectedSql := "SELECT * FROM go_TypeTest WHERE varcharTest=? AND id=?"
		expectedParams := "varcharTest,id"
		assertNamedSql(t, inputSql, expectedSql, expectedParams)
	})

	t.Run("no_params", func(t *testing.T) {
		inputSql := "SELECT * FROM go_TypeTest WHERE varcharTest='@varcharTest' AND id='@id'"
		expectedSql := "SELECT * FROM go_TypeTest WHERE varcharTest='@varcharTest' AND id='@id'"
		expectedParams := ""
		assertNamedSql(t, inputSql, expectedSql, expectedParams)
	})

	// 并发测试
	t.Run("concurrent_test", func(t *testing.T) {
		var errGroup errgroup.Group
		for i := 0; i < 10; i++ {
			errGroup.Go(func() error {
				// 测试一组典型的SQL语句
				testCases := map[string][]string{
					"SELECT * FROM go_TypeTest WHERE id=@id":                              {"SELECT * FROM go_TypeTest WHERE id=?", "id"},
					"SELECT * FROM go_TypeTest WHERE id=@id AND varcharTest=@varcharTest": {"SELECT * FROM go_TypeTest WHERE id=? AND varcharTest=?", "id,varcharTest"},
				}

				for inputSql, expected := range testCases {
					namedParsedResult := parseMySqlNamedSql(inputSql)
					if namedParsedResult.Sql != expected[0] || strings.Join(namedParsedResult.Names, ",") != expected[1] {
						return fmt.Errorf("expected sql=%s, param=%s\nActual sql=%s, param=%s",
							expected[0], expected[1],
							namedParsedResult.Sql, strings.Join(namedParsedResult.Names, ","))
					}
				}
				return nil
			})
		}

		if err := errGroup.Wait(); err != nil {
			t.Errorf("concurrent test failed: %v", err)
		}
	})
}

func Test_extendInParams(t *testing.T) {
	assertInParams := func(t *testing.T, sql string, params []any, expSQL string, expParams []any) {
		gotSQL, gotParams := extendInParams(sql, params)
		if gotSQL != expSQL {
			t.Errorf("Expected SQL: %s, got: %s", expSQL, gotSQL)
		}
		if !reflect.DeepEqual(gotParams, expParams) {
			t.Errorf("Expected Params: %v, got: %v", expParams, gotParams)
		}
	}

	t.Run("single", func(t *testing.T) {
		sql := "select 1 from t where id = ?"
		params := []any{1}
		expSQL := "select 1 from t where id = ?"
		expParams := []any{1}
		assertInParams(t, sql, params, expSQL, expParams)
	})

	t.Run("slice1", func(t *testing.T) {
		sql := "select 1 from t where id in (?)"
		params := []any{[]int{1, 2, 3}}
		expSQL := "select 1 from t where id in (?,?,?)"
		expParams := []any{1, 2, 3}
		assertInParams(t, sql, params, expSQL, expParams)
	})

	t.Run("singleWithSlice", func(t *testing.T) {
		sql := "select 1 from t where id!=? AND id in (?)"
		params := []any{5, []int{1, 2, 3}}
		expSQL := "select 1 from t where id!=? AND id in (?,?,?)"
		expParams := []any{5, 1, 2, 3}
		assertInParams(t, sql, params, expSQL, expParams)
	})

	t.Run("empty", func(t *testing.T) {
		sql := "select 1 from t where id in (?)"
		params := []any{[]int{}}
		expSQL := "select 1 from t where id in (NULL)"
		expParams := []any{}
		assertInParams(t, sql, params, expSQL, expParams)
	})

	t.Run("regular", func(t *testing.T) {
		sql := "select 1 from t where name = ? and age = ?"
		params := []any{"Alice", 30}
		expSQL := "select 1 from t where name = ? and age = ?"
		expParams := []any{"Alice", 30}
		assertInParams(t, sql, params, expSQL, expParams)
	})
}

func Test_bindMySqlArgs(t *testing.T) {
	t.Run("map1", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE id=@id"
		args := []any{
			map[string]any{
				"id": 1,
			},
		}
		wantSql := "SELECT * FROM go_TypeTest WHERE id=?"
		wantParam := []any{1}

		fixedSql, gotArgs, err := bindArgs(oriSql, args...)
		if err != nil {
			t.Error(err)
			return
		}
		if fixedSql != wantSql {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() sql = %v, wantSql %v", fixedSql, wantSql)
		}
		if !reflect.DeepEqual(gotArgs, wantParam) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() args = %v, wantParam %v", gotArgs, wantParam)
		}
	})

	t.Run("map2", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE idv2=@id_id"
		args := []any{
			map[string]any{
				"id_id": 1,
			},
		}
		wantSql := "SELECT * FROM go_TypeTest WHERE idv2=?"
		wantParam := []any{1}

		fixedSql, gotArgs, err := bindArgs(oriSql, args...)
		if err != nil {
			t.Error(err)
			return
		}
		if fixedSql != wantSql {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() sql = %v, wantSql %v", fixedSql, wantSql)
		}
		if !reflect.DeepEqual(gotArgs, wantParam) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() args = %v, wantParam %v", gotArgs, wantParam)
		}
	})

	t.Run("map3", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE idv2=@id_id AND id=@id"
		args := []any{
			map[string]any{
				"id_id": 1,
				"id":    2,
			},
		}
		wantSql := "SELECT * FROM go_TypeTest WHERE idv2=? AND id=?"
		wantParam := []any{1, 2}

		fixedSql, gotArgs, err := bindArgs(oriSql, args...)
		if err != nil {
			t.Error(err)
			return
		}
		if fixedSql != wantSql {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() sql = %v, wantSql %v", fixedSql, wantSql)
		}
		if !reflect.DeepEqual(gotArgs, wantParam) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() args = %v, wantParam %v", gotArgs, wantParam)
		}
	})

	t.Run("map_name_err", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE id=@id1 OR id=@id2"
		args := []any{
			map[string]any{
				"id": 1,
			},
		}

		_, _, err := bindArgs(oriSql, args...)
		if !errors.Is(err, sqlmer.ErrParseParamFailed) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() error = %v, wantErr %v", err, sqlmer.ErrParseParamFailed)
		}
	})

	t.Run("index", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE id=@p1"
		args := []any{1}
		wantSql := "SELECT * FROM go_TypeTest WHERE id=?"
		wantParam := []any{1}

		fixedSql, gotArgs, err := bindArgs(oriSql, args...)
		if err != nil {
			t.Error(err)
			return
		}
		if fixedSql != wantSql {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() sql = %v, wantSql %v", fixedSql, wantSql)
		}
		if !reflect.DeepEqual(gotArgs, wantParam) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() args = %v, wantParam %v", gotArgs, wantParam)
		}
	})

	t.Run("index_index_err1", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE id=@p1 AND id=@p2"
		args := []any{1}

		_, _, err := bindArgs(oriSql, args...)
		if !errors.Is(err, sqlmer.ErrParseParamFailed) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() error = %v, wantErr %v", err, sqlmer.ErrParseParamFailed)
		}
	})

	t.Run("index_index_err2", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE id=@p3"
		args := []any{1}

		_, _, err := bindArgs(oriSql, args...)
		if !errors.Is(err, sqlmer.ErrParseParamFailed) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() error = %v, wantErr %v", err, sqlmer.ErrParseParamFailed)
		}
	})

	t.Run("index_index_err3", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE id=@test"
		args := []any{1}

		_, _, err := bindArgs(oriSql, args...)
		if !errors.Is(err, sqlmer.ErrParseParamFailed) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() error = %v, wantErr %v", err, sqlmer.ErrParseParamFailed)
		}
	})

	t.Run("index_index_err4", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE id=@pttt"
		args := []any{1}

		_, _, err := bindArgs(oriSql, args...)
		if !errors.Is(err, sqlmer.ErrParseParamFailed) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() error = %v, wantErr %v", err, sqlmer.ErrParseParamFailed)
		}
	})

	t.Run("index_reuse_index", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE id=@p1 AND id=@p1"
		args := []any{1}
		wantSql := "SELECT * FROM go_TypeTest WHERE id=? AND id=?"
		wantParam := []any{1, 1}

		fixedSql, gotArgs, err := bindArgs(oriSql, args...)
		if err != nil {
			t.Error(err)
			return
		}
		if fixedSql != wantSql {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() sql = %v, wantSql %v", fixedSql, wantSql)
		}
		if !reflect.DeepEqual(gotArgs, wantParam) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() args = %v, wantParam %v", gotArgs, wantParam)
		}
	})

	t.Run("params_more_than_names", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE id=@p3 AND id=@p3"
		args := []any{1, 2, 3, 4, 5, 6, 7}
		wantSql := "SELECT * FROM go_TypeTest WHERE id=? AND id=?"
		wantParam := []any{3, 3}

		fixedSql, gotArgs, err := bindArgs(oriSql, args...)
		if err != nil {
			t.Error(err)
			return
		}
		if fixedSql != wantSql {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() sql = %v, wantSql %v", fixedSql, wantSql)
		}
		if !reflect.DeepEqual(gotArgs, wantParam) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() args = %v, wantParam %v", gotArgs, wantParam)
		}
	})

	t.Run("inwhere1", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE id IN (@ids)"
		args := []any{
			map[string]any{
				"ids": []int{1, 2, 3},
			},
		}
		wantSql := "SELECT * FROM go_TypeTest WHERE id IN (?,?,?)"
		wantParam := []any{1, 2, 3}

		fixedSql, gotArgs, err := bindArgs(oriSql, args...)
		if err != nil {
			t.Error(err)
			return
		}
		if fixedSql != wantSql {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() sql = %v, wantSql %v", fixedSql, wantSql)
		}
		if !reflect.DeepEqual(gotArgs, wantParam) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() args = %v, wantParam %v", gotArgs, wantParam)
		}
	})

	t.Run("inwhere2", func(t *testing.T) {
		oriSql := "SELECT * FROM go_TypeTest WHERE id!=@noid AND id IN (@ids)"
		args := []any{
			map[string]any{
				"noid": 4,
				"ids":  []int{1, 2, 3},
			},
		}
		wantSql := "SELECT * FROM go_TypeTest WHERE id!=? AND id IN (?,?,?)"
		wantParam := []any{4, 1, 2, 3}

		fixedSql, gotArgs, err := bindArgs(oriSql, args...)
		if err != nil {
			t.Error(err)
			return
		}
		if fixedSql != wantSql {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() sql = %v, wantSql %v", fixedSql, wantSql)
		}
		if !reflect.DeepEqual(gotArgs, wantParam) {
			t.Errorf("mysqlDbClient.bindMsSqlArgs() args = %v, wantParam %v", gotArgs, wantParam)
		}
	})
}
