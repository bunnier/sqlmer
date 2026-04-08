package qm_namedsql

import (
	"reflect"
	"testing"
)

func Test_extendInParams_single(t *testing.T) {
	sql := "select 1 from t where id = ?"
	params := []any{1}
	expSQL := "select 1 from t where id = ?"
	expParams := []any{1}

	gotSQL, gotParams := extendInParams(sql, params)
	if gotSQL != expSQL {
		t.Errorf("expected sql=%s, got=%s", expSQL, gotSQL)
	}
	if !reflect.DeepEqual(gotParams, expParams) {
		t.Errorf("expected params=%v, got=%v", expParams, gotParams)
	}
}

func Test_extendInParams_slice(t *testing.T) {
	sql := "select 1 from t where id in (?)"
	params := []any{[]int{1, 2, 3}}
	expSQL := "select 1 from t where id in (?,?,?)"
	expParams := []any{1, 2, 3}

	gotSQL, gotParams := extendInParams(sql, params)
	if gotSQL != expSQL {
		t.Errorf("expected sql=%s, got=%s", expSQL, gotSQL)
	}
	if !reflect.DeepEqual(gotParams, expParams) {
		t.Errorf("expected params=%v, got=%v", expParams, gotParams)
	}
}

func Test_extendInParams_single_with_slice(t *testing.T) {
	sql := "select 1 from t where id!=? AND id in (?)"
	params := []any{5, []int{1, 2, 3}}
	expSQL := "select 1 from t where id!=? AND id in (?,?,?)"
	expParams := []any{5, 1, 2, 3}

	gotSQL, gotParams := extendInParams(sql, params)
	if gotSQL != expSQL {
		t.Errorf("expected sql=%s, got=%s", expSQL, gotSQL)
	}
	if !reflect.DeepEqual(gotParams, expParams) {
		t.Errorf("expected params=%v, got=%v", expParams, gotParams)
	}
}

func Test_extendInParams_empty_slice(t *testing.T) {
	sql := "select 1 from t where id in (?)"
	params := []any{[]int{}}
	expSQL := "select 1 from t where id in (NULL)"
	expParams := []any{}

	gotSQL, gotParams := extendInParams(sql, params)
	if gotSQL != expSQL {
		t.Errorf("expected sql=%s, got=%s", expSQL, gotSQL)
	}
	if !reflect.DeepEqual(gotParams, expParams) {
		t.Errorf("expected params=%v, got=%v", expParams, gotParams)
	}
}

func Test_extendInParams_regular(t *testing.T) {
	sql := "select 1 from t where name = ? and age = ?"
	params := []any{"Alice", 30}
	expSQL := "select 1 from t where name = ? and age = ?"
	expParams := []any{"Alice", 30}

	gotSQL, gotParams := extendInParams(sql, params)
	if gotSQL != expSQL {
		t.Errorf("expected sql=%s, got=%s", expSQL, gotSQL)
	}
	if !reflect.DeepEqual(gotParams, expParams) {
		t.Errorf("expected params=%v, got=%v", expParams, gotParams)
	}
}
