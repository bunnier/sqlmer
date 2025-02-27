package wrap

import (
	"github.com/bunnier/sqlmer"
)

var _ sqlmer.TransactionKeeper = (*WrappedTransactionKeeper)(nil)

// WrappedTransactionKeeper 将包包裹 TransactionKeeper 的所有执行方法，以提供慢日志等能力。
type WrappedTransactionKeeper struct {
	sqlmer.DbClient                            // 内嵌 DbClient 实例。
	transactionKeeper sqlmer.TransactionKeeper // 原始的 TransactionKeeper 实例。
}

func extendExecWrapTx(tx sqlmer.TransactionKeeper, wrapFunc WrapFunc) *WrappedTransactionKeeper {
	return &WrappedTransactionKeeper{
		DbClient:          Extend(tx, wrapFunc),
		transactionKeeper: tx,
	}
}

// Commit 用于提交事务。
func (t *WrappedTransactionKeeper) Commit() error {
	return t.transactionKeeper.Close()
}

// Rollback 用于回滚事务。
func (t *WrappedTransactionKeeper) Rollback() error {
	return t.transactionKeeper.Close()
}

// Close 用于优雅关闭事务，创建事务后可 defer 执行本方法。
func (t *WrappedTransactionKeeper) Close() error {
	return t.transactionKeeper.Close()
}
