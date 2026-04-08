package qm_namedsql

import (
	"errors"
	"reflect"
	"testing"

	"github.com/bunnier/sqlmer"
)

// BindQuestionMarkArgs 为 MySQL、SQLite 共用实现，集成测试集中在本文件；驱动包不再重复测试。

func Test_BindQuestionMarkArgs_basic_map_binding(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE id=@id"
	args := []any{
		map[string]any{
			"id": 1,
		},
	}
	wantSql := "SELECT * FROM go_TypeTest WHERE id=?"
	wantParam := []any{1}

	fixedSql, gotArgs, err := BindQuestionMarkArgs(oriSql, args...)
	if err != nil {
		t.Fatal(err)
	}
	if fixedSql != wantSql {
		t.Errorf("got sql = %v, want %v", fixedSql, wantSql)
	}
	if !reflect.DeepEqual(gotArgs, wantParam) {
		t.Errorf("got args = %v, want %v", gotArgs, wantParam)
	}
}

func Test_BindQuestionMarkArgs_map_binding_with_underscore(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE idv2=@id_id"
	args := []any{
		map[string]any{
			"id_id": 1,
		},
	}
	wantSql := "SELECT * FROM go_TypeTest WHERE idv2=?"
	wantParam := []any{1}

	fixedSql, gotArgs, err := BindQuestionMarkArgs(oriSql, args...)
	if err != nil {
		t.Fatal(err)
	}
	if fixedSql != wantSql {
		t.Errorf("got sql = %v, want %v", fixedSql, wantSql)
	}
	if !reflect.DeepEqual(gotArgs, wantParam) {
		t.Errorf("got args = %v, want %v", gotArgs, wantParam)
	}
}

func Test_BindQuestionMarkArgs_map_binding_multiple_params(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE idv2=@id_id AND id=@id"
	args := []any{
		map[string]any{
			"id_id": 1,
			"id":    2,
		},
	}
	wantSql := "SELECT * FROM go_TypeTest WHERE idv2=? AND id=?"
	wantParam := []any{1, 2}

	fixedSql, gotArgs, err := BindQuestionMarkArgs(oriSql, args...)
	if err != nil {
		t.Fatal(err)
	}
	if fixedSql != wantSql {
		t.Errorf("got sql = %v, want %v", fixedSql, wantSql)
	}
	if !reflect.DeepEqual(gotArgs, wantParam) {
		t.Errorf("got args = %v, want %v", gotArgs, wantParam)
	}
}

func Test_BindQuestionMarkArgs_map_binding_missing_param(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE id=@id1 OR id=@id2"
	args := []any{
		map[string]any{
			"id": 1,
		},
	}

	_, _, err := BindQuestionMarkArgs(oriSql, args...)
	if !errors.Is(err, sqlmer.ErrParseParamFailed) {
		t.Errorf("error = %v, want %v", err, sqlmer.ErrParseParamFailed)
	}
}

func Test_BindQuestionMarkArgs_basic_index_binding(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE id=@p1"
	args := []any{1}
	wantSql := "SELECT * FROM go_TypeTest WHERE id=?"
	wantParam := []any{1}

	fixedSql, gotArgs, err := BindQuestionMarkArgs(oriSql, args...)
	if err != nil {
		t.Fatal(err)
	}
	if fixedSql != wantSql {
		t.Errorf("got sql = %v, want %v", fixedSql, wantSql)
	}
	if !reflect.DeepEqual(gotArgs, wantParam) {
		t.Errorf("got args = %v, want %v", gotArgs, wantParam)
	}
}

func Test_BindQuestionMarkArgs_index_binding_insufficient_params(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE id=@p1 AND id=@p2"
	args := []any{1}

	_, _, err := BindQuestionMarkArgs(oriSql, args...)
	if !errors.Is(err, sqlmer.ErrParseParamFailed) {
		t.Errorf("error = %v, want %v", err, sqlmer.ErrParseParamFailed)
	}
}

