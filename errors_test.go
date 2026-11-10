package sqlmer

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
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

func reusableErrorParams() ([]any, func()) {
	payload := []byte("original")
	raw := sql.RawBytes("original")
	label := &testStringer{"original"}
	values := map[string]any{"value": "original"}
	params := []any{payload, sql.Named("raw", raw), label, values, "original"}
	return params, func() {
		copy(payload, "modified")
		copy(raw, "modified")
		label.value = "modified"
		values["value"] = "modified"
		params[4] = "modified"
	}
}

func TestGetExecutingSqlError_preserves_text_after_parameter_reuse(t *testing.T) {
	params, reuse := reusableErrorParams()
	underlying := errors.New("driver rejected")
	want := fmt.Sprintf("%s\nraw error: %s\nsql:\ninput sql=SELECT @p1\nexecuting sql=SELECT ?\n%s",
		ErrExecutingSql, underlying, printSqlParams(params))
	err := getExecutingSqlError(underlying, "SELECT @p1", "SELECT ?", params)
	reuse()
	if got := err.Error(); got != want {
		t.Fatalf("error text changed after parameter reuse:\ngot: %s\nwant: %s", got, want)
	}
	if !errors.Is(err, ErrExecutingSql) {
		t.Fatalf("error = %v, want ErrExecutingSql", err)
	}
}

func TestGetSqlError_preserves_text_after_parameter_reuse(t *testing.T) {
	params, reuse := reusableErrorParams()
	want := fmt.Sprintf("%s\nsql:\ninput sql=UPDATE t SET value=@p1\n%s",
		ErrExpectedSizeWrong, printSqlParams(params))
	err := getSqlError(ErrExpectedSizeWrong, "UPDATE t SET value=@p1", params)
	reuse()
	if got := err.Error(); got != want {
		t.Fatalf("error text changed after parameter reuse:\ngot: %s\nwant: %s", got, want)
	}
	if !errors.Is(err, ErrExpectedSizeWrong) {
		t.Fatalf("error = %v, want ErrExpectedSizeWrong", err)
	}
}

func TestGetExecutingSqlError_nil_time_parameter(t *testing.T) {
	var timestamp *time.Time
	err := getExecutingSqlError(errors.New("driver rejected"), "SELECT @p1", "SELECT ?", []any{timestamp})
	if !errors.Is(err, ErrExecutingSql) {
		t.Fatalf("error = %v, want ErrExecutingSql", err)
	}
	if !strings.Contains(err.Error(), "@p1=<nil>") {
		t.Fatalf("error = %v, want nil parameter text", err)
	}
}

func TestGetSqlError_nil_named_time_parameter(t *testing.T) {
	var timestamp *time.Time
	err := getSqlError(ErrExpectedSizeWrong, "UPDATE t SET value=@time", []any{sql.Named("time", timestamp)})
	if !errors.Is(err, ErrExpectedSizeWrong) {
		t.Fatalf("error = %v, want ErrExpectedSizeWrong", err)
	}
	if !strings.Contains(err.Error(), "@time=<nil>") {
		t.Fatalf("error = %v, want nil parameter text", err)
	}
}

func TestCutLongStringParams(t *testing.T) {
	originalMaxLength := MaxLengthErrorValue
	defer func() { MaxLengthErrorValue = originalMaxLength }()

	// 为了便于测试，将最大长度设置为较小的值
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
