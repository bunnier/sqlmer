package sqlen

import (
	"log"

	"github.com/bunnier/sqlmer/internal/testenv"
)

var testConf testenv.TestConf

// 这个init用于加载测试用的数据库配置。
func init() {
	conf, err := testenv.LoadTestConfig("../../test_conf.yml")
	if err != nil {
		log.Fatalf("testenv.LoadTestConfig error, err=%v", err)
	}
	testConf = conf
}