func Test_BindQuestionMarkArgs_index_binding_invalid_reference(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE id=@p3"
	args := []any{1}

	_, _, err := BindQuestionMarkArgs(oriSql, args...)
	if !errors.Is(err, sqlmer.ErrParseParamFailed) {
		t.Errorf("error = %v, want %v", err, sqlmer.ErrParseParamFailed)
	}
}

func Test_BindQuestionMarkArgs_invalid_param_format(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE id=@test"
	args := []any{1}

	_, _, err := BindQuestionMarkArgs(oriSql, args...)
	if !errors.Is(err, sqlmer.ErrParseParamFailed) {
		t.Errorf("error = %v, want %v", err, sqlmer.ErrParseParamFailed)
	}
}

func Test_BindQuestionMarkArgs_invalid_param_name(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE id=@pttt"
	args := []any{1}

	_, _, err := BindQuestionMarkArgs(oriSql, args...)
	if !errors.Is(err, sqlmer.ErrParseParamFailed) {
		t.Errorf("error = %v, want %v", err, sqlmer.ErrParseParamFailed)
	}
}

func Test_BindQuestionMarkArgs_reuse_index_param(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE id=@p1 AND id=@p1"
	args := []any{1}
	wantSql := "SELECT * FROM go_TypeTest WHERE id=? AND id=?"
	wantParam := []any{1, 1}

	fixedSql, gotArgs, err := BindQuestionMarkArgs(oriSql, args...)
	if err != nil {
		t.Fatal(err)
	}
	if fixedSql != wantSql {
		t.Errorf("got sql = %v, want %v", fixedSql, wantSql)
	}
	if !reflect.DeepEqual(gotArgs, wantParam) {
		t.Errorf("got args = %v, want %v", gotArgs, wantParam)
	}
}

func Test_BindQuestionMarkArgs_excess_params(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE id=@p3 AND id=@p3"
	args := []any{1, 2, 3, 4, 5, 6, 7}
	wantSql := "SELECT * FROM go_TypeTest WHERE id=? AND id=?"
	wantParam := []any{3, 3}

	fixedSql, gotArgs, err := BindQuestionMarkArgs(oriSql, args...)
	if err != nil {
		t.Fatal(err)
	}
	if fixedSql != wantSql {
		t.Errorf("got sql = %v, want %v", fixedSql, wantSql)
	}
	if !reflect.DeepEqual(gotArgs, wantParam) {
		t.Errorf("got args = %v, want %v", gotArgs, wantParam)
	}
}

func Test_BindQuestionMarkArgs_array_binding_in_clause(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE id IN (@ids)"
	args := []any{
		map[string]any{
			"ids": []int{1, 2, 3},
		},
	}
	wantSql := "SELECT * FROM go_TypeTest WHERE id IN (?,?,?)"
	wantParam := []any{1, 2, 3}

	fixedSql, gotArgs, err := BindQuestionMarkArgs(oriSql, args...)
	if err != nil {
		t.Fatal(err)
	}
	if fixedSql != wantSql {
		t.Errorf("got sql = %v, want %v", fixedSql, wantSql)
	}
	if !reflect.DeepEqual(gotArgs, wantParam) {
		t.Errorf("got args = %v, want %v", gotArgs, wantParam)
	}
}

func Test_BindQuestionMarkArgs_complex_binding_with_array(t *testing.T) {
	oriSql := "SELECT * FROM go_TypeTest WHERE id!=@noid AND id IN (@ids)"
	args := []any{
		map[string]any{
			"noid": 4,
			"ids":  []int{1, 2, 3},
		},
	}
	wantSql := "SELECT * FROM go_TypeTest WHERE id!=? AND id IN (?,?,?)"
	wantParam := []any{4, 1, 2, 3}

	fixedSql, gotArgs, err := BindQuestionMarkArgs(oriSql, args...)
	if err != nil {
		t.Fatal(err)
	}
	if fixedSql != wantSql {
		t.Errorf("got sql = %v, want %v", fixedSql, wantSql)
	}
	if !reflect.DeepEqual(gotArgs, wantParam) {
		t.Errorf("got args = %v, want %v", gotArgs, wantParam)
	}
}
