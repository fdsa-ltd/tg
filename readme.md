# 微前端框架

## 后台技术

使用golang的http server + html template 实现一个轻量的微前端后台服务容器。主要完成：
1. 动态配置扫描(包含微前端应用的加载)
2. 通过模板生成微前端的基座
3. 请求路由（包含微前端动态url的处理和请求代理）

## 前台技术

1. 使用single-spa作为前端基座
1. 使用vue统一前端基础操作
1. 使用基于bootstrap布局网页

## go使用方法

Golang支持交叉编译，也就是说你在32位平台的机器上开发，可以编译生成64位平台上的可执行程序。

交叉编译依赖下面几个环境变量：

$GOARCH    目标平台（编译后的目标平台）的处理器架构（386、amd64、arm）
$GOOS          目标平台（编译后的目标平台）的操作系统（darwin、freebsd、linux、windows）

```cmd
# build for linux
set GOOS=linux
set GOARCH=amd64
go build app.go

# build for windows
set GOOS=windows
set GOARCH=amd64
go build app.go
```

