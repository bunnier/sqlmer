package named2qm

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/sync/errgroup"
)

func Test_ParseNamedSqlToQuestionMark_escape_at_symbol(t *testing.T) {
	inputSql := "SELECT * FROM go_TypeTest WHERE @@id=1"
	expectedSql := "SELECT * FROM go_TypeTest WHERE @id=1"
	expectedParams := ""

	result := ParseNamedSqlToQuestionMark(inputSql)
	if result.Sql != expectedSql || strings.Join(result.Names, ",") != expectedParams {
		t.Errorf("expected sql=%s, param=%s\nActual sql=%s, param=%s",
			expectedSql, expectedParams, result.Sql, strings.Join(result.Names, ","))
	}
}

func Test_ParseNamedSqlToQuestionMark_single_param(t *testing.T) {
	inputSql := "SELECT * FROM go_TypeTest WHERE id=@id"
	expectedSql := "SELECT * FROM go_TypeTest WHERE id=?"
	expectedParams := "id"

	result := ParseNamedSqlToQuestionMark(inputSql)
	if result.Sql != expectedSql || strings.Join(result.Names, ",") != expectedParams {
		t.Errorf("expected sql=%s, param=%s\nActual sql=%s, param=%s",
			expectedSql, expectedParams, result.Sql, strings.Join(result.Names, ","))
	}
}

func Test_ParseNamedSqlToQuestionMark_duplicate_param(t *testing.T) {
	inputSql := "SELECT * FROM go_TypeTest WHERE id=@id AND id=@id"
	expectedSql := "SELECT * FROM go_TypeTest WHERE id=? AND id=?"
	expectedParams := "id,id"

	result := ParseNamedSqlToQuestionMark(inputSql)
	if result.Sql != expectedSql || strings.Join(result.Names, ",") != expectedParams {
		t.Errorf("expected sql=%s, param=%s\nActual sql=%s, param=%s",
			expectedSql, expectedParams, result.Sql, strings.Join(result.Names, ","))
	}
}

func Test_ParseNamedSqlToQuestionMark_string_literal_param(t *testing.T) {
	inputSql := "SELECT * FROM go_TypeTest WHERE id=@id AND varcharTest='@varcharTest'"
	expectedSql := "SELECT * FROM go_TypeTest WHERE id=? AND varcharTest='@varcharTest'"
	expectedParams := "id"

	result := ParseNamedSqlToQuestionMark(inputSql)
	if result.Sql != expectedSql || strings.Join(result.Names, ",") != expectedParams {
		t.Errorf("expected sql=%s, param=%s\nActual sql=%s, param=%s",
			expectedSql, expectedParams, result.Sql, strings.Join(result.Names, ","))
	}
}

func Test_ParseNamedSqlToQuestionMark_multiple_params(t *testing.T) {
	inputSql := "SELECT * FROM go_TypeTest WHERE id=@id AND varcharTest=@varcharTest"
	expectedSql := "SELECT * FROM go_TypeTest WHERE id=? AND varcharTest=?"
	expectedParams := "id,varcharTest"

	result := ParseNamedSqlToQuestionMark(inputSql)
	if result.Sql != expectedSql || strings.Join(result.Names, ",") != expectedParams {
		t.Errorf("expected sql=%s, param=%s\nActual sql=%s, param=%s",
			expectedSql, expectedParams, result.Sql, strings.Join(result.Names, ","))
	}
}

func Test_ParseNamedSqlToQuestionMark_different_order_params(t *testing.T) {
	inputSql := "SELECT * FROM go_TypeTest WHERE varcharTest=@varcharTest AND id=@id"
	expectedSql := "SELECT * FROM go_TypeTest WHERE varcharTest=? AND id=?"
	expectedParams := "varcharTest,id"

	result := ParseNamedSqlToQuestionMark(inputSql)
	if result.Sql != expectedSql || strings.Join(result.Names, ",") != expectedParams {
		t.Errorf("expected sql=%s, param=%s\nActual sql=%s, param=%s",
			expectedSql, expectedParams, result.Sql, strings.Join(result.Names, ","))
	}
}

func Test_ParseNamedSqlToQuestionMark_no_params(t *testing.T) {
	inputSql := "SELECT * FROM go_TypeTest WHERE varcharTest='@varcharTest' AND id='@id'"
	expectedSql := "SELECT * FROM go_TypeTest WHERE varcharTest='@varcharTest' AND id='@id'"
	expectedParams := ""

	result := ParseNamedSqlToQuestionMark(inputSql)
	if result.Sql != expectedSql || strings.Join(result.Names, ",") != expectedParams {
		t.Errorf("expected sql=%s, param=%s\nActual sql=%s, param=%s",
			expectedSql, expectedParams, result.Sql, strings.Join(result.Names, ","))
	}
}

func Test_ParseNamedSqlToQuestionMark_concurrent(t *testing.T) {
	var errGroup errgroup.Group
	for i := 0; i < 10; i++ {
		errGroup.Go(func() error {
			result1 := ParseNamedSqlToQuestionMark("SELECT * FROM go_TypeTest WHERE id=@id")
			if result1.Sql != "SELECT * FROM go_TypeTest WHERE id=?" || strings.Join(result1.Names, ",") != "id" {
				return fmt.Errorf("case1: got sql=%s names=%s", result1.Sql, strings.Join(result1.Names, ","))
			}
			result2 := ParseNamedSqlToQuestionMark("SELECT * FROM go_TypeTest WHERE id=@id AND varcharTest=@varcharTest")
			if result2.Sql != "SELECT * FROM go_TypeTest WHERE id=? AND varcharTest=?" || strings.Join(result2.Names, ",") != "id,varcharTest" {
				return fmt.Errorf("case2: got sql=%s names=%s", result2.Sql, strings.Join(result2.Names, ","))
			}
			return nil
		})
	}
	if err := errGroup.Wait(); err != nil {
		t.Errorf("concurrent test failed: %v", err)
	}
}

func Test_ParseNamedSqlToQuestionMark_no_params_not_cached(t *testing.T) {
	noParamSql := "SELECT * FROM t WHERE id = 999_no_cache_marker"
	ParseNamedSqlToQuestionMark(noParamSql)

	_, ok := parsedSqlCache.load(noParamSql)
	if ok {
		t.Error("expected no-param SQL to not be cached, but it was")
	}
}

func Test_ParseNamedSqlToQuestionMark_with_params_is_cached(t *testing.T) {
	paramSql := "SELECT * FROM t WHERE id=@qm_namedsql_cache_test_id"
	result := ParseNamedSqlToQuestionMark(paramSql)

	cached, ok := parsedSqlCache.load(paramSql)
	if !ok {
		t.Error("expected parameterized SQL to be cached, but it was not")
	}
	if cached.Sql != result.Sql {
		t.Errorf("cached sql mismatch: expected=%s, got=%s", result.Sql, cached.Sql)
	}
}
