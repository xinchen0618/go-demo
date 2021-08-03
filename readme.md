## Golang

### 技术栈

- 路由      Gin         https://github.com/gin-gonic/gin
- 配置      viper       https://github.com/spf13/viper
- Mysql     GoRose      https://github.com/gohouse/gorose
- Redis     go-redis    https://github.com/go-redis/redis
- 登录      jwt-go      https://github.com/dgrijalva/jwt-go
- 日志      zap         https://github.com/uber-go/zap
- 优雅停止  endless     https://github.com/fvbock/endless
- 命令行    cli          https://github.com/urfave/cli


###  规范

- 项目布局  
  
  Standard Go Project Layout  https://github.com/golang-standards/project-layout
  
- 编码规范  
  
  Uber Go 语言编码规范 https://github.com/xxjwxc/uber_go_guide_cn


### 目录结构

```
- cmd/                  项目入口
  - cli/                命令行
  - restful/            RESTful API   
- config/               配置
  - di/                 服务注入
  - config.yaml         公共配置. 环境配置 config_{RUNTIME_ENV}.yaml, 同键名环境配置覆盖公共配置. dev环境配置不参与版本控制.
  - config_prod.yaml    生产环境配置
  - config_testing.yaml 测试环境配置
  - constants.go        常量定义. Redis key统一在此定义避免冲突.
  - init.go             配置初始化. cmd中的应用都应首先调用此中Init()方法初始化配置.
- internal/             内部应用代码库
  - action/             Cli action
  - controller/         RESTful控制器
  - router/             RESTful路由
    - router.go         路由注册入口. 路由声明按业务分拆到不同文件, 然后统一在此注册.
  - service/            公共业务逻辑
    - cache_service.go  资源缓存服务
- pkg/                  外部应用可以使用的库代码
  - ginx/               gin增强方法
  - gox/                golang增强方法
- go.mod                包管理  
```


### 环境定义

环境变量 `RUNTIME_ENV` 指定执行环境. 默认为生产环境. 参考 <a href="https://en.wikipedia.org/wiki/Deployment_environment" target="_blank">Deployment environment</a>

- `dev`       开发环境
- `testing`   测试环境
- `stage`     预发布环境
- `prod`      生产环境


### 日志

错误日志会打印到console, 同时记录到错误日志文件. 错误日志文件路径在`config/`中配置, 默认为`/var/log/golang_error.log`. 注意文件要有读写权限.


### RESTful

#### 指南

RESTful参考 <a href="https://www.vinaysahni.com/best-practices-for-a-pragmatic-restful-api" target="_blank">Best Practices for Designing a Pragmatic RESTful API</a>

#### 流程

`cmd/restful/main.go` -> `internal/router/` -> `internal/controller/` [-> `internal/service/`]

    - `internal/router/` 路由, API版本在此控制, Major[.Minor], 比如 /v1, /v1.1, API出现向下不兼容且旧版仍需继续使用的情况, ~~比如不升级的旧版APP,~~ 新增Minor版本号. 业务出现结构性变化, 新增Major版本号.
    - `internal/controller/` 用于处理业务, 事务控制尽量放置在这里, 放置在 `internal/service/` 中容易出现事务嵌套的问题.
    - `internal/service/` 用于封装公共的业务逻辑, 为可选.
  
#### 登录

- 登录流程

  - 校验账户信息
  - 生成JWT Token
  - 以 `jwt:<userId>:<JWT Signature>`的格式记录入redis白名单
  - JWT Token返回给客户端

- 校验登录

  - 客户端请求时Header携带JWT Token `Authorization: Bearer <token>`
  - 校验JWT Token
  - 校验redis白名单
  
- 退出登录
 
  - 校验登录
  - 删除对应redis白名单

#### 运行

- 开发&测试环境使用Air实时热重载

  注意, 是否配置了Go mod代理 `export GOPROXY=https://goproxy.cn,direct`, 是否安装了Air `go get github.com/cosmtrek/air`, 是否配置了Go bin路径 `export PATH=$PATH:$HOME/go/bin`.

```
cd go-demo/cmd/restful
RUNTIME_ENV=testing air
```

- 预发布&生产环境执行编译好的程序

  实际上会提前编译好直接部署到机器上.

```
# 启动
cd go-demo/cmd/restful
go build  
(RUNTIME_ENV=prod ./restful &> /dev/null &)

# 优雅重启
kill -SIGHUP $(ps aux | grep -v grep | grep restful | awk '{print $2}')

# 优雅停止
kill -SIGINT $(ps aux | grep -v grep | grep restful | awk '{print $2}')
```


### Cli

Cli按业务维度分两级. 使用方式形如

```
cd go-demo/cmd/cli
go build
./cli <task> <action> [param]
```


### 缓存

- 资源缓存

  缓存以资源对象为单位. 

  - `POST`/`PUT`资源时`service.CacheService.Set()`设置缓存
  - `GET`资源时`service.CacheService.Get()`获取缓存(缓存不存在时会建立)
  - `DELETE`资源时`service.CacheService.Delete()`删除缓存

- 业务缓存

  针对业务设计的缓存. key统一在`config/constants.go`中定义.
