package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/mysql"
)

func main() {
	var dbClient sqlmer.DbClient // 这是本库的主接口，统一了各种数据库的 API 操作。
	var err error                // 本库同时提供了 error/panic 两套 API，为了 demo 更为简洁，后续主要通过 panic(Must) 版本 API 演示。

	// 这里使用 MySQL 做示范，SQL Server 也提供了一致的 API 和相应的参数解析逻辑。
	if dbClient, err = mysql.NewMySqlDbClient(
		"test:test@tcp(127.0.0.1:1433)/test",
		sqlmer.WithConnTimeout(time.Second*30), // 连接超时。
		sqlmer.WithExecTimeout(time.Second*30), // 读写超时(执行语句时候，如果没有指定超时时间，默认用这个)。
	); err != nil {
		log.Fatal(err)
	}

	// 创建/删除 测试表。
	dbClient.MustExecute(`
		CREATE TABLE demo(
			Id int(11) NOT NULL AUTO_INCREMENT,
			Name VARCHAR(10) NOT NULL,
			Age INT NOT NULL,
			PRIMARY KEY (Id),
			KEY demo (Id))`)
	defer dbClient.MustExecute("DROP TABLE demo")

	// 通过 context 设置超时时间。。
	ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
	if _, err = dbClient.ExecuteContext(ctx, "SELECT sleep(3)"); err != nil {
		fmt.Println("timeout: " + err.Error()) // 预期内的超时~
	}

	// 索引方式插入数据，@p1..@pn，分别对应第 1..n 个参数。
	dbClient.MustExecute("INSERT INTO demo(Name, Age) VALUES(@p1, @p2)", "rui", 1)
	dbClient.MustExecute("INSERT INTO demo(Name, Age) VALUES(@p1, @p2)", "bao", 2)

	// 命名参数查询数据，命名参数采用 map，key 为 sql 语句 @ 之后的参数名，value 为值。
	dataMap := dbClient.MustGet("SELECT * FROM demo WHERE Name=@name", map[string]any{"name": "rui"})
	fmt.Println(dataMap) // Output: map[Age:1 Id:1 Name:rui]

	// 获取第一行第一列，DBNull 和 未命中都会返回 nil，因此提供了第二返回值 hit（bool 类型）来区分是 DBNull 和无数据，这里不是可空字段因此无需判断。
	name, _ := dbClient.MustScalar("SELECT Name FROM demo WHERE Name=@p1", "rui")
	fmt.Println(name.(string)) // Output: rui

	// 如果喜欢标准库风格，这里也提供了增强版本的 sql.Rows，支持 SliceScan、MapScan。
	rows := dbClient.MustRows("SELECT Name, now() FROM demo WHERE Name IN (@p1, @p2)", "rui", "bao")
	for rows.Next() {
		// SliceScan 会自动判断列数及列类型，用 []any 方式返回。
		if dataSlice, err := rows.SliceScan(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(dataSlice...)
			// Output:
			// rui 2022-04-09 22:35:33 +0000 UTC
			// bao 2022-04-09 22:35:33 +0000 UTC
		}
	}
	// 和标准库一样，Rows 的 Err 和 Close 返回的错误记得要处理哦～
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	if err = rows.Close(); err != nil {
		log.Fatal(err)
	}

	rowNum, _ := dbClient.MustScalar("SELECT count(1) FROM demo")
	fmt.Println(rowNum) // Output: 2

	trans := dbClient.MustCreateTransaction() // 事务操作也支持和 DbClient 几乎一致的 API。
	trans.MustExecute("DELETE FROM demo WHERE Id=1")

	embeddedTrans := trans.MustCreateTransaction() // 支持嵌套事务。
	embeddedTrans.MustExecute("DELETE FROM demo WHERE Id=2")
	embeddedTrans.MustCommit()
	embeddedTrans.MustClose() // 注意：嵌套事务也需要 Close。

	trans.MustCommit()
	trans.MustClose()

	rowNum, _ = dbClient.MustScalar("SELECT count(1) FROM demo")
	fmt.Println(rowNum) // Output: 0
}
