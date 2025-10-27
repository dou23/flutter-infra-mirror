package proxy

import (
	"flutter-mirror/internal/config"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var ReleasesJSONWindowsName = "releases_windows.json"
var ReleasesJSONLinuxName = "releases_linux.json"
var ReleasesJSONMacOSName = "releases_macos.json"
var ReleasesJsonPath = "/flutter_infra_release/releases/%s"
var FIRStr = "/flutter_infra_release"

// 正在下载的文件映射，避免并发下载同一文件
var downloadingFiles = make(map[string]bool)
var downloadingMutex sync.RWMutex

// AdvancedMirrorHandler with local caching capability
func AdvancedMirrorHandler(w http.ResponseWriter, r *http.Request) {
	targetURL := "https://storage.flutter-io.cn"

	// 获取当前工作目录作为缓存目录的基础路径
	var currentDir string
	var err error
	if len(config.GetConfig().CacheRootPath) != 0 {
		currentDir = config.GetConfig().CacheRootPath
	} else {
		currentDir, err = os.Getwd()
	}

	if err != nil {
		log.Printf("ERROR: Failed to get current directory: %v", err)
		http.Error(w, "Failed to get current directory", http.StatusInternalServerError)
		return
	}

	cacheDir := filepath.Join(currentDir, "cache")
	log.Printf("INFO: Using cache directory: %s", cacheDir)

	// 检查 /flutter_infra_release/releases/releases_os.json 特例 需要特殊处理以确保缓存更新
	// releases_os.json 文件经常更新，但文件名不变，因此每次请求都需要重新下载并更新缓存
	ReleasesJsonWindowsPath := fmt.Sprintf(ReleasesJsonPath, ReleasesJSONWindowsName)
	ReleasesJsonLinuxPath := fmt.Sprintf(ReleasesJsonPath, ReleasesJSONLinuxName)
	ReleasesJsonMacOSPath := fmt.Sprintf(ReleasesJsonPath, ReleasesJSONMacOSName)

	// 创建缓存目录（如果不存在）
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Printf("ERROR: Failed to create cache directory: %v", err)
		http.Error(w, "Failed to create cache directory", http.StatusInternalServerError)
		return
	}

	// 生成本地文件路径
	localFilePath := filepath.Join(cacheDir, r.URL.Path)
	log.Printf("INFO: Processing request for path: %s", r.URL.Path)

	// 特殊处理 releases JSON 文件
	if r.URL.Path == ReleasesJsonWindowsPath || r.URL.Path == ReleasesJsonLinuxPath || r.URL.Path == ReleasesJsonMacOSPath {
		handleReleasesJSON(w, r, targetURL, localFilePath)
		return
	}

	// 检查是否有其他goroutine正在下载此文件
	if isFileDownloading(localFilePath) {
		// 等待下载完成
		waitForDownload(localFilePath)
	}

	// 检查文件是否存在于缓存中且完整
	if isFileValid(localFilePath) {
		// 文件存在于缓存中且完整，直接提供本地资源
		log.Printf("CACHE HIT: Serving from cache: %s", r.URL.Path)
		http.ServeFile(w, r, localFilePath)
		log.Printf("SUCCESS: Delivered cached content: %s", r.URL.Path)
		return
	}

	// 文件不在缓存中或不完整，从远程服务器下载
	downloadAndCacheFile(w, r, targetURL, localFilePath)
}

