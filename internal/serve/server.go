package serve

import (
	"flutter-mirror/internal/proxy"
	"fmt"
	"log"
	"net/http"
)

type Server interface {
	Start(addr string) error
}

type ProxyServer struct {
}

func NewServer() Server {
	// Register handlers
	http.HandleFunc("/", proxy.AdvancedMirrorHandler)
	return &ProxyServer{}
}

func (s *ProxyServer) Start(addr string) error {
	// Start server
	fmt.Printf("Starting mirror server on port %s\n", addr)
	fmt.Printf("Access mirrored content via: http://%s/[path-to-resource]\n", addr)
	// https://storage.flutter-io.cn
	err := http.ListenAndServe(addr, nil)
	log.Fatal(err)
	return err
}
