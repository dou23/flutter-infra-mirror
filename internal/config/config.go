package config

import "flag"

type Config struct {
	CacheRootPath string // 缓存目录
	Port          int    // 镜像服务监听端口
	IP            string // 镜像服务监听IP
}

var _config Config

func SetConfig(config Config) {
	_config = config
}

func GetConfig() *Config {
	return &_config
}

func ParseConf() Config {
	// 定义镜像服务配置
	conf := Config{}

	// 获取命令行参数
	flag.IntVar(&conf.Port, "port", 0, "Port for the mirror server")
	if conf.Port == 0 {
		flag.IntVar(&conf.Port, "p", 8050, "Port for the mirror server")
	}
	flag.StringVar(&conf.IP, "ip", "0.0.0.0", "IP address for the mirror server")

	flag.StringVar(&conf.CacheRootPath, "cache", "", "Root path for cache storage")

	// 解析命令行参数
	flag.Parse()

	// 设置全局配置
	SetConfig(conf)
	return conf
}
