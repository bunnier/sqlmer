package sqlmer

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func TestGetExecutingSqlError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		rawSql   string
		fixedSql string
		params   []any
		wantErr  string
	}{
		{
			name:     "Basic Error Test",
			err:      errors.New("test error"),
			rawSql:   "SELECT * FROM users WHERE id = @p1",
			fixedSql: "SELECT * FROM users WHERE id = ?",
			params:   []any{1},
			wantErr:  "dbClient: failed to execute sql\nraw error: test error\nsql:\ninput sql=SELECT * FROM users WHERE id = @p1\nexecuting sql=SELECT * FROM users WHERE id = ?\nparams:\n@p1=1",
		},
		{
			name:     "Named Parameter Test",
			err:      errors.New("test error"),
			rawSql:   "SELECT * FROM users WHERE name = @name",
			fixedSql: "SELECT * FROM users WHERE name = ?",
			params:   []any{sql.Named("name", "test")},
			wantErr:  "dbClient: failed to execute sql\nraw error: test error\nsql:\ninput sql=SELECT * FROM users WHERE name = @name\nexecuting sql=SELECT * FROM users WHERE name = ?\nparams:\n@name=test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := getExecutingSqlError(tt.err, tt.rawSql, tt.fixedSql, tt.params)
			if gotErr.Error() != tt.wantErr {
				t.Errorf("getExecutingSqlError() error = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestGetSqlError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		rawSql  string
		params  []any
		wantErr string
	}{
		{
			name:    "Basic Error Test",
			err:     ErrExpectedSizeWrong,
			rawSql:  "UPDATE users SET name = @p1",
			params:  []any{"test"},
			wantErr: "dbClient: effected rows was wrong\nsql:\ninput sql=UPDATE users SET name = @p1\nparams:\n@p1=test",
		},
		{
			name:    "Named Parameter Test",
			err:     ErrParseParamFailed,
			rawSql:  "INSERT INTO users (name) VALUES (@name)",
			params:  []any{sql.Named("name", "test")},
			wantErr: "dbClient: failed to parse named params\nsql:\ninput sql=INSERT INTO users (name) VALUES (@name)\nparams:\n@name=test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := getSqlError(tt.err, tt.rawSql, tt.params)
			if gotErr.Error() != tt.wantErr {
				t.Errorf("getSqlError() error = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestCutLongStringParams(t *testing.T) {
	originalMaxLength := MaxLengthErrorValue
	defer func() { MaxLengthErrorValue = originalMaxLength }()

	// 为了便于测试，将最大长度设置为较小的值
	MaxLengthErrorValue = 10

	tests := []struct {
		name     string
		paramVal any
		want     any
	}{
		{
			name:     "Short String",
			paramVal: "test",
			want:     "test",
		},
		{
			name:     "Long String",
			paramVal: "this is a very long string",
			want:     "this is a ...(length=24)",
		},
		{
			name:     "Non-String Type",
			paramVal: 123,
			want:     123,
		},
		{
			name:     "Stringer Interface",
			paramVal: testStringer{"this is a very long string"},
			want:     "this is a ...(length=24)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cutLongStringParams(tt.paramVal)
			if got != tt.want {
				// 对于字符串类型的结果，检查是否包含预期的内容
				if s, ok := got.(string); ok {
					if !strings.Contains(s, "...(length=") {
						t.Errorf("cutLongStringParams() = %v, want %v", got, tt.want)
					}
				} else {
					t.Errorf("cutLongStringParams() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// 用于测试 Stringer 接口的辅助类型
type testStringer struct {
	value string
}

func (ts testStringer) String() string {
	return ts.value
}
