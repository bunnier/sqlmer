package sqlmer

import (
	"log"

	"github.com/bunnier/sqlmer/internal/testenv"
)

var testConf testenv.TestConf

// 这个 init 用于加载测试用的数据库配置。
func init() {
	conf, err := testenv.LoadTestConfig("test_conf.yml")
	if err != nil {
		log.Fatal(err)
	}
	testConf = conf
}
