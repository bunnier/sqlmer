package mssql

import (
	"testing"

	"errors"

	"github.com/bunnier/sqlmer"
)

func Test_MssqlTransaction(t *testing.T) {
	mssqlClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = mssqlClient.Execute(
		`
INSERT INTO go_TypeTest (TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest, BinaryTest)
	VALUES (30, 30, 30, 30, N'行30', 'Row30', N'行30', 'Row30', '2021-07-30 15:38:39.583', '2021-07-30 15:38:50.4257813', '2021-07-30', '12:30:01.345', 30.123, 30.12345, 30.45678999, 1);
INSERT INTO go_TypeTest (TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest, BinaryTest)
	VALUES (31, 31, 31, 31, N'行31', 'Row31', N'行31', 'Row31', '2021-07-31 15:38:39.583', '2021-07-31 15:38:50.4257813', '2021-07-31', '12:31:01.345', 31.123, 31.12345, 31.45678999, 1);`); err != nil {
		t.Fatal(err)
	}
	TransactionFuncTest(t, mssqlClient)
	// 清理数据
	if _, err = mssqlClient.Execute("DELETE FROM go_TypeTest WHERE TinyIntTest IN(30, 31)"); err != nil {
		t.Errorf("clean error: %v", err)
	}
}

func TransactionFuncTest(t *testing.T, dbClient sqlmer.DbClient) {
	// 测试事务回滚。
	t.Run("rollback", func(t *testing.T) {
		TransactionRollback(t, dbClient)
	})

	// 测试事务提交。
	t.Run("commit", func(t *testing.T) {
		TransactionCommit(t, dbClient)
	})

	// 测试事务提交。
	t.Run("commitAfterRollback", func(t *testing.T) {
		TransactionCommitAfterRollback(t, dbClient)
	})

	// 测试事务提交。
	t.Run("rollbackAfterCommit", func(t *testing.T) {
		TransactionRollbackAfterCommit(t, dbClient)
	})

	// 测试嵌套事务提交。
	t.Run("embeddedCommit", func(t *testing.T) {
		tx, err := dbClient.CreateTransaction()
		if err != nil {
			t.Errorf("dbClient.CreateTransaction() error = %v, wantErr nil", err)
			return
		}
		defer tx.MustClose()

		TransactionEmbeddedCommit(t, 3, 0, tx)
		tx.Commit()
	})

	// 测试嵌套事务回滚。
	t.Run("embeddedRollback", func(t *testing.T) {
		tx, err := dbClient.CreateTransaction()
		if err != nil {
			t.Errorf("dbClient.CreateTransaction() error = %v, wantErr nil", err)
			return
		}
		defer tx.MustClose()

		TransactionEmbeddedRollback(t, 3, 0, tx)
		tx.Commit()
	})
}

func TransactionRollback(t *testing.T, dbClient sqlmer.DbClient) {
	tx, err := dbClient.CreateTransaction()
	if err != nil {
		t.Errorf("dbClient.CreateTransaction() error = %v, wantErr nil", err)
		return
	}
	defer tx.MustClose()

	res, err := tx.Execute("DELETE FROM go_TypeTest WHERE TinyIntTest=30") // Mysql & Sqlserver都有这个表，这条记录其它测试用例都没用到。
	if err != nil {
		t.Errorf("transactionKeeper.Execute() error = %v, wantErr nil", err)
		return
	}
	if res != 1 {
		t.Errorf("transactionKeeper.Execute() effectRows = %v, wantRows 1", res)
		return
	}
	if err = tx.Rollback(); err != nil { // 回滚。
		t.Errorf("transactionKeeper.Rollback() error = %v, wantErr nil", err)
		return
	}

	if exist, err := dbClient.Exists("SELECT 1 FROM go_TypeTest WHERE TinyIntTest=30"); err != nil { // 记录需要依然存在。
		t.Errorf("transactionKeeper.Exists() error = %v, wantErr nil", err)
		return
	} else if !exist {
		t.Errorf("transactionKeeper.Exists() not exist, want exist")
		return
	}
}

func TransactionCommit(t *testing.T, dbClient sqlmer.DbClient) {
	tx, err := dbClient.CreateTransaction()
	if err != nil {
		t.Errorf("dbClient.CreateTransaction() error = %v, wantErr nil", err)
		return
	}
	defer tx.MustClose()

	res, err := tx.Execute("DELETE FROM go_TypeTest WHERE TinyIntTest=30")
	if err != nil {
		t.Errorf("transactionKeeper.Execute() error = %v, wantErr nil", err)
		return
	}
	if res != 1 {
		t.Errorf("transactionKeeper.Execute() effectRows = %v, wantRows %d", res, 1)
		return
	}
	if err = tx.Commit(); err != nil { // 提交。
		t.Errorf("transactionKeeper.Commit() error = %v, wantErr nil", err)
		return
	}

	if exist, err := dbClient.Exists("SELECT 1 FROM go_TypeTest WHERE TinyIntTest=30"); err != nil { // 记录需要依然存在。
		t.Errorf("transactionKeeper.Exists() error = %v, wantErr nil", err)
		return
	} else if exist {
		t.Errorf("transactionKeeper.Exists() exist, want not exist")
		return
	}
}

