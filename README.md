# sqlmer

[![License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![Go](https://github.com/bunnier/sqlmer/actions/workflows/go.yml/badge.svg)](https://github.com/bunnier/sqlmer/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bunnier/sqlmer)](https://goreportcard.com/report/github.com/bunnier/sqlmer)
[![Go Reference](https://pkg.go.dev/badge/github.com/bunnier/sqlmer.svg)](https://pkg.go.dev/github.com/bunnier/sqlmer)

## 功能简介

数据库访问库，目前支持 MySQL 和 SQL Server。

- SQL语句提供了统一的 `命名参数`、`索引参数` 语法，可直观的使用 map 作为 SQL 语句参数，并以 map 或 slice 方式返回；
- 提供了面向map的 `参数`、`结果集` 交互接口，事务和非事务访问均可通过相同接口完成；
- 扩展了原生 sql.Rows / sql.Row，使其支持 MapScan 以及 SliceScan；

## 简单样例

```go
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
		"test:test@tcp(127.0.0.1:1433)/test?parseTime=true",
		sqlmer.WithConnTimeout(time.Second*30), // 连接超时。
		sqlmer.WithExecTimeout(time.Second*30), // 读写超时(执行语句时候，如果没有指定超时时间，默认用这个)。
	); err != nil {
		fmt.Println(err)
		return
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

	// 获取增强后的 sql.Rows（支持 SliceScan、MapScan）。
	sliceRows := dbClient.MustRows("SELECT Name, now() FROM demo WHERE Name IN (@p1, @p2)", "rui", "bao")
	for sliceRows.Next() {
		// SliceScan 会自动判断列数及列类型，用 []any 方式返回。
		if dataSlice, err := sliceRows.SliceScan(); err != nil {
			log.Fatal(err)
			return
		} else {
			fmt.Println(dataSlice...)
			// Output:
			// rui 2022-04-09 22:35:33 +0000 UTC
			// bao 2022-04-09 22:35:33 +0000 UTC
		}
	}
	sliceRows.Close()

	if sliceRows.Err() != nil {
		log.Fatal(err)
		return
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
```

## 类型映射

> nullable 的字段，如果值为 null，默认均以 nil 返回。

### MySql

| DB datatype                                        | Go datatype |
|----------------------------------------------------|-------------|
| varchar / char / text                              | string      |
| tiny int / small int / int / unsigned int / bigint | int64       |
| float / double                                     | float64     |
| decimal                                            | string      |
| date / datetime / timestamp                        | time.Time   |
| bit                                                | []byte      |

### SQL Server

| DB datatype                              | Go datatype |
|------------------------------------------|-------------|
| nvarchar / nchar / varchar / char / text | string      |
| small int / tiny int / int / bigint      | int64       |
| float / real                             | float64     |
| small money / money / decimal            | string      |
| date / datetime / datetime2 / time       | time.Time   |
| binary / varbinary                       | []byte      |
| bit                                      | bool        |

## 测试用例

测试用例 Schema：

1. 编辑 `test_conf.yml` 文件，相应数据库的连接字符串；
2. 如果第 1 步配置的连接字符串有 DDL 权限，可通过调用 `go run ./internal/testcmd/main.go -a PREPARE -c test_conf.yml` 来同时准备 MySQL / SQL Server 环境，如果没有 DDL 权限可自行直接执行 `internal/testenv/*_prepare.sql` 准备环境；
3. 如果第 1 步配置的连接字符串有 DDL 权限，测试结束后可以通过 `go run ./internal/testcmd/main.go -a CLEAN -c test_conf.yml` 销毁测试表，如果没有 DDL 权限可自行直接执行 `internal/testenv/*_clean.sql` 销毁测试表；

另外，如果你和我一样使用 VSCode 作为开发工具，可在配置好 `test_conf.yml` 之后，直接使用 .vscode 中编写好的 Task 来准备环境。
