package sqlmer_test

import (
	"testing"

	"github.com/bunnier/sqlmer"
	"github.com/pkg/errors"
)

func Test_MysqlTransaction(t *testing.T) {
	mysqlClient, err := getMySqlDbClient()
	if err != nil {
		t.Fatal(err)
	}
	// 准备数据。
	if _, err = mysqlClient.Execute("INSERT INTO go_TypeTest(VarcharTest, dateTest, dateTimeTest, timestampTest, decimalTest) VALUES (N'Row100', '2021-07-03', '2021-07-03 15:38:50.425','2021-07-03 15:38:50.425', 3.45678999)"); err != nil {
		t.Fatalf("prepare error: %v", err)
	}
	if _, err = mysqlClient.Execute("INSERT INTO go_TypeTest(VarcharTest, dateTest, dateTimeTest, timestampTest, decimalTest) VALUES (N'Row101', '2021-07-03', '2021-07-03 15:38:50.425','2021-07-03 15:38:50.425', 3.45678999)"); err != nil {
		t.Fatalf("prepare error: %v", err)
	}
	TransactionFuncTest(t, mysqlClient)
	// 清理数据
	if _, err = mysqlClient.Execute("DELETE FROM go_TypeTest WHERE VarcharTest IN(N'Row101', N'Row100')"); err != nil {
		t.Errorf("clean error: %v", err)
	}
}

func Test_MssqlTransaction(t *testing.T) {
	mssqlClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = mssqlClient.Execute("INSERT INTO go_TypeTest (NvarcharTest, VarcharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, DecimalTest) VALUES (N'行100', 'Row100', '2021-07-04 15:38:39.583', '2021-07-04 15:38:50.4257813', '2021-07-04', '12:01:04.345', 4.45678999);"); err != nil {
		t.Fatal(err)
	}
	if _, err = mssqlClient.Execute("INSERT INTO go_TypeTest (NvarcharTest, VarcharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, DecimalTest) VALUES (N'行101', 'Row101', '2021-07-04 15:38:39.583', '2021-07-04 15:38:50.4257813', '2021-07-04', '12:01:04.345', 4.45678999);"); err != nil {
		t.Fatal(err)
	}
	TransactionFuncTest(t, mssqlClient)
	// 清理数据
	if _, err = mssqlClient.Execute("DELETE FROM go_TypeTest WHERE VarcharTest IN('Row101', 'Row100')"); err != nil {
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
		defer tx.Close()
		TransactionEmbeddedCommit(t, 3, 0, tx)
	})

	// 测试嵌套事务回滚。
	t.Run("embeddedRollback", func(t *testing.T) {
		tx, err := dbClient.CreateTransaction()
		if err != nil {
			t.Errorf("dbClient.CreateTransaction() error = %v, wantErr nil", err)
			return
		}
		defer tx.Close()
		TransactionEmbeddedRollback(t, 3, 0, tx)
	})
}

func TransactionRollback(t *testing.T, dbClient sqlmer.DbClient) {
	tx, err := dbClient.CreateTransaction()
	if err != nil {
		t.Errorf("dbClient.CreateTransaction() error = %v, wantErr nil", err)
		return
	}
	defer tx.Close()
	res, err := tx.Execute("DELETE FROM go_TypeTest WHERE VarcharTest='Row100'") // Mysql & Sqlserver都有这个表，这条记录其它测试用例都没用到。
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

	if exist, err := dbClient.Exists("SELECT 1 FROM go_TypeTest WHERE VarcharTest='Row100'"); err != nil { // 记录需要依然存在。
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
	defer tx.Close()
	res, err := tx.Execute("DELETE FROM go_TypeTest WHERE VarcharTest='Row100'")
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

	if exist, err := dbClient.Exists("SELECT 1 FROM go_TypeTest WHERE VarcharTest='Row100'"); err != nil { // 记录需要依然存在。
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
	defer tx.Close()
	res, err := tx.Execute("DELETE FROM go_TypeTest WHERE VarcharTest='Row101'")
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

	if exist, err := dbClient.Exists("SELECT 1 FROM go_TypeTest WHERE VarcharTest='Row101'"); err != nil { // 记录需要依然存在。
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
	defer tx.Close()
	res, err := tx.Execute("DELETE FROM go_TypeTest WHERE VarcharTest='Row101'")
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

	if exist, err := dbClient.Exists("SELECT 1 FROM go_TypeTest WHERE VarcharTest='Row101'"); err != nil { // 记录需要依然存在。
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
		t.Errorf("dbClient.CreateTransaction() Embeddeddly error = %v, wantErr nil", err)
		return
	}

	// 嵌套下去。
	if maxDepth > currentDepth {
		currentDepth++
		TransactionEmbeddedCommit(t, 3, currentDepth, tx)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("dbClient.Commit() Embeddeddly error = %v, wantErr nil", err)
		return
	}
	defer tx.Close()
}

func TransactionEmbeddedRollback(t *testing.T, maxDepth int, currentDepth int, tx sqlmer.TransactionKeeper) {
	tx, err := tx.CreateTransaction()
	if err != nil {
		t.Errorf("dbClient.CreateTransaction() Embeddeddly error = %v, wantErr nil", err)
		return
	}

	if maxDepth > currentDepth {
		currentDepth++
		TransactionEmbeddedRollback(t, 3, currentDepth, tx)
	}

	if err := tx.Rollback(); err != nil {
		t.Errorf("dbClient.Rollback() Embeddeddly error = %v, wantErr nil", err)
		return
	}

	defer tx.Close()
}
