package cmd

import (
	"flag"
	"flutter-mirror/internal/config"
	"flutter-mirror/internal/serve"
)

func main() {

	// 定义镜像服务配置
	conf := config.Config{}

	// 获取命令行参数
	flag.StringVar(&conf.Port, "port", "", "Port for the mirror server")
	if conf.Port == "" || len(conf.Port) == 0 {
		flag.StringVar(&conf.Port, "p", ":8050", "Port for the mirror server")
	}

	flag.StringVar(&conf.CacheRootPath, "cache", "", "Root path for cache storage")

	// 解析命令行参数
	flag.Parse()

	// 设置全局配置
	config.SetConfig(conf)

	serve.NewServer().Start(conf.Port)
}
