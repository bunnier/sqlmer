package sqlmer

import (
	"database/sql"

	"github.com/pkg/errors"
)

var _ DbClient = (*sqlxDbTransactionKeeper)(nil)
var _ TransactionKeeper = (*sqlxDbTransactionKeeper)(nil)

// sqlxDbClient 是通过sqlx实现的TransactionKeeper结构。
type sqlxDbTransactionKeeper struct {
	*internalDbClient
	Tx *sql.Tx

	// 当前事务是否已经完结，若完结则不允许再执行数据库操作。
	transactionCompleted bool

	// 事务的嵌套层级。TransactionKeeper 接口继承了 DbClient，所以具有 CreateTransaction 方法。
	// 刚创建的事务嵌套层级为0，事务内再次创建事务时+1，并返回（复用）当前实例。
	embeddedLevel int
}

// Commit 用于提交事务。
func (transKeeper *sqlxDbTransactionKeeper) Commit() error {
	if transKeeper.embeddedLevel > 0 {
		// 说明是嵌套事务，直接返回交给上层提交。
		return nil
	}

	if transKeeper.transactionCompleted {
		return errors.WithMessage(ErrTran, "trans has already completed")
	}

	transKeeper.transactionCompleted = true
	return transKeeper.Tx.Commit()
}

// Rollback 用于回滚事务。
func (transKeeper *sqlxDbTransactionKeeper) Rollback() error {
	if transKeeper.embeddedLevel > 0 {
		// 说明是嵌套事务，直接返回交给上层回滚。
		return nil
	}

	if transKeeper.transactionCompleted {
		return errors.WithMessage(ErrTran, "trans has already completed")
	}

	transKeeper.transactionCompleted = true
	return transKeeper.Tx.Rollback()
}

// Close 用于优雅关闭事务，创建事务后应defer执行本方法。
func (transKeeper *sqlxDbTransactionKeeper) Close() error {
	transKeeper.embeddedLevel--
	if transKeeper.embeddedLevel != 0 || transKeeper.transactionCompleted {
		return nil
	}
	return transKeeper.Rollback()
}

// CreateTransaction 用于开始一个事务。
func (transKeeper *sqlxDbTransactionKeeper) CreateTransaction() (TransactionKeeper, error) {
	transKeeper.embeddedLevel++
	return transKeeper, nil
}
