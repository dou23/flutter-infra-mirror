package main

import (
	"flag"
	"flutter-mirror/internal/config"
)

func parseConf() config.Config {
	// 定义镜像服务配置
	conf := config.Config{}

	// 获取命令行参数
	flag.IntVar(&conf.Port, "port", 8050, "Port for the mirror server")
	flag.StringVar(&conf.IP, "ip", "0.0.0.0", "IP address for the mirror server")

	flag.StringVar(&conf.CacheRootPath, "cache", "", "Root path for cache storage")

	// 解析命令行参数
	flag.Parse()

	// 设置全局配置
	config.SetConfig(conf)
	return conf
}
