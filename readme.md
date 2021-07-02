## Golang

### 技术栈

- 路由   Gin https://github.com/gin-gonic/gin
- 配置   viper https://github.com/spf13/viper
- Mysql GoRose https://github.com/gohouse/gorose
- Redis go-redis  https://github.com/go-redis/redis
- 登录   jwt-go  https://github.com/dgrijalva/jwt-go

### 目录结构

```
- config/               配置
  - config.yaml         公共配置. 环境配置 config_{RUNTIME_ENV}.yaml, dev环境配置不参与版本控制.
- controllers/          控制器
- di                    服务注入
  - services.go         Di注册服务
- routers/              Restful路由
- services/             公共业务逻辑
- utils                 工具包
  - utils.go            工具方法
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

错误日志路径 `/var/log/golang_error.log`

### RESTful

- RESTful指南参考 <a href="https://www.vinaysahni.com/best-practices-for-a-pragmatic-restful-api" target="_blank">Best Practices for Designing a Pragmatic RESTful API</a>

- 流程

  `main.go` -> `routers/` -> `controllers/` [-> `services/`]

    - `routes/` 路由, API版本在此控制, Major[.Minor], 比如 /v1, /v1.1, API出现向下不兼容且旧版仍需继续使用的情况, ~~比如不升级的旧版APP,~~ 新增Minor版本号. 业务出现结构性变化, 新增Major版本号.
    - `controllers/` 用于处理业务, 事务控制尽量放置在这里, 放置在 `services/` 中容易出现事务嵌套的问题.
    - `services/` 用于封装公共的业务逻辑, 为可选.
    