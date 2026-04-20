package sqlmer

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestGetExecutingSqlError(t *testing.T) {
	t.Run("Basic Error Test", func(t *testing.T) {
		err := errors.New("test error")
		rawSql := "SELECT * FROM users WHERE id = @p1"
		fixedSql := "SELECT * FROM users WHERE id = ?"
		params := []any{1}
		wantErr := "dbClient: failed to execute sql\nraw error: test error\nsql:\ninput sql=SELECT * FROM users WHERE id = @p1\nexecuting sql=SELECT * FROM users WHERE id = ?\nparams:\n@p1=1"

		gotErr := getExecutingSqlError(err, rawSql, fixedSql, params)
		if gotErr.Error() != wantErr {
			t.Errorf("getExecutingSqlError() error = %v, want %v", gotErr, wantErr)
		}
	})

	t.Run("Named Parameter Test", func(t *testing.T) {
		err := errors.New("test error")
		rawSql := "SELECT * FROM users WHERE name = @name"
		fixedSql := "SELECT * FROM users WHERE name = ?"
		params := []any{sql.Named("name", "test")}
		wantErr := "dbClient: failed to execute sql\nraw error: test error\nsql:\ninput sql=SELECT * FROM users WHERE name = @name\nexecuting sql=SELECT * FROM users WHERE name = ?\nparams:\n@name=test"

		gotErr := getExecutingSqlError(err, rawSql, fixedSql, params)
		if gotErr.Error() != wantErr {
			t.Errorf("getExecutingSqlError() error = %v, want %v", gotErr, wantErr)
		}
	})
}

func TestGetSqlError(t *testing.T) {
	t.Run("Basic Error Test", func(t *testing.T) {
		err := ErrExpectedSizeWrong
		rawSql := "UPDATE users SET name = @p1"
		params := []any{"test"}
		wantErr := "dbClient: effected rows was wrong\nsql:\ninput sql=UPDATE users SET name = @p1\nparams:\n@p1=test"

		gotErr := getSqlError(err, rawSql, params)
		if gotErr.Error() != wantErr {
			t.Errorf("getSqlError() error = %v, want %v", gotErr, wantErr)
		}
	})

	t.Run("Named Parameter Test", func(t *testing.T) {
		err := ErrParseParamFailed
		rawSql := "INSERT INTO users (name) VALUES (@name)"
		params := []any{sql.Named("name", "test")}
		wantErr := "dbClient: failed to parse named params\nsql:\ninput sql=INSERT INTO users (name) VALUES (@name)\nparams:\n@name=test"

		gotErr := getSqlError(err, rawSql, params)
		if gotErr.Error() != wantErr {
			t.Errorf("getSqlError() error = %v, want %v", gotErr, wantErr)
		}
	})
}

func TestSqlContextErrorUnwrapAndIs(t *testing.T) {
	t.Run("Executing SQL path unwraps driver error", func(t *testing.T) {
		underlying := errors.New("driver rejected")
		err := getExecutingSqlError(underlying, "SELECT 1", "SELECT 1", nil)
		if !errors.Is(err, ErrExecutingSql) {
			t.Fatal("errors.Is(..., ErrExecutingSql) = false, want true")
		}
		if !errors.Is(err, underlying) {
			t.Fatal("errors.Is(..., underlying) = false, want true")
		}
		var sce *SqlContextError
		if !errors.As(err, &sce) {
			t.Fatal("errors.As to *SqlContextError failed")
		}
		if sce.RawSQL != "SELECT 1" || sce.FixedSQL != "SELECT 1" {
			t.Fatalf("SqlContextError fields: got RawSQL=%q FixedSQL=%q", sce.RawSQL, sce.FixedSQL)
		}
	})

	t.Run("Non-executing path unwraps wrapped sentinel", func(t *testing.T) {
		inner := fmt.Errorf("%w: expected: %d, actually: %d", ErrExpectedSizeWrong, 2, 1)
		err := getSqlError(inner, "UPDATE t SET x=1", nil)
		if !errors.Is(err, ErrExpectedSizeWrong) {
			t.Fatal("errors.Is(..., ErrExpectedSizeWrong) = false, want true")
		}
		if errors.Is(err, ErrExecutingSql) {
			t.Fatal("errors.Is(..., ErrExecutingSql) = true, want false")
		}
	})
}

func TestCutLongStringParams(t *testing.T) {
	originalMaxLength := MaxLengthErrorValue
	defer func() { MaxLengthErrorValue = originalMaxLength }()

	// 为了便于测试，将最大长度设置为较小的值。
	MaxLengthErrorValue = 10

	t.Run("Short String", func(t *testing.T) {
		paramVal := "test"
		want := "test"
		got := cutLongStringParams(paramVal)
		if got != want {
			t.Errorf("cutLongStringParams() = %v, want %v", got, want)
		}
	})

	t.Run("Long String", func(t *testing.T) {
		paramVal := "this is a very long string"
		want := "this is a ...(length=24)"
		got := cutLongStringParams(paramVal)
		if s, ok := got.(string); ok {
			if !strings.Contains(s, "...(length=") {
				t.Errorf("cutLongStringParams() = %v, want %v", got, want)
			}
		} else {
			t.Errorf("cutLongStringParams() = %v, want %v", got, want)
		}
	})

	t.Run("Non-String Type", func(t *testing.T) {
		paramVal := 123
		want := 123
		got := cutLongStringParams(paramVal)
		if got != want {
			t.Errorf("cutLongStringParams() = %v, want %v", got, want)
		}
	})

	t.Run("Stringer Interface", func(t *testing.T) {
		paramVal := testStringer{"this is a very long string"}
		want := "this is a ...(length=24)"
		got := cutLongStringParams(paramVal)
		if s, ok := got.(string); ok {
			if !strings.Contains(s, "...(length=") {
				t.Errorf("cutLongStringParams() = %v, want %v", got, want)
			}
		} else {
			t.Errorf("cutLongStringParams() = %v, want %v", got, want)
		}
	})
}

// 用于测试 Stringer 接口的辅助类型
type testStringer struct {
	value string
}

func (ts testStringer) String() string {
	return ts.value
}
