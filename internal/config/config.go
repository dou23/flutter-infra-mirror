package config

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