// handleReleasesJSON 处理 releases JSON 文件的特殊逻辑
func handleReleasesJSON(w http.ResponseWriter, r *http.Request, targetURL, localFilePath string) {
	// 标记文件正在下载
	markFileAsDownloading(localFilePath)
	defer markFileAsDownloaded(localFilePath)

	// 拼接远程文件URL
	sourceURL := targetURL + r.URL.EscapedPath()
	log.Printf("CACHE MISS: Downloading releases JSON from remote: %s", sourceURL)

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

		// 检查文件是否存在于缓存中
		if isFileValid(localFilePath) {
			// 文件存在于缓存中且完整，直接提供本地资源
			log.Printf("CACHE HIT: Serving from cache: %s", r.URL.Path)
			http.ServeFile(w, r, localFilePath)
			log.Printf("SUCCESS: Delivered cached content: %s", r.URL.Path)
			return
		}

		return
	}

	// 创建临时文件
	tempFilePath := localFilePath + ".tmp"
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		log.Printf("ERROR: Failed to create temporary file %s: %v", tempFilePath, err)
		http.Error(w, "Failed to create temporary file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

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

	// 同时写入临时文件和响应
	bytesWritten, err := io.Copy(io.MultiWriter(tempFile, w), resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed while copying response: %v", err)
		// 删除临时文件
		os.Remove(tempFilePath)
		return
	}

	// 关闭临时文件
	tempFile.Close()

	// 原子性地将临时文件重命名为目标文件
	if err := os.Rename(tempFilePath, localFilePath); err != nil {
		log.Printf("ERROR: Failed to rename temporary file: %v", err)
		os.Remove(tempFilePath)
		return
	}

	log.Printf("SUCCESS: Cached and served %s (%d bytes)", r.URL.Path, bytesWritten)
}

// downloadAndCacheFile 下载并缓存文件
func downloadAndCacheFile(w http.ResponseWriter, r *http.Request, targetURL, localFilePath string) {
	// 标记文件正在下载
	markFileAsDownloading(localFilePath)
	defer markFileAsDownloaded(localFilePath)

	// 文件不在缓存中或不完整，从远程服务器下载
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

	// 创建临时文件
	tempFilePath := localFilePath + ".tmp"
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		log.Printf("ERROR: Failed to create temporary file %s: %v", tempFilePath, err)
		http.Error(w, "Failed to create temporary file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

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

	// 同时写入临时文件和响应
	bytesWritten, err := io.Copy(io.MultiWriter(tempFile, w), resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed while copying response: %v", err)
		// 删除临时文件
		os.Remove(tempFilePath)
		return
	}

	// 关闭临时文件
	tempFile.Close()

	// 原子性地将临时文件重命名为目标文件
	if err := os.Rename(tempFilePath, localFilePath); err != nil {
		log.Printf("ERROR: Failed to rename temporary file: %v", err)
		os.Remove(tempFilePath)
		return
	}

	log.Printf("SUCCESS: Cached and served %s (%d bytes)", r.URL.Path, bytesWritten)
}

// isFileValid 检查文件是否存在且完整
func isFileValid(filePath string) bool {
	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Printf("ERROR: Failed to stat file %s: %v", filePath, err)
		return false
	}

	// 检查文件大小是否合理（可以添加更复杂的验证逻辑）
	if fileInfo.Size() == 0 {
		log.Printf("WARN: File %s has zero size, treating as invalid", filePath)
		return false
	}

	// 检查是否存在临时文件（表示下载未完成）
	tempFilePath := filePath + ".tmp"
	if _, err := os.Stat(tempFilePath); err == nil {
		// 临时文件存在，说明之前的下载未完成，删除不完整的文件
		log.Printf("WARN: Found incomplete download for %s, removing", filePath)
		os.Remove(filePath)
		return false
	}

	return true
}

// isFileDownloading 检查文件是否正在下载
func isFileDownloading(filePath string) bool {
	downloadingMutex.RLock()
	defer downloadingMutex.RUnlock()
	return downloadingFiles[filePath]
}

// markFileAsDownloading 标记文件正在下载
func markFileAsDownloading(filePath string) {
	downloadingMutex.Lock()
	defer downloadingMutex.Unlock()
	downloadingFiles[filePath] = true
}

// markFileAsDownloaded 标记文件下载完成
func markFileAsDownloaded(filePath string) {
	downloadingMutex.Lock()
	defer downloadingMutex.Unlock()
	delete(downloadingFiles, filePath)
}

// waitForDownload 等待文件下载完成
func waitForDownload(filePath string) {
	// 简单实现：等待一段时间，实际项目中可以使用channel等更优雅的方式
	// 这里只是示例，实际应用中应该使用更完善的同步机制
	for i := 0; i < 100; i++ {
		downloadingMutex.RLock()
		isDownloading := downloadingFiles[filePath]
		downloadingMutex.RUnlock()

		if !isDownloading {
			break
		}

		// 等待100ms
		// time.Sleep(100 * time.Millisecond)
	}
}
