package main

import (
	"flutter-mirror/internal/serve"
	"fmt"
)

func main() {
	conf := parseConf()
	addr := fmt.Sprintf("%s:%d", conf.IP, conf.Port)
	serve.NewServer().Start(addr)
}
