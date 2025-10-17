package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)


// AdvancedMirrorHandler with local caching capability
func AdvancedMirrorHandler(w http.ResponseWriter, r *http.Request) {
	targetURL := "https://storage.flutter-io.cn"

	// 获取当前工作目录作为缓存目录的基础路径
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("ERROR: Failed to get current directory: %v", err)
		http.Error(w, "Failed to get current directory", http.StatusInternalServerError)
		return
	}
	cacheDir := filepath.Join(currentDir, "cache")
	log.Printf("INFO: Using cache directory: %s", cacheDir)

	// 创建缓存目录（如果不存在）
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Printf("ERROR: Failed to create cache directory: %v", err)
		http.Error(w, "Failed to create cache directory", http.StatusInternalServerError)
		return
	}

	// 生成本地文件路径
	localFilePath := filepath.Join(cacheDir, r.URL.Path)
	log.Printf("INFO: Processing request for path: %s", r.URL.Path)

	// 检查文件是否存在于缓存中
	if _, err := os.Stat(localFilePath); err == nil {
		// 文件存在于缓存中，直接提供本地资源
		log.Printf("CACHE HIT: Serving from cache: %s", r.URL.Path)
		http.ServeFile(w, r, localFilePath)
		log.Printf("SUCCESS: Delivered cached content: %s", r.URL.Path)
		return
	}

	// 文件不在缓存中，从远程服务器下载
	sourceURL := targetURL + r.URL.EscapedPath()
	log.Printf("CACHE MISS: Downloading from remote: %s", sourceURL)

	// 为文件创建目录结构
	if dir := filepath.Dir(localFilePath); dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Printf("ERROR: Failed to create directory structure: %v", err)
			http.Error(w, "Failed to create directory structure", http.StatusInternalServerError)
			return
		}
	}

	// 下载文件
	log.Printf("INFO: Starting download of %s", sourceURL)
	resp, err := http.Get(sourceURL)
	if err != nil {
		log.Printf("ERROR: Failed to download file from %s: %v", sourceURL, err)
		http.Error(w, "Failed to download file", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 检查请求是否成功
	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Remote server returned status %d for %s", resp.StatusCode, sourceURL)
		http.Error(w, "Failed to fetch resource", resp.StatusCode)
		return
	}
	log.Printf("INFO: Successfully connected to remote server, status: %d", resp.StatusCode)

	// 创建本地文件
	localFile, err := os.Create(localFilePath)
	if err != nil {
		log.Printf("ERROR: Failed to create local file %s: %v", localFilePath, err)
		http.Error(w, "Failed to create local file", http.StatusInternalServerError)
		return
	}
	defer localFile.Close()
	log.Printf("INFO: Created local file for caching: %s", localFilePath)

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	log.Printf("INFO: Copied %d header entries from remote response", len(resp.Header))

	// 设置状态码
	w.WriteHeader(resp.StatusCode)
	log.Printf("INFO: Set response status code: %d", resp.StatusCode)

	// 同时写入本地文件和响应
	bytesWritten, err := io.Copy(io.MultiWriter(localFile, w), resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed while copying response: %v", err)
		return
	}

	log.Printf("SUCCESS: Cached and served %s (%d bytes)", r.URL.Path, bytesWritten)
}

func main() {
	// Register handlers
	http.HandleFunc("/", AdvancedMirrorHandler)

	// Start server
	port := ":8050"
	fmt.Printf("Starting mirror server on port %s\n", port)
	fmt.Printf("Access mirrored content via: http://localhost%s/[path-to-resource]\n", port)
	// https://storage.flutter-io.cn
	log.Fatal(http.ListenAndServe(port, nil))
}
