package mysql_test

import (
	"testing"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/mysql"
)

func Test_WithQuestionMarkNamedSqlCacheCapacity(t *testing.T) {
	_, err := sqlmer.NewDbClientConfig(mysql.WithQuestionMarkNamedSqlCacheCapacity(8))
	if err != nil {
		t.Fatal(err)
	}
}

func Test_WithQuestionMarkNamedSqlCacheCapacity_invalid_capacity(t *testing.T) {
	_, err := sqlmer.NewDbClientConfig(mysql.WithQuestionMarkNamedSqlCacheCapacity(0))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
