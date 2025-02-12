package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/mysql"
)

var dbClient sqlmer.DbClient // 这是本库的主接口，统一了各种数据库的 API 操作。
var err error                // 本库同时提供了 error/panic 两套 API，为了 demo 更为简洁，后续主要通过 panic(Must) 版本 API 演示。

func init() {
	// 这里使用 MySQL 做示范，SQL Server 也提供了一致的 API 和相应的参数解析逻辑。
	if dbClient, err = mysql.NewMySqlDbClient(
		"test:testpwd@tcp(127.0.0.1:3306)/test",
		sqlmer.WithConnTimeout(time.Second*30), // 连接超时。
		sqlmer.WithExecTimeout(time.Second*30), // 读写超时(执行语句时候，如果没有指定超时时间，默认用这个)。
	); err != nil {
		log.Fatal(err)
	}
}

func main() {
	prepare()
	defer purge()

	selectionDemo()
	ormDemo()
	ormWithFieldConvert()
	transactionDemo()
	timeoutDemo()
}

func prepare() {
	// 创建/删除 测试表。
	dbClient.MustExecute(`
		CREATE TABLE IF NOT EXISTS demo (
			Id int(11) NOT NULL AUTO_INCREMENT,
			Name VARCHAR(10) NOT NULL,
			Age INT NOT NULL,
			Scores VARCHAR(200) NOT NULL,
			PRIMARY KEY (Id),
			KEY demo (Id)
		)
	`)

	// 索引方式插入数据，@p1..@pN，分别对应第 1..n 个参数。
	dbClient.MustExecute("INSERT INTO demo(Name, Age, Scores) VALUES(@p1, @p2, @p3)", "rui", 1, "SCORES:1,3,5,7")
	dbClient.MustExecute("INSERT INTO demo(Name, Age, Scores) VALUES(@p1, @p2, @p3)", "bao", 2, "SCORES:2,4,6,8")
}

func purge() {
	dbClient.MustExecute("DROP TABLE demo")
}

func selectionDemo() {
	// 命名参数查询数据，命名参数采用 map，key 为 sql 语句 @ 之后的参数名，value 为值。
	dataMap := dbClient.MustGet("SELECT * FROM demo WHERE Name=@Name", map[string]any{"Name": "rui"})
	fmt.Println(dataMap) // Output: map[Age:1 Id:1 Name:rui Scores:SCORES:1,3,5,7]

	// 命名参数查询数据，命名参数采用 struct ，字段名为 sql 语句 @ 之后的参数名，字段值为参数值。
	type Params struct {
		Name string
	}
	dataMap = dbClient.MustGet("SELECT * FROM demo WHERE Name=@Name", Params{Name: "rui"})
	fmt.Println(dataMap) // Output: map[Age:1 Id:1 Name:rui Scores:SCORES:1,3,5,7]

	// 可提供多个参数，DbClient 会自动进行参数合并，优先取靠后的参数中的同名字段（ struct 和 map 可互相覆盖）。
	dataMap = dbClient.MustGet("SELECT * FROM demo WHERE Name=@Name", Params{Name: "rui"}, map[string]any{"Name": "bao"})
	fmt.Println(dataMap) // Output: map[Age:2 Id:2 Name:bao Scores:SCORES:2,4,6,8]

	// 可通过 @p1.。。@pN 方式，指定参数的位置，参数位置从 1 开始。
	name, _ := dbClient.MustScalar("SELECT Name FROM demo WHERE Name=@p1", "rui")
	fmt.Println(name.(string))

	// 可混用 struct / map / 索引参数，DbClient 会自动进行参数合并。
	// 下面这个语句的 3 个参数，DbClient 进行合并后最后的参数列表是： @p1=rui, @Name="bao"
	count, _ := dbClient.MustScalar("SELECT COUNT(1) FROM demo WHERE Name=@p1 OR Name=@Name",
		map[string]any{"Name": "other"},
		"rui",
		Params{Name: "bao"},
	)
	fmt.Println(count.(int64)) // Output: 2

	// 如果喜欢标准库风格，这里也提供了增强版本的 sql.Rows，支持 SliceScan、MapScan。
	// 注意：
	//   - 这里用到了 slice/array 的参数展开特性
	//   - 如果传入的 slice/array 长度为 0，会被解析为 NULL ，会导致 in/not in 语句均为 false；
	rows := dbClient.MustRows("SELECT Name, now() FROM demo WHERE Name IN (@p1)", []any{"rui", "bao"})
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
}

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

// 演示如何使用数据库事务。
func transactionDemo() {
	rowNum, _ := dbClient.MustScalar("SELECT count(1) FROM demo")
	fmt.Println(rowNum) // Output: 2

	// CreateTransaction 返回一个 TransactionKeeper ，
	// 它实现了 DbClient ，所以数据库操作方式和一般的 DbClient 完全一致。
	trans := dbClient.MustCreateTransaction()

	// 如果 TransactionKeeper.Commit/MustCommit 没有被调用，则 Close 操作会回滚事务；
	// 若事务已提交，则 Close 操作仅关闭连接。
	defer trans.MustClose()

	trans.MustExecute("DELETE FROM demo WHERE Id=1")

	// 支持使用嵌套事务。
	func() {
		embeddedTrans := trans.MustCreateTransaction()
		defer embeddedTrans.MustClose() // 注意：嵌套事务也需要 Close 。

		embeddedTrans.MustExecute("DELETE FROM demo WHERE Id=2")
		embeddedTrans.MustCommit()
	}()

	// 提交外层的事务。
	trans.MustCommit()

	rowNum, _ = dbClient.MustScalar("SELECT count(1) FROM demo")
	fmt.Println(rowNum) // Output: 0
}

// 演示如何设置超时时间。 DbClient 中的方法，都有一个 Context 版，支持传入 Context 以设置超时。
// 如 Execute 对应 ExecuteContext 。
func timeoutDemo() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
	if _, err = dbClient.ExecuteContext(ctx, "SELECT sleep(3)"); err != nil {
		fmt.Println("timeout: " + err.Error()) // 预期内的超时~
	}
}
