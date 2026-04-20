package mysql

import (
	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/named2qm"
)

// WithSqlParserCacheCapacity 用于为 MySQL 配置命名参数 SQL 解析缓存容量。
func WithSqlParserCacheCapacity(cacheCapacity int) sqlmer.DbClientOption {
	return func(config *sqlmer.DbClientConfig) error {
		binder, err := named2qm.NewQuestionMarkSqlBinder(cacheCapacity)
		if err != nil {
			return err
		}

		return sqlmer.WithBindArgsFunc(binder.BindQuestionMarkArgs)(config)
	}
}
