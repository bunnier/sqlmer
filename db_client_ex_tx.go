package sqlmer

var _ TransactionKeeper = (*TransactionKeeperEx)(nil)

// TransactionKeeperEx 扩展 TransactionKeeper ，增加 DbClientEx 的功能和 Must 版本的事务 API。
type TransactionKeeperEx struct {
	*DbClientEx
	TransactionKeeper
}

// ExtendTx 加强 TransactionKeeper 。
//   - 提供 Must 版本的 API；
func ExtendTx(raw TransactionKeeper) *TransactionKeeperEx {
	return &TransactionKeeperEx{
		DbClientEx:        Extend(raw),
		TransactionKeeper: raw,
	}
}

// CreateTransactionEx 基于 DbClient.CreateTransaction 创建一个 TransactionKeeperEx 实例。
func (c *DbClientEx) CreateTransactionEx() (tran *TransactionKeeperEx, err error) {
	if tx, err := c.DbClient.CreateTransaction(); err != nil {
		return nil, err
	} else {
		return ExtendTx(tx), nil
	}
}

// MustCreateTransactionEx（和 MustCreateTransaction 一致） 用于开始一个事务。
// returns:
//
//	@tran 返回一个TransactionKeeperEx 实例（实现了 TransactionKeeper、DbClient 接口） 接口的对象，在上面执行的语句会在同一个事务中执行。
func (c *DbClientEx) MustCreateTransactionEx() (tran *TransactionKeeperEx) {
	return c.MustCreateTransaction()
}

// MustCommit 用于提交事务。
func (transKeeper *TransactionKeeperEx) MustCommit() {
	if err := transKeeper.Commit(); err != nil {
		panic(err)
	}
}

// MustRollback 用于回滚事务。
func (transKeeper *TransactionKeeperEx) MustRollback() {
	if err := transKeeper.Rollback(); err != nil {
		panic(err)
	}
}

// MustClose 用于优雅关闭事务，创建事务后务必执行本方法或 Close 方法。
func (transKeeper *TransactionKeeperEx) MustClose() {
	if err := transKeeper.Close(); err != nil {
		panic(err)
	}
}
