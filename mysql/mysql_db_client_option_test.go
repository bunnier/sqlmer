package mysql_test

import (
	"testing"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/mysql"
)

func Test_WithSqlParserCacheCapacity(t *testing.T) {
	_, err := sqlmer.NewDbClientConfig(mysql.WithSqlParserCacheCapacity(8))
	if err != nil {
		t.Fatal(err)
	}
}

func Test_WithSqlParserCacheCapacity_invalid_capacity(t *testing.T) {
	_, err := sqlmer.NewDbClientConfig(mysql.WithSqlParserCacheCapacity(0))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
