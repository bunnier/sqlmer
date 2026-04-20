package sqlite

import (
	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/named2qm"
)

// WithQuestionMarkNamedSqlCacheCapacity 用于为 SQLite 配置命名参数 SQL 解析缓存容量。
func WithQuestionMarkNamedSqlCacheCapacity(cacheCapacity int) sqlmer.DbClientOption {
	return func(config *sqlmer.DbClientConfig) error {
		binder, err := named2qm.NewQuestionMarkSqlBinder(cacheCapacity)
		if err != nil {
			return err
		}

		return sqlmer.WithBindArgsFunc(binder.BindQuestionMarkArgs)(config)
	}
}
