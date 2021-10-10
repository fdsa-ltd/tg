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

# TG使用

## 配置文件
配置文件默认./tg.json, 可以通过参数：-c path指定。

- root指定前端文件目录
- port指定服务端口号
- templates 指定动态模板文件，可以设置动态加载微前端
- routers 指定动态路由
  - Name    路由名称
  - Uri     路由资源路径
  - Asserts []断言
  - Filters []处理

## 断言配置

- time from end - 时间在from 和 end 之间
- host keys... - host在keys之中
- mothod in keys... - method 在keys之中
- path keys... - path以keys这中的开头
- ip in keys... - 远程语法IP在keys之中
- query has key[=v] - query中包含key=v的情况
-  cookie has content - cookie中包含content的情况
-  header has key[=v] - header中包含key=v的情况

## 处理配置

- path insert path - 在path前面增加路径
- path append path - 在path之后增加路径
- path remove n - 将path前n段移除
- header k v - 在header中设置k:v
- cookie k v - 在cookie中设置k:v