func TransactionCommitAfterRollback(t *testing.T, dbClient sqlmer.DbClient) {
	tx, err := dbClient.CreateTransaction()
	if err != nil {
		t.Errorf("dbClient.CreateTransaction() error = %v, wantErr nil", err)
		return
	}
	defer tx.MustClose()

	res, err := tx.Execute("DELETE FROM go_TypeTest WHERE TinyIntTest=31")
	if err != nil {
		t.Errorf("transactionKeeper.Execute() error = %v, wantErr nil", err)
		return
	}
	if res != 1 {
		t.Errorf("transactionKeeper.Execute() effectRows = %v, wantRows %d", res, 1)
		return
	}
	if err = tx.Rollback(); err != nil { // 回滚。
		t.Errorf("transactionKeeper.Rollback() error = %v, wantErr nil", err)
		return
	}
	if err = tx.Commit(); err == nil { // 回滚后提交。
		t.Errorf("transactionKeeper.Commit() error is nil, wantErr NewDbTransError")

		if !errors.Is(err, sqlmer.ErrTran) {
			t.Errorf("internalDbClient.Commit() error = %v, wantErr DbTransError", err)
		}
	}

	if exist, err := dbClient.Exists("SELECT 1 FROM go_TypeTest WHERE TinyIntTest=31"); err != nil { // 记录需要依然存在。
		t.Errorf("transactionKeeper.Exists() error = %v, wantErr nil", err)
		return
	} else if !exist {
		t.Errorf("transactionKeeper.Exists() exist, want exist")
		return
	}
}

func TransactionRollbackAfterCommit(t *testing.T, dbClient sqlmer.DbClient) {
	tx, err := dbClient.CreateTransaction()
	if err != nil {
		t.Errorf("dbClient.CreateTransaction() error = %v, wantErr nil", err)
		return
	}
	defer tx.MustClose()

	res, err := tx.Execute("DELETE FROM go_TypeTest WHERE TinyIntTest=31")
	if err != nil {
		t.Errorf("transactionKeeper.Execute() error = %v, wantErr nil", err)
		return
	}
	if res != 1 {
		t.Errorf("transactionKeeper.Execute() effectRows = %v, wantRows 1", res)
		return
	}
	if err = tx.Commit(); err != nil { // 提交。
		t.Errorf("transactionKeeper.Commit() error = %v, wantErr nil", err)
		return
	}
	if err = tx.Rollback(); err == nil { // 提交后回滚。
		t.Errorf("transactionKeeper.Rollback() error is nil, wantErr NewDbTransError")

		if !errors.Is(err, sqlmer.ErrTran) {
			t.Errorf("internalDbClient.Rollback() error = %v, wantErr DbTransError", err)
		}
	}

	if exist, err := dbClient.Exists("SELECT 1 FROM go_TypeTest WHERE TinyIntTest=31"); err != nil { // 记录需要依然存在。
		t.Errorf("transactionKeeper.Exists() error = %v, wantErr nil", err)
		return
	} else if exist {
		t.Errorf("transactionKeeper.Exists() exist, want not exist")
		return
	}
}

func TransactionEmbeddedCommit(t *testing.T, maxDepth int, currentDepth int, tx sqlmer.TransactionKeeper) {
	tx, err := tx.CreateTransaction()
	if err != nil {
		t.Errorf("dbClient.CreateTransaction() Embeddedly error = %v, wantErr nil", err)
		return
	}
	defer tx.MustClose()

	// 嵌套下去。
	if maxDepth > currentDepth {
		currentDepth++
		TransactionEmbeddedCommit(t, 3, currentDepth, tx)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("dbClient.Commit() Embeddedly error = %v, wantErr nil", err)
		return
	}
}

func TransactionEmbeddedRollback(t *testing.T, maxDepth int, currentDepth int, tx sqlmer.TransactionKeeper) {
	tx, err := tx.CreateTransaction()
	if err != nil {
		t.Errorf("dbClient.CreateTransaction() Embeddedly error = %v, wantErr nil", err)
		return
	}
	defer tx.MustClose()

	if maxDepth > currentDepth {
		currentDepth++
		TransactionEmbeddedRollback(t, 3, currentDepth, tx)
	}

	if err := tx.Rollback(); err != nil {
		t.Errorf("dbClient.Rollback() Embeddedly error = %v, wantErr nil", err)
		return
	}
}
