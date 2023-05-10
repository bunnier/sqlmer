package testenv

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/mssql"
	"github.com/bunnier/sqlmer/mysql"
)

type schema struct {
	Create string
	Drop   string
}

const (
	DefaultMysqlConnection = "testuser:testuser@tcp(127.0.0.1:3306)/test"
	DefaultMssqlConnection = "server=127.0.0.1.124; database=test; user id=testuser;password=testuser;"
	DefaultTimeout         = 15 * time.Second
)

type Conf struct {
	Mysql     string
	SqlServer string
}

var TestConf Conf = Conf{
	Mysql:     DefaultMysqlConnection,
	SqlServer: DefaultMssqlConnection,
}

// 加载自定义配置。若给定一个 .json 文件，则读取该文件；否则认为给定的是一个目录，读取该目录下的 .db.json 文件。
func TryInitConfig(path string) {
	if !strings.HasSuffix(strings.ToLower(path), ".json") {
		path += "/.db.json"
	}

	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("cannot read .db.json, use the default database: " + err.Error())
		return
	}

	var conf struct {
		Mysql     string `json:"mysql"`
		SqlServer string `json:"sqlserver"`
	}
	err = json.Unmarshal(content, &conf)
	if err != nil {
		fmt.Println(".db.json format error, use the default database: " + err.Error())
		return
	}

	TestConf.Mysql = conf.Mysql
	TestConf.SqlServer = conf.SqlServer
}

func NewMysqlClient() (sqlmer.DbClient, error) {
	return mysql.NewMySqlDbClient(
		TestConf.Mysql,
		sqlmer.WithConnTimeout(DefaultTimeout),
		sqlmer.WithExecTimeout(DefaultTimeout),
	)
}

func NewSqlServerClient() (sqlmer.DbClient, error) {
	return mssql.NewMsSqlDbClient(
		TestConf.SqlServer,
		sqlmer.WithConnTimeout(DefaultTimeout),
		sqlmer.WithExecTimeout(DefaultTimeout),
	)
}
