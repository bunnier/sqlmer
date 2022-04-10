package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/bunnier/sqlmer/internal/testenv"
	"github.com/bunnier/sqlmer/mssql"
	"github.com/bunnier/sqlmer/mysql"
	"golang.org/x/sync/errgroup"
)

var action string     // 行为。
var configFile string // 配置文件路径。

func init() {
	flag.StringVar(&action, "a", "", "execute action.")
	flag.StringVar(&configFile, "c", ".", "conf file path")
	flag.Parse()
}

// 这个 cmd 用于准备或销毁测试环境数据。
func main() {
	testenv.TryInitConfig(configFile)

	switch action {
	// 准备测试表。
	case "PREPARE":
		Prepare()

	// 销毁测试表。
	case "CLEAN":
		Clean()

	default:
		log.Fatal("undefined action:" + action)
	}
}

func Prepare() {
	var errgroup errgroup.Group

	// 初始化SqlServer测试表。
	errgroup.Go(func() error {
		db, err := getDb(mssql.DriverName, testenv.TestConf.SqlServer)
		if err != nil {
			return err
		}
		return testenv.CreateMssqlSchema(db)
	})

	// 初始化 MySql 测试表。
	errgroup.Go(func() error {
		db, err := getDb(mysql.DriverName, testenv.TestConf.Mysql)
		if err != nil {
			return err
		}
		return testenv.CreateMysqlSchema(db)
	})

	if err := errgroup.Wait(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Prepared.")
}

func Clean() {
	var errgroup errgroup.Group

	// 销毁 SqlServer 测试表。
	errgroup.Go(func() error {
		db, err := getDb(mssql.DriverName, testenv.TestConf.SqlServer)
		if err != nil {
			return err
		}
		return testenv.DropMssqlSchema(db)
	})

	// 销毁 MySql 测试表。
	errgroup.Go(func() error {
		db, err := getDb(mysql.DriverName, testenv.TestConf.Mysql)
		if err != nil {
			return err
		}
		return testenv.DropMysqlSchema(db)
	})

	if err := errgroup.Wait(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Cleaned.")
}

func getDb(driver string, Dsn string) (*sql.DB, error) {
	db, err := sql.Open(driver, Dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, err
}
