package config

import (
	"io/ioutil"

	"github.com/quanxiang-cloud/audit/pkg/misc/elastic2"
	"github.com/quanxiang-cloud/audit/pkg/misc/kafka"
	"github.com/quanxiang-cloud/audit/pkg/misc/logger"

	"gopkg.in/yaml.v2"
)

// Conf 全局配置文件
var Conf *Config

// DefaultPath 默认配置路径
var DefaultPath = "./configs/config.yml"

// Config 配置文件
type Config struct {
	Port    string          `yaml:"port"`
	Model   string          `yaml:"model"`
	GEOIP   string          `yaml:"geoIP"`
	Log     logger.Config   `yaml:"log"`
	Elastic elastic2.Config `yaml:"elastic"`
	Kafka   kafka.Config    `yaml:"kafka"`
	Handler Handler         `yaml:"handler"`
}

// NewConfig 获取配置配置
func NewConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, &Conf)
	if err != nil {
		return nil, err
	}

	return Conf, nil
}

// Handler handler
type Handler struct {
	Topic          []string `yaml:"topic"`
	Group          string   `yaml:"group"`
	NumOfProcessor int      `yaml:"numOfProcessor"`
	Buffer         int      `yaml:"buffer"`
}
