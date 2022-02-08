# sqlmer

[![Go](https://github.com/bunnier/sqlmer/actions/workflows/go.yml/badge.svg)](https://github.com/bunnier/sqlmer/actions/workflows/go.yml)

## 功能简介

数据库访问库，目前支持MySql和Sql Server。

- SQL语句提供了统一的 `命名参数`、`索引参数` 语法，可直观的使用 map 作为SQL语句参数，并以 map 或 slice 方式返回；
- 提供了面向map的 `参数`、`结果集` 交互接口，事务和非事务访问均可通过相同接口完成；
- 扩展了原生 sql.Rows/sql.Row，使其支持 MapScan 以及 SliceScan；

## API文档

https://pkg.go.dev/github.com/bunnier/sqlmer

> Tips: 主交互接口为[DbClient](/db_client.go)。

## 简单样例

```go
func main() {
	// 获取 DbClient，这里使用 SqlServer 做示范，MySql 也提供了一致的 API 和相应的参数解析逻辑。
	dbClient, err := mssql.NewMsSqlDbClient(
		"server=127.0.0.1:1433; database=test; user id=dev;password=qwer1234;",
		sqlmer.WithConnTimeout(time.Second*15), // 连接超时。
		sqlmer.WithExecTimeout(time.Second*15), // 读写超时(执行语句时候，如果没有指定超时时间，默认用这个)。
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 创建测试表。
	_, err = dbClient.Execute(`
CREATE TABLE MainDemo(
	Id INT PRIMARY KEY IDENTITY(1,1) NOT NULL,
	Name VARCHAR(10) NOT NULL,
	Age INT NOT NULL
)`)
	if err != nil {
		log.Fatal(err)
		return
	}

	// 有超时时间的查询。
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	_, err = dbClient.ExecuteContext(ctx, "WAITFOR DELAY '00:00:02'")
	if err != nil {
		fmt.Println("timeout: " + err.Error()) // 预期内的超时~
	}

	// 索引方式插入数据，@p1...@pn，分别对应第1-n个参数。
	_, err = dbClient.Execute("INSERT INTO MainDemo(Name, Age) VALUES(@p1, @p2)", "rui", 1)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = dbClient.Execute("INSERT INTO MainDemo(Name, Age) VALUES(@p1, @p2)", "bao", 2)
	if err != nil {
		log.Fatal(err)
		return
	}

	// 命名参数查询数据，命名参数采用map，key为sql语句@之后的参数名，value为值。
	data, err := dbClient.Get("SELECT * FROM MainDemo WHERE Name=@name", map[string]interface{}{"name": "rui"})
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(data) // map方式返回，需要结构的，需要自己转换。

	// 获取第一行第一列。
	name, err := dbClient.Scalar("SELECT Name FROM MainDemo WHERE Name=@p1", "rui")
	if err != nil {
		log.Fatal(err)
		return
	}
	
	// 返回interface{}，类型可以自己转换。已经统一了Sql Server和MySql返回的类型（注意：Decimal使用string返回）。
	fmt.Println(name.(string)) 

	// 获取增强后的sql.Rows（支持SliceScan、MapScan）。
	sliceRows, err := dbClient.Rows("SELECT Name, GETDATE() FROM MainDemo WHERE Name IN (@p1, @p2)", "rui", "bao")
	if err != nil {
		log.Fatal(err)
		return
	}
	for sliceRows.Next() {
		sliceRow, err := sliceRows.SliceScan() // SliceScan用[]interface{}方式返回。
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println(sliceRow...)
	}

	if sliceRows.Err() != nil {
		log.Fatal(err)
		return
	}

	// 删除测试表。
	_, err = dbClient.Execute("DROP TABLE MainDemo")
	if err != nil {
		log.Fatal(err)
		return
	}
}
```

## 测试用例

运行测试用例需要：

1. 配置`test_conf.yml`文件，目前必须 SqlServer/MySql 均配置上才能完整运行测试用例；
2. 调用`go run ./internal/testcmd/main.go -a PREPARE -c test_conf.yml`来准备环境；
3. 测试结束后可以通过`go run ./internal/testcmd/main.go -a CLEAN -c test_conf.yml`销毁测试表；

如果你和我一样使用VSCode作为开发工具，可在配置好`test_conf.yml`之后，直接使用.vscode中编写好的Task来准备环境。
