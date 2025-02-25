package sqlmer

var _ TransactionKeeper = (*TransactionKeeperEx)(nil)

// TransactionKeeperEx 扩展 TransactionKeeper ，增加 DbClientEx 的功能和 Must 版本的事务 API。
type TransactionKeeperEx struct {
	*DbClientEx
	TransactionKeeper
}

// MustCreateTransaction 用于开始一个事务。
func (transKeeper *TransactionKeeperEx) MustCreateTransaction() TransactionKeeper {
	if trans, err := transKeeper.CreateTransaction(); err != nil {
		panic(err)
	} else {
		return trans
	}
}

// CreateTransactionEx 基于 DbClient.CreateTransaction 创建一个 TransactionKeeperEx 实例。
func (c *DbClientEx) CreateTransactionEx() (tran *TransactionKeeperEx, err error) {
	t, err := c.DbClient.CreateTransaction()
	if err != nil {
		return nil, err
	}

	return &TransactionKeeperEx{
		DbClientEx:        Extend(t),
		TransactionKeeper: t,
	}, nil
}

// MustCreateTransactionEx 基于 DbClient.MustCreateTransaction 创建一个 TransactionKeeperEx 实例。
func (c *DbClientEx) MustCreateTransactionEx() (tran *TransactionKeeperEx) {
	t := c.MustCreateTransaction()
	return &TransactionKeeperEx{
		DbClientEx:        Extend(t),
		TransactionKeeper: t,
	}
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
