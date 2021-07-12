package testenv

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// TestConf 为测试用例配置文件的结构。
type TestConf struct {
	MySql     string `yaml:"mysql"`     // 测试用例使用的MySql连接字符串。
	SqlServer string `yaml:"sqlserver"` // 测试用例使用的SqlServer连接字符串。
}

// LoadTestConfig 用于读取yaml定义的配置文件，并转换为相应protobuf定义的结构。
func LoadTestConfig(confPath string) (TestConf, error) {
	yamlBytes, err := ioutil.ReadFile(confPath)
	if err != nil {
		return TestConf{}, err
	}

	var testConf TestConf
	err = yaml.Unmarshal(yamlBytes, &testConf)
	return testConf, err
}
