package mysql_test

import (
	"testing"

	"errors"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/testenv"
)

func Test_MysqlTransaction(t *testing.T) {
	mysqlClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatal(err)
	}

	// 准备数据。
	if _, err = mysqlClient.Execute(`INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest)
	VALUES (31, 31, 31, 31, 31, N'行31', '行31char', '行31text','2021-07-31','2021-07-31 15:38:50.425','2021-07-31 15:38:50.425', 31.456, 31.15678, 31.45678999, 0);`); err != nil {
		t.Fatalf("prepare error: %v", err)
	}
	if _, err = mysqlClient.Execute(`INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest)
	VALUES (30, 30, 30, 30, 30, N'行30', '行30char', '行30text','2021-07-30','2021-07-30 15:38:50.425','2021-07-30 15:38:50.425', 30.456, 30.15678, 30.45678999, 0);
	`); err != nil {
		t.Fatalf("prepare error: %v", err)
	}
	TransactionFuncTest(t, mysqlClient)
	// 清理数据
	if _, err = mysqlClient.Execute("DELETE FROM go_TypeTest WHERE VarcharTest IN(N'行30', N'行31')"); err != nil {
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
		if err := tx.Commit(); err != nil {
			t.Errorf("dbClient.Commit() Embeddedly error = %v, wantErr nil", err)
			return
		}
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
		if err := tx.Commit(); err != nil {
			t.Errorf("dbClient.Commit() Embeddedly error = %v, wantErr nil", err)
			return
		}
	})
}

func TransactionRollback(t *testing.T, dbClient sqlmer.DbClient) {
	tx, err := dbClient.CreateTransaction()
	if err != nil {
		t.Errorf("dbClient.CreateTransaction() error = %v, wantErr nil", err)
		return
	}
	defer tx.MustClose()

	res, err := tx.Execute("DELETE FROM go_TypeTest WHERE VarcharTest='行30'") // Mysql & Sqlserver都有这个表，这条记录其它测试用例都没用到。
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

	if exist, err := dbClient.Exists("SELECT 1 FROM go_TypeTest WHERE VarcharTest='行30'"); err != nil { // 记录需要依然存在。
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

	res, err := tx.Execute("DELETE FROM go_TypeTest WHERE VarcharTest='行30'")
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

	if exist, err := dbClient.Exists("SELECT 1 FROM go_TypeTest WHERE VarcharTest='行30'"); err != nil { // 记录需要依然存在。
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

	res, err := tx.Execute("DELETE FROM go_TypeTest WHERE VarcharTest='行31'")
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

	if exist, err := dbClient.Exists("SELECT 1 FROM go_TypeTest WHERE VarcharTest='行31'"); err != nil { // 记录需要依然存在。
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

	res, err := tx.Execute("DELETE FROM go_TypeTest WHERE VarcharTest='行31'")
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

	if exist, err := dbClient.Exists("SELECT 1 FROM go_TypeTest WHERE VarcharTest='行31'"); err != nil { // 记录需要依然存在。
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
		t.Errorf("dbClient.CreateTransaction() Embeddeddly error = %v, wantErr nil", err)
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
