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
	testSqls := map[string][]string{
		"SELECT * FROM go_TypeTest WHERE @@id=1":                                  {"SELECT * FROM go_TypeTest WHERE @id=1", ""},
		"SELECT * FROM go_TypeTest WHERE id=@id":                                  {"SELECT * FROM go_TypeTest WHERE id=?", "id"},
		"SELECT * FROM go_TypeTest WHERE id=@id AND id=@id":                       {"SELECT * FROM go_TypeTest WHERE id=? AND id=?", "id,id"},
		"SELECT * FROM go_TypeTest WHERE id=@id AND varcharTest='@varcharTest'":   {"SELECT * FROM go_TypeTest WHERE id=? AND varcharTest='@varcharTest'", "id"},
		"SELECT * FROM go_TypeTest WHERE id=@id AND varcharTest=@varcharTest":     {"SELECT * FROM go_TypeTest WHERE id=? AND varcharTest=?", "id,varcharTest"},
		"SELECT * FROM go_TypeTest WHERE varcharTest=@varcharTest AND id=@id":     {"SELECT * FROM go_TypeTest WHERE varcharTest=? AND id=?", "varcharTest,id"},
		"SELECT * FROM go_TypeTest WHERE varcharTest='@varcharTest' AND id='@id'": {"SELECT * FROM go_TypeTest WHERE varcharTest='@varcharTest' AND id='@id'", ""},
	}

	var errGroup errgroup.Group
	for i := 0; i < 10; i++ { // 这边测试下并发，开10个goroutine并行测试。
		errGroup.Go(func() error {
			for inputSql, expected := range testSqls {
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
		t.Errorf("mysqlDbClient.parseMySqlNamedSql() err = %v, wantErr = nil", err)
		return
	}
}

func TestExtendInParams(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		params    []any
		expSQL    string
		expParams []any
	}{
		{
			"single",
			"select 1 from t where id = ?",
			[]any{1},
			"select 1 from t where id = ?",
			[]any{1},
		},
		{
			"slice1",
			"select 1 from t where id in (?)",
			[]any{[]int{1, 2, 3}},
			"select 1 from t where id in (?,?,?)",
			[]any{1, 2, 3},
		},
		{
			"singleWithSlice",
			"select 1 from t where id!=? AND id in (?)",
			[]any{5, []int{1, 2, 3}},
			"select 1 from t where id!=? AND id in (?,?,?)",
			[]any{5, 1, 2, 3},
		},
		{
			"empty",
			"select 1 from t where id in (?)",
			[]any{[]int{}},
			"select 1 from t where id in (NULL)",
			[]any{},
		},
		{
			"regular",
			"select 1 from t where name = ? and age = ?",
			[]any{"Alice", 30},
			"select 1 from t where name = ? and age = ?",
			[]any{"Alice", 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotParams := extendInParams(tt.sql, tt.params)
			if gotSQL != tt.expSQL {
				t.Errorf("Expected SQL: %s, got: %s", tt.expSQL, gotSQL)
			}
			if !reflect.DeepEqual(gotParams, tt.expParams) {
				t.Errorf("Expected Params: %v, got: %v", tt.expParams, gotParams)
			}
		})
	}
}

func Test_bindMySqlArgs(t *testing.T) {
	testCases := []struct {
		name      string
		oriSql    string
		args      []any
		wantSql   string
		wantParam []any
		wantErr   error
	}{
		{
			"map1",
			"SELECT * FROM go_TypeTest WHERE id=@id",
			[]any{
				map[string]any{
					"id": 1,
				},
			},
			"SELECT * FROM go_TypeTest WHERE id=?",
			[]any{1},
			nil,
		},
		{
			"map2",
			"SELECT * FROM go_TypeTest WHERE idv2=@id_id",
			[]any{
				map[string]any{
					"id_id": 1,
				},
			},
			"SELECT * FROM go_TypeTest WHERE idv2=?",
			[]any{1},
			nil,
		},
		{
			"map3",
			"SELECT * FROM go_TypeTest WHERE idv2=@id_id AND id=@id",
			[]any{
				map[string]any{
					"id_id": 1,
					"id":    2,
				},
			},
			"SELECT * FROM go_TypeTest WHERE idv2=? AND id=?",
			[]any{1, 2},
			nil,
		},
		{
			"map_name_err",
			"SELECT * FROM go_TypeTest WHERE id=@id1 OR id=@id2",
			[]any{
				map[string]any{
					"id": 1,
				},
			},
			"",
			nil,
			sqlmer.ErrParseParamFailed,
		},
		{
			"index",
			"SELECT * FROM go_TypeTest WHERE id=@p1",
			[]any{1},
			"SELECT * FROM go_TypeTest WHERE id=?",
			[]any{1},
			nil,
		},
		{
			"index_index_err1",
			"SELECT * FROM go_TypeTest WHERE id=@p1 AND id=@p2",
			[]any{1},
			"",
			nil,
			sqlmer.ErrParseParamFailed,
		},
		{
			"index_index_err2",
			"SELECT * FROM go_TypeTest WHERE id=@p3",
			[]any{1},
			"",
			nil,
			sqlmer.ErrParseParamFailed,
		},
		{
			"index_index_err3",
			"SELECT * FROM go_TypeTest WHERE id=@test",
			[]any{1},
			"",
			nil,
			sqlmer.ErrParseParamFailed,
		},
		{
			"index_index_err4",
			"SELECT * FROM go_TypeTest WHERE id=@pttt",
			[]any{1},
			"",
			nil,
			sqlmer.ErrParseParamFailed,
		},
		{
			"index_reuse_index",
			"SELECT * FROM go_TypeTest WHERE id=@p1 AND id=@p1",
			[]any{1},
			"SELECT * FROM go_TypeTest WHERE id=? AND id=?",
			[]any{1, 1},
			nil,
		},
		{
			"params_more_than_names",
			"SELECT * FROM go_TypeTest WHERE id=@p3 AND id=@p3",
			[]any{1, 2, 3, 4, 5, 6, 7},
			"SELECT * FROM go_TypeTest WHERE id=? AND id=?",
			[]any{3, 3},
			nil,
		},
		{
			"inwhere1",
			"SELECT * FROM go_TypeTest WHERE id IN (@ids)",
			[]any{
				map[string]any{
					"ids": []int{1, 2, 3},
				},
			},
			"SELECT * FROM go_TypeTest WHERE id IN (?,?,?)",
			[]any{1, 2, 3},
			nil,
		},
		{
			"inwhere2",
			"SELECT * FROM go_TypeTest WHERE id!=@noid AND id IN (@ids)",
			[]any{
				map[string]any{
					"noid": 4,
					"ids":  []int{1, 2, 3},
				},
			},
			"SELECT * FROM go_TypeTest WHERE id!=? AND id IN (?,?,?)",
			[]any{4, 1, 2, 3},
			nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			fixedSql, args, err := bindArgs(tt.oriSql, tt.args...)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("mysqlDbClient.bindMsSqlArgs() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else {
				if err != nil {
					t.Error(err)
					return
				}
				if fixedSql != tt.wantSql {
					t.Errorf("mysqlDbClient.bindMsSqlArgs() sql = %v, wantSql %v", fixedSql, tt.wantSql)
				}

				if !reflect.DeepEqual(args, tt.wantParam) {
					t.Errorf("mysqlDbClient.bindMsSqlArgs() args = %v, wantParam %v", args, tt.wantParam)
				}
			}
		})
	}
}
