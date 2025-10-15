# flutter-infra-mirror
https://storage.googleapis.com/flutter_infra_release/ 请求和文件缓存的镜像

这个项目目前思路很简单: 运行项目启动服务http://localhost:8050->设置FLUTTER_STORAGE_BASE_URL:http://localhost:8050->当Flutter运行时，访问assets便会先请求服务，服务会先检查本地缓存，有缓存直接返回，没有缓存就去请求https://storage.flutter-io.cn然后缓存数据，再返回响应

需要把环境变量FLUTTER_STORAGE_BASE_URL修改为http://localhost:8050

目前端口写死了8050
