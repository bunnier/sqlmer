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
	flag.StringVar(&configFile, "c", "test_conf.yml", "conf file path")
	flag.Parse()
}

// 这个 cmd 用于准备或销毁测试环境数据。
func main() {
	conf, err := testenv.LoadTestConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	switch action {
	// 准备测试表。
	case "PREPARE":
		Prepare(conf)

	// 销毁测试表。
	case "CLEAN":
		Clean(conf)

	default:
		log.Fatal("undefined action:" + action)
	}
}

func Prepare(testConf testenv.TestConf) {
	var errgroup errgroup.Group

	// 初始化SqlServer测试表。
	errgroup.Go(func() error {
		db, err := getDb(mssql.DriverName, testConf.SqlServer)
		if err != nil {
			return err
		}
		return testenv.CreateMssqlSchema(db)
	})

	// 初始化 MySql 测试表。
	errgroup.Go(func() error {
		db, err := getDb(mysql.DriverName, testConf.MySql)
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

func Clean(testConf testenv.TestConf) {
	var errgroup errgroup.Group

	// 销毁 SqlServer 测试表。
	errgroup.Go(func() error {
		db, err := getDb(mssql.DriverName, testConf.SqlServer)
		if err != nil {
			return err
		}
		return testenv.DropMssqlSchema(db)
	})

	// 销毁 MySql 测试表。
	errgroup.Go(func() error {
		db, err := getDb(mysql.DriverName, testConf.MySql)
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

func getDb(driver string, connectionString string) (*sql.DB, error) {
	db, err := sql.Open(driver, connectionString)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, err
}
