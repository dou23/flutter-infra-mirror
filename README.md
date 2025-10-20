# Flutter Infra 镜像服务器

一个用于Flutter基础设施资源的本地镜像服务器，可缓存远程资源以提高下载速度和可靠性。

## 概述

该项目创建一个本地HTTP服务器，作为Flutter基础设施资源的镜像。当Flutter工具请求资源时，服务器会：
1. 首先检查本地缓存
2. 如果未缓存，则从远程服务器下载(https://storage.flutter-io.cn)
3. 在本地缓存该资源
4. 将资源提供给客户端

这种方法减少了对外部网络连接的依赖，并加快了重复资源下载的速度。

## 工作原理

Flutter工具 → http://localhost:8050 → 检查缓存 → (缓存命中) 提供本地文件 → (缓存未命中) 下载 → 缓存 → 提供


## 快速开始

1. 运行镜像服务器：
   ```bash
   go run cmd/mirror/main.go
   ```

2. 配置Flutter使用本地镜像：
   Linux:
   ```bash
   export FLUTTER_STORAGE_BASE_URL=http://localhost:8050
   ```
   Windows (命令提示符):
   ```cmd:
   set FLUTTER_STORAGE_BASE_URL=http://localhost:8050
   ```
   PowerShell (PowerShell):
   ```powershell:
   $env:FLUTTER_STORAGE_BASE_URL="http://localhost:8050"
   ```

3. 正常使用Flutter：
   ```bash
   flutter pub get
   flutter upgrade
   # 所有基础设施请求现在都会通过本地镜像
   ```

# 配置选项

## 端口配置

默认情况下，服务器运行在8050端口。您可以使用命令行参数更改：

```bash 
go run cmd/mirror/main.go -port 8080
```

IP地址配置
默认监听所有网络接口(0.0.0.0)，可以通过以下参数指定：

```bash
go run cmd/mirror/main.go -ip 127.0.0.1
```

## 缓存位置配置
默认缓存存储在运行目录下的cache文件夹中，可通过以下参数指定：

```bash
go run cmd/mirror/main.go -cache /path/to/cache/directory
```

## 完整命令行参数
* -port : 指定服务器端口 (默认: 8050)
* -ip : 指定监听的IP地址 (默认: 0.0.0.0)
* -cache : 指定缓存目录路径 (默认: 当前目录下的cache文件夹)

```bash
go run cmd/mirror/main.go -port 9090 -ip 127.0.0.1 -cache /tmp/flutter-cache
```

## 功能特性

* ✅ Flutter基础设施资源的本地缓存
* ✅ 透明代理到远程服务器
* ✅ 自动创建目录结构
* ✅ 全面的日志记录
* ✅ HTTP状态码保持
* ✅ 请求头转发
* ✅ 可配置的端口、IP和缓存路径

# 项目结构
├── cmd/
│   └── mirror/
│       └── main.go          # 程序入口点
├── internal/
│   ├── config/              # 配置管理
│   ├── proxy/               # 代理和缓存逻辑
│   ├── serve/               # HTTP服务器实现
│   └── store/               # 数据存储接口
└── README.md

## 使用建议

1. 持久缓存：从固定位置运行服务器以在会话间保持缓存
2. 磁盘空间：监控缓存目录大小，因为它会随着使用而增长
3. 网络问题：如果遇到网络问题，请检查远程服务器(https://storage.flutter-io.cn)是否可访问

## 日志说明
服务器会输出以下类型的日志信息：

* CACHE HIT: 从缓存提供文件
* CACHE MISS: 需要从远程下载文件
* INFO: 一般信息日志
* ERROR: 错误信息

## 故障排除
1. 端口冲突：如果8050端口被占用，请使用-port标志指定不同端口
2. 权限问题：确保服务器对运行目录有写权限
3. 缓存问题：删除cache目录以清除所有缓存资源
4. 网络连接：确认可以访问 https://storage.flutter-io.cn

## 构建可执行文件
您可以构建一个独立的可执行文件以便部署：

```bash
go build -o flutter-mirror cmd/mirror/main.go
./flutter-mirror
```

## 系统要求

* Go 1.16或更高版本

## 贡献
欢迎提交问题和拉取请求来改进这个镜像服务器。