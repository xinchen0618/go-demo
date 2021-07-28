## Golang

### 技术栈

- 路由      Gin         https://github.com/gin-gonic/gin
- 配置      viper       https://github.com/spf13/viper
- Mysql     GoRose      https://github.com/gohouse/gorose
- Redis     go-redis    https://github.com/go-redis/redis
- 登录      jwt-go      https://github.com/dgrijalva/jwt-go
- 日志      zap         https://github.com/uber-go/zap
- 优雅停止  endless     https://github.com/fvbock/endless


###  编码规范

参考 <a href="https://github.com/xxjwxc/uber_go_guide_cn" target="_blank">Uber Go 语言编码规范</a>


### 目录结构

```
- config/               配置
  - config.yaml         公共配置. 环境配置 config_{RUNTIME_ENV}.yaml, 同键名环境配置覆盖公共配置. dev环境配置不参与版本控制.
  - config_prod.yaml    生产环境配置
  - config_testing.yaml 测试环境配置
  - constants.go        常量定义. Redis key统一在此定义避免冲突.
- controller/           控制器
- di                    服务注入
  - services.go         Di注册服务
- router/               Restful路由
  - router.go           路由注册入口. 路由声明按业务分拆到不同文件, 然后统一在此注册.
- service/              公共业务逻辑
- util                  工具包
  - api_utils.go        API工具方法
  - tool_utils.go       常用工具方法
- go.mod                包管理  
- main.go               入口  

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

- RESTful指南参考 <a href="https://www.vinaysahni.com/best-practices-for-a-pragmatic-restful-api" target="_blank">Best Practices for Designing a Pragmatic RESTful API</a>

- 流程

  `main.go` -> `router/` -> `controller/` [-> `service/`]

    - `router/` 路由, API版本在此控制, Major[.Minor], 比如 /v1, /v1.1, API出现向下不兼容且旧版仍需继续使用的情况, ~~比如不升级的旧版APP,~~ 新增Minor版本号. 业务出现结构性变化, 新增Major版本号.
    - `controller/` 用于处理业务. 事务控制尽量放置在这里, 放置在 `service/` 中容易出现事务嵌套的问题.
    - `service/` 用于封装公共的业务逻辑, 为可选.


### 运行

- 开发&测试环境使用Air实时热重载
  
  注意, 是否配置了Go mod代理 `export GOPROXY=https://goproxy.cn,direct`, 是否配置了Go bin路径 `export PATH=$PATH:$HOME/go/bin`.

```
cd go-demo
go mod download
go get github.com/cosmtrek/air
RUNTIME_ENV=testing air
```

- 预发布&生产执行编译好的程序
  
  实际上会提前编译好直接部署到机器上

```
# 启动
cd go-demo
go build  
(RUNTIME_ENV=prod ./go-demo &> /dev/null &)

# 优雅重启
kill -SIGHUP $(ps aux | grep -v grep | grep go-demo | awk '{print $2}')

# 优雅停止
kill -SIGINT $(ps aux | grep -v grep | grep go-demo | awk '{print $2}')
```