package sqlmer

import (
	"database/sql"
	"fmt"
)

var _ DbClient = (*abstractTransactionKeeper)(nil)
var _ TransactionKeeper = (*abstractTransactionKeeper)(nil)
var _ ErrorTransactionKeeper = (*abstractTransactionKeeper)(nil)
var _ MustTransactionKeeper = (*abstractTransactionKeeper)(nil)

// abstractTransactionKeeper 是通过 TransactionKeeper 结构。
type abstractTransactionKeeper struct {
	*AbstractDbClient
	Tx *sql.Tx

	// 当前事务是否已经完结，若完结则不允许再执行数据库操作。
	transactionCompleted bool

	// 事务的嵌套层级。TransactionKeeper 接口内嵌了 DbClient，也具有 CreateTransaction 方法。
	// 刚创建的事务嵌套层级为 0，事务内再次创建事务时 +1，并返回（复用）当前实例。
	embeddedLevel int
}

// Commit 用于提交事务。
func (transKeeper *abstractTransactionKeeper) Commit() error {
	if transKeeper.embeddedLevel > 0 {
		// 说明是嵌套事务，直接返回交给上层提交。
		return nil
	}

	if transKeeper.transactionCompleted {
		return fmt.Errorf("%w: trans has already completed", ErrTran)
	}

	transKeeper.transactionCompleted = true
	return transKeeper.Tx.Commit()
}

// Rollback 用于回滚事务。
func (transKeeper *abstractTransactionKeeper) Rollback() error {
	if transKeeper.embeddedLevel > 0 {
		// 说明是嵌套事务，直接返回交给上层回滚。
		return nil
	}

	if transKeeper.transactionCompleted {
		return fmt.Errorf("%w: trans has already completed", ErrTran)
	}

	transKeeper.transactionCompleted = true
	return transKeeper.Tx.Rollback()
}

// Close 用于优雅关闭事务，创建事务后可 defer 执行本方法。
func (transKeeper *abstractTransactionKeeper) Close() error {
	transKeeper.embeddedLevel--
	if transKeeper.embeddedLevel != -1 || transKeeper.transactionCompleted {
		return nil
	}
	return transKeeper.Rollback()
}

// CreateTransaction 用于开始一个事务。
func (transKeeper *abstractTransactionKeeper) CreateTransaction() (TransactionKeeper, error) {
	transKeeper.embeddedLevel++
	return transKeeper, nil
}

// MustCommit 用于提交事务。
func (transKeeper *abstractTransactionKeeper) MustCommit() {
	if err := transKeeper.Commit(); err != nil {
		panic(err)
	}
}

// MustRollback 用于回滚事务。
func (transKeeper *abstractTransactionKeeper) MustRollback() {
	if err := transKeeper.Rollback(); err != nil {
		panic(err)
	}
}

// MustClose 用于优雅关闭事务，创建事务后可 defer 执行本方法。
func (transKeeper *abstractTransactionKeeper) MustClose() {
	if err := transKeeper.Close(); err != nil {
		panic(err)
	}
}

// MustCreateTransaction 用于开始一个事务。
func (transKeeper *abstractTransactionKeeper) MustCreateTransaction() TransactionKeeper {
	if trans, err := transKeeper.CreateTransaction(); err != nil {
		panic(err)
	} else {
		return trans
	}
}
