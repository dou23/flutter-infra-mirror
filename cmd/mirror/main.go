package main

import (
	"flutter-mirror/internal/config"
	"flutter-mirror/internal/serve"
	"fmt"
)

func main() {
	conf := config.ParseConf()
	addr := fmt.Sprintf("%s:%d", conf.IP, conf.Port)
	serve.NewServer().Start(addr)
}
