package sqlmer

import (
	"github.com/pkg/errors"
)

// ErrConnect 是数据库连接相关错误。
var ErrConnect = errors.New("sqlmer: connect")

// ErrSql 是SQL语句相关错误。
var ErrSql = errors.New("sqlmer: sql")

// ErrTran 是数据库事务相关错误。
var ErrTran = errors.New("sqlmer: transaction")
