# sqlmer

[![License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![Go](https://github.com/bunnier/sqlmer/actions/workflows/go.yml/badge.svg)](https://github.com/bunnier/sqlmer/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bunnier/sqlmer)](https://goreportcard.com/report/github.com/bunnier/sqlmer)
[![Go Reference](https://pkg.go.dev/badge/github.com/bunnier/sqlmer.svg)](https://pkg.go.dev/github.com/bunnier/sqlmer)

## 功能简介

一个面向原生 SQL 的数据库操作库，在保持 SQL 语句灵活性的同时，通过统一的接口设计、便捷的参数处理和轻量级的 ORM 映射能力，让数据库操作更加优雅高效。目前支持 MySQL / SQLite / SQL Server 。

### 核心特性

- **面向原生 SQL**：核心交互以原生 SQL 语句为基础，提供灵活的结果映射能力，让数据库操作更加简单高效；
- **增强的参数处理**：提供了直观的 `命名参数` 和 `索引参数` 语法，其中 `命名参数` 支持使用 map/struct 作为参数，并允许参数重写，让 SQL 及参数更易维护和重用；
- **轻量化 ORM 体验**：支持将查询结果灵活映射为 Map 或类型安全的结构体，并提供智能的字段匹配机制，让数据操作更加简单直观；
- **一致的接口设计**：提供了直观统一的数据交互接口，事务和非事务操作保持一致的使用体验；
- **增强的原生接口**：针对原生 API 爱好者，也提供了增强的 sql.Rows/sql.Row，增加了自动检测列类型和数量的能力，可智能地将数据 Scan 到合适的类型，轻松装载到 Map 或 Slice；
- **强大的事务支持**：提供了支持嵌套使用的事务 API，优雅地处理复杂的事务场景；
- **便捷的 IN 查询**：支持直接传递 slice/array 类型参数到 IN 子句，无需手动拼接语句；
- **灵活的装饰器模式**：通过 `wrap` 包提供的装饰器模式，可以轻松注入慢日志、监控指标、链路追踪等自定义逻辑，让系统监控和诊断更加便捷；
- **智能的连接管理**：根据数据库配置自动初始化连接生命周期（该特性仅支持 MySQL）；

## 快速上手

让我们通过一系列示例来展示 sqlmer 的强大功能。以下示例将展示如何使用 sqlmer 进行数据库操作，包括基础查询、ORM 映射、事务处理等核心特性。

> 本小节完整例子：[example](https://github.com/bunnier/sqlmer/blob/main/example/main.go)

### 获取 sqlmer

```bash
go get github.com/bunnier/sqlmer
```

### 连接数据库

首先，让我们创建一个数据库连接。sqlmer 提供了统一的接口设计，无论是 MySQL / SQLite / SQL Server，都可以使用相同的 API 进行操作。

```go
var (
	// 这是本库的基础接口，提供了访问数据库的的核心 API 操作。
	dbClient sqlmer.DbClient

	// 这是本库的扩展接口（推荐使用）。
	// DbClientEx 实现了 DbClient 的所有接口，并增加 must 版本 API、ORM 等扩展功能。
	// 后续演示为了简洁，使用了 panic 版本 API，因此主要使用这个接口。
	dbClientEx *sqlmer.DbClientEx
)

func init() {
	// 这里使用 SQLite (sqlite 包)做示范，其它支持的数据库也提供了一致的 API 和相应的参数解析逻辑。
	var err error
	if dbClient, err = sqlite.NewSqliteDbClient(
		"demo.db",
		sqlmer.WithConnTimeout(time.Second*30), // 连接超时。
		sqlmer.WithExecTimeout(time.Second*30), // 读写超时(执行语句时候，如果没有指定超时时间，默认用这个)。
	); err != nil {
		log.Fatal(err)
	}

	dbClientEx = sqlmer.Extend(dbClient)
}
```

### 基础查询操作

sqlmer 支持多种参数传递方式，包括命名参数和索引参数，也支持参数的重写覆盖逻辑，让 SQL 语句更易维护和重用。

```go
// 准备演示数据。
func prepare() {
	// 创建/删除 测试表。
	dbClientEx.MustExecute(`
		CREATE TABLE IF NOT EXISTS demo (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			Age INTEGER NOT NULL,
			Scores TEXT NOT NULL
		)
	`)

	// 索引方式插入数据，@p1..@pN，分别对应第 1..n 个参数。
	dbClientEx.MustExecute("INSERT INTO demo(Name, Age, Scores) VALUES(@p1, @p2, @p3)", "rui", 1, "SCORES:1,3,5,7")
	dbClientEx.MustExecute("INSERT INTO demo(Name, Age, Scores) VALUES(@p1, @p2, @p3)", "bao", 2, "SCORES:2,4,6,8")
}

// 开始演示基础查询功能了~
func selectionDemo() {
	// 命名参数查询数据，参数采用 map 时：key 为 sql 语句 @ 之后的参数名，value 为值。
	dataMap := dbClientEx.MustGet("SELECT * FROM demo WHERE Name=@Name", map[string]any{"Name": "rui"})
	fmt.Println(dataMap) // Output: map[Age:1 Id:1 Name:rui Scores:SCORES:1,3,5,7]

	// 命名参数查询数据，参数采用 struct 时：字段名为 sql 语句 @ 之后的参数名，字段值为参数值。
	type Params struct {
		Name string
	}
	dataMap = dbClientEx.MustGet("SELECT * FROM demo WHERE Name=@Name", Params{Name: "rui"})
	fmt.Println(dataMap) // Output: map[Age:1 Id:1 Name:rui Scores:SCORES:1,3,5,7]

	// 可提供多个参数，DbClient 会自动进行参数合并，优先取靠后的参数中的同名字段（ struct 和 map 可互相覆盖）。
	dataMap = dbClientEx.MustGet("SELECT * FROM demo WHERE Name=@Name", Params{Name: "rui"}, map[string]any{"Name": "bao"})
	fmt.Println(dataMap) // Output: map[Age:2 Id:2 Name:bao Scores:SCORES:2,4,6,8]

	// 可通过 @p1...@pN 方式，指定参数的位置，参数位置从 1 开始。
	name, _ := dbClientEx.MustScalar("SELECT Name FROM demo WHERE Name=@p1", "rui")
	fmt.Println(name.(string))

	// 可混用 struct / map / 索引参数，DbClient 会自动进行参数合并。
	// 下面这个语句的 3 个参数，DbClient 进行合并后最后的参数列表是： @p1=rui, @Name="bao"
	count, _ := dbClientEx.MustScalar(
		"SELECT COUNT(1) FROM demo WHERE Name=@p1 OR Name=@Name",
		map[string]any{"Name": "other"},
		"rui",
		Params{Name: "bao"},
	)
	fmt.Println(count.(int64)) // Output: 2

	// 如果喜欢标准库风格，这里也提供了增强版本的 sql.Rows，支持 SliceScan、MapScan。
	// 注意：
	//   - 这里用到了 slice/array 的参数展开特性
	//   - 如果传入的 slice/array 长度为 0，会被解析为 NULL ，会导致 in/not in 语句均为 false；
	rows := dbClientEx.MustRows("SELECT Name, datetime('now') FROM demo WHERE Name IN (@p1)", []any{"rui", "bao"})
	for rows.Next() {
		// SliceScan 会自动判断列数及列类型，用 []any 方式返回。
		if dataSlice, err := rows.SliceScan(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(dataSlice...)
			// Output:
			// rui 2022-04-09 22:35:33
			// bao 2022-04-09 22:35:33
		}
	}
	// 和标准库一样，Rows 的 Err 和 Close 返回的错误记得要处理哦～
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	if err := rows.Close(); err != nil {
		log.Fatal(err)
	}
}
```

### 轻量级 ORM 映射

sqlmer 提供了简单而强大的轻量级 ORM 映射功能，支持将查询结果直接映射到 Go 结构体。支持驼峰和下划线分割的名称的首字母模糊匹配，让数据操作更加灵活：

```go
// 演示如何使用轻量化 ORM 功能。将数据库的行，映射到 Go struct 。
//
// 由于 Go 语言的限制， struct 的字段必须是大写字母开头的，可能和数据库命名规范不一致。
// 在 ORM 的映射中，可以支持驼峰和下划线分割的名称的首字母模糊匹配，例如：
// struct 字段 GoodName 可以自动匹配到数据库字段 GoodName/goodName/good_name 。
func ormDemo() {
	// 用于表示对应行的数据的类型。
	type Row struct {
		Id     int
		Name   string
		Age    int
		Scores string
	}

	// 轻量化 ORM 的 API 定义在 DbClientEx 里，需要扩展（ Extend ） DbClient 得到。
	clientEx := sqlmer.Extend(dbClient)

	// 获取一行。
	var row Row
	clientEx.MustGetStruct(&row, "SELECT * FROM demo WHERE Id=1")
	fmt.Printf("%v\n", row) // Output: {1 rui 1 1,3,5,7}

	// 获取一个列表。
	//
	// 由于 Golang 不支持方法级的泛型，这里需要通过第一个参数传入一个模板（这里是 Row{} ），指定需要将数据行映射到什么类型；
	// API 返回模板类型的 slice ，以 any 表示，需要再进行类型转换（这里是 []Row ）。
	rows := clientEx.MustListOf(Row{}, "SELECT * FROM demo").([]Row)
	fmt.Println("Rows:")
	for _, v := range rows {
		fmt.Printf("%v\n", v)
	}
	// Output:
	// Rows:
	// {1 rui 1 SCORES:1,3,5,7}
	// {2 bao 2 SCORES:2,4,6,8}

	// 模板可以是 struct 也可以是其指针，指针的写法：
	// clientEx.MustListOf(new(Row), "SELECT * FROM demo").([]*Row)
}

// 演示如何在 ORM 过程中，如果 struct 的目标字段的类型和数据库的类型不能兼容时，如何通过代码定制转换过程。
func ormWithFieldConvert() {
	// 目标类型。
	type Row struct {
		Id     int
		Name   string
		Age    int
		Scores []int
	}

	// 这里利用 Golang 的 struct 内嵌特性，让外层 struct 的同名字段隐藏掉内层的，
	// ORM 赋值过程中， Scores 会赋值到外层，先用兼容的类型（这里是 string ）将数据库的值取出来。
	wrapper := &struct {
		Row
		Scores string
	}{}
	sqlmer.Extend(dbClient).MustGetStruct(wrapper, "SELECT * FROM demo WHERE Id=1")

	// 在已经取到值的基础上，用一段代码，将处理后的值赋值给最终的目标。
	// 这里 Scores 字段的格式是 SCORES: N1,N2,N3,... ，我们的目标格式是将数字部分转换为 []int 。
	row := wrapper.Row
	scores := strings.TrimPrefix(wrapper.Scores, "SCORES:")
	for _, v := range strings.Split(scores, ",") {
		i, _ := strconv.Atoi(v)
		row.Scores = append(row.Scores, i)
	}

	fmt.Printf("%v\n", row) // Output: {1 rui 1 [1 3 5 7]}
}
```

### 事务处理

sqlmer 提供了强大的事务支持，包括嵌套事务，让复杂的事务场景处理变得简单：

```go
// 演示如何使用数据库事务。
func transactionDemo() {
	rowNum, _ := dbClientEx.MustScalar("SELECT count(1) FROM demo")
	fmt.Println(rowNum) // Output: 2

	// CreateTransaction 返回一个 TransactionKeeper，提供了 DbClient 一致的 API。
	// CreateTransactionEx 返回一个 TransactionKeeperEx ，提供了 DbClientEx 一致的 API。
	trans := dbClientEx.MustCreateTransactionEx()

	// 如果 TransactionKeeper.Commit/MustCommit 没有被调用，则 Close 操作会回滚事务；
	// 若事务已提交，则 Close 操作仅关闭连接。
	defer trans.MustClose()

	trans.MustExecute("DELETE FROM demo WHERE Id=1")

	// 支持使用嵌套事务。
	func() {
		embeddedTrans := trans.MustCreateTransactionEx()
		defer embeddedTrans.MustClose() // 注意：嵌套事务也需要 Close 。

		embeddedTrans.MustExecute("DELETE FROM demo WHERE Id=2")
		embeddedTrans.MustCommit()
	}()

	// 提交外层的事务。
	trans.MustCommit()

	rowNum, _ = dbClientEx.MustScalar("SELECT count(1) FROM demo")
	fmt.Println(rowNum) // Output: 0
}
```

### 超时控制

所有数据库操作都支持通过 Context 设置超时，提供更好的系统稳定性：

```go
// 演示如何设置超时时间。 DbClient 中的方法，都有一个 Context 版，支持传入 Context 以设置超时。
// 如 Execute 对应 ExecuteContext 。
func timeoutDemo() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancelFunc()
	// 使用递归查询模拟耗时操作 (SQLite 无法使用 sleep)。
	if _, err := dbClient.ExecuteContext(ctx, "WITH RECURSIVE r(i) AS (VALUES(0) UNION ALL SELECT i+1 FROM r WHERE i < 1000000) SELECT count(*) FROM r"); err != nil {
		fmt.Println("timeout: " + err.Error()) // 预期内的超时~
	}
}
```

### 慢日志、监控指标、链路追踪

`wrap` 子包提供了装饰器模式的 `DbClient` ，可以简单的注入你的慢日志、监控指标、链路追踪等逻辑，下面演示如何注入慢日志。

```go
// 演示如何通过 wrap 包提供的装饰器注入慢日志。
func decoratedDemo() {
	// 通过 wrap 包提供的扩展注入慢日志逻辑。
	dbClient := wrap.Extend(dbClient, func(sql string, args []any) func(error) {
		// 函数会立刻执行。
		startTime := time.Now()

		// 返回的函数会在语句执行后执行。
		return func(err error) {
			duration := time.Since(startTime)
			if duration <= 100*time.Microsecond {
				return
			}

			fmt.Printf("[SlowSql]Sql=%s, Duration=%d(ms), Err=%v", sql, duration/time.Millisecond, err)
		}
	})

	// 因为还是 DbClient 接口，还可以继续扩展为 DbClientEx。
	clientEx := sqlmer.Extend(dbClient)

	// 使用递归查询模拟耗时操作 (SQLite 无法使用 sleep)。
	clientEx.Execute("WITH RECURSIVE r(i) AS (VALUES(0) UNION ALL SELECT i+1 FROM r WHERE i < 500000) SELECT count(*) FROM r") // Output: [SlowSql]Sql=..., Duration=...(ms), Err=<nil>
}
```

## 类型映射

> nullable 的列，如果值为 NULL，均以 nil 返回。

### MySql

| DB datatype                                        | Go datatype |
|----------------------------------------------------|-------------|
| varchar / char / text                              | string      |
| tiny int / small int / int / unsigned int / bigint | int64       |
| float / double                                     | float64     |
| decimal                                            | string      |
| date / datetime / timestamp                        | time.Time   |
| bit                                                | []byte      |

### SQLite

| DB datatype                    | Go datatype |
|--------------------------------|-------------|
| text / varchar / char / clob   | string      |
| integer / int / tiny int / ... | int64       |
| real / double / float          | float64     |
| time / date / datetime         | time.Time   |
| blob                           | []byte      |

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

1. 默认的数据库连接字符串定义在 `internal/testenv/conf.go` 里，默认使用本机的测试库。当需要自定义测试库连接字符串时，可在项目的根（和当前 README 同级）下新建 `.db.json`，当此文件存在时，会自动加载此文件。文件格式为：

		{
			"mysql": "testuser:testuser@tcp(127.0.0.1:3306)/test",
			"sqlserver": "server=127.0.0.1; database=test; user id=testuser;password=testuser;",
			"sqlite": "test.db"
		}

2. 如果第 1 步配置的连接字符串有 DDL 权限，可通过调用 `go run ./internal/testcmd/main.go -a PREPARE` 来同时准备 MySQL / SQL Server 环境，如果没有 DDL 权限可自行直接执行 `internal/testenv/*_prepare.sql` 准备环境；
3. 如果第 1 步配置的连接字符串有 DDL 权限，测试结束后可以通过 `go run ./internal/testcmd/main.go -a CLEAN` 销毁测试表，如果没有 DDL 权限可自行直接执行 `internal/testenv/*_clean.sql` 销毁测试表。

另外，如果你和我一样使用 VSCode 作为开发工具，可直接使用 .vscode 中编写好的 Task 来准备环境。

## 其他语言的版本

.net 版： [cmstar/Data](https://github.com/cmstar/Data)