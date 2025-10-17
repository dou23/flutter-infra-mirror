package serve

import (
	"flutter-mirror/internal/proxy"
	"fmt"
	"log"
	"net/http"
)

type Server interface {
	Start(port string) error
}

type ProxyServer struct {
}

func NewServer() Server {
		// Register handlers
	http.HandleFunc("/", proxy.AdvancedMirrorHandler)
	return &ProxyServer{}
}

func (s *ProxyServer) Start(port string) error {
	// Start server
	fmt.Printf("Starting mirror server on port %s\n", port)
	fmt.Printf("Access mirrored content via: http://localhost%s/[path-to-resource]\n", port)
	// https://storage.flutter-io.cn
	err := http.ListenAndServe(port, nil)
	log.Fatal(err)
	return err
}
