## Golang

### 技术栈

- API           Gin          https://github.com/gin-gonic/gin
- Mysql         GoRose       https://github.com/gohouse/gorose
- Redis         go-redis     https://github.com/go-redis/redis
- 登录          jwt-go        https://github.com/dgrijalva/jwt-go
- 日志          zap           https://github.com/uber-go/zap
- 优雅停止      endless       https://github.com/fvbock/endless
- 命令行         urfave/cli   https://github.com/urfave/cli
- 计划任务        gocron       https://github.com/go-co-op/gocron
- WorkerPool    pond          https://github.com/alitto/pond
- 消息队列       Asynq         https://github.com/hibiken/asynq


###  规范

- 项目布局 
  
  <a href="https://github.com/golang-standards/project-layout" target="_blank">Standard Go Project Layout</a>
  
- 编码规范 
  
  <a href="https://github.com/xxjwxc/uber_go_guide_cn" target="_blank">Uber Go 语言编码规范</a>


### 目录结构

```
- cmd/                  项目入口
  - demo-api/           API   
  - demo-cli/           命令行
  - demo-cron/          计划任务
  - demo-queue/         消息队列
- config/               配置
  - consts/             常量定义
    - redis_key.go      Redis key统一在此定义避免冲突
  - di/                 服务注入
    - logger.go         日志服务
    - worker_pool.go    goroutine池服务
    - queue.go          消息队列服务
  - config.go           配置实现
  - config_common.go    公共配置
  - config_prod.go      生产环境配置
  - config_testing.go   测试环境配置
- internal/             内部应用代码
  - action/             命令行action
  - cron/               计划任务  
  - controller/         API控制器
  - router/             API路由
    - router.go         路由注册入口. 路由声明按业务分拆到不同文件, 然后统一在此注册.
  - middleware/         API中间件  
  - task/               消息队列任务 
  - service/            公共业务逻辑
    - cache_service.go  资源缓存
    - queue_service.go  消息队列 
- pkg/                  外部应用可以使用的库代码
  - ginx/               gin增强方法. 此包中出现error会向客户端返回4xx/500错误, 调用时捕获到error直接结束业务逻辑即可.
  - gox/                golang增强方法
- go.mod                包管理  
```


### 环境定义

环境变量 `RUNTIME_ENV` 指定执行环境. 默认为生产环境. 参考 <a href="https://en.wikipedia.org/wiki/Deployment_environment" target="_blank">Deployment environment</a>

- `dev`       开发环境. 开发人员的个人环境.
- `testing`   测试环境
- `stage`     预发布环境
- `prod`      生产环境


### 配置

- 为什么放弃使用`Viper`

  为了仅部署编译后的可执行文件, 就可直接运行, 不受YAML、TOML等配置文件位置的制约.

- 多环境配置
  
  环境配置 config_<RUNTIME_ENV>.go, 同键名环境配置覆盖公共配置. dev环境配置不参与版本控制.

- 使用

  `config.Get()`, `config.GetInt()`, `config.GetString()`, `config.GetIntSlice()`, `config.GetStringSlice()`


### 日志

- 日志文件

  **错误日志**会记录到日志文件, 同时打印到console. 错误日志文件路径在`config/`中配置, 默认为`/var/log/golang_error.log`. 注意文件要有读写权限.

- 使用

  `zap.L().Error()`, `zap.L().Warn()`, `zap.L().Info()`


### WorkerPool 

使用WorkerPool(Goroutine池)旨在解决两个问题 

- Goroutine使用资源上限 
- 优雅处理Goroutine中panic
 
#### 使用

- go func

  ```
  di.WorkerPool().Submit(func)
  ```

- Wait Group
  
  ```
  wpg := di.WorkerPool().Group()
  wpg.Submit(func)
  wpg.Wait()
  ```


### API

#### 规范

遵循RESTful规范, 参考指南<a href="https://www.vinaysahni.com/best-practices-for-a-pragmatic-restful-api" target="_blank">Best Practices for Designing a Pragmatic RESTful API</a>

#### 流程

`cmd/demo-api/main.go` -> `internal/router/` [-> `internal/middleware/`] -> `internal/controller/` [-> `internal/service/`]

- `internal/router/` 路由, API版本在此控制, Major[.Minor], 比如 /v1, /v1.1, API出现向下不兼容且旧版仍需继续使用的情况, ~~比如不升级的旧版APP,~~ 新增Minor版本号. 业务出现结构性变化, 新增Major版本号.
- `internal/middleware/` 中间件, 可选.
- `internal/controller/` 业务处理, 事务控制尽量放置在这里, 放置在 `internal/service/` 中容易出现事务嵌套的问题.
- `internal/service/` 公共业务逻辑封装, 可选.
  
#### 登录

- 登录流程

  - 校验账户信息
  - 生成JWT Token
  - 以 `jwt:<userId>:<JWT Signature>`的格式记录入Redis白名单
  - JWT Token返回给客户端

- 校验登录

  - 客户端请求时Header携带JWT Token `Authorization: Bearer <token>`
  - 校验JWT Token
  - 校验Redis白名单
  
- 退出登录
 
  - 校验登录
  - 删除对应Redis白名单

#### 运行

- 开发&测试环境使用gowatch实时热重载

  注意, 是否配置了Go mod代理 `go env -w GOPROXY=https://goproxy.cn,direct`, 是否安装了gowatch `go get github.com/silenceper/gowatch`, 是否配置了Go bin路径 `export PATH=$PATH:$HOME/go/bin`.

  ```
  cd cmd/demo-api
  RUNTIME_ENV=testing gowatch
  ```

- 预发布&生产环境执行编译好的程序

  实际上会提前编译好, 直接将可执行文件部署到机器上, 使用supervisor执行.

  ```
  # 启动
  cd cmd/demo-api
  go build  
  (RUNTIME_ENV=prod ./demo-api &> /dev/null &)

  # 优雅重启
  kill -SIGHUP $(ps aux | grep -v grep | grep demo-api | awk '{print $2}')

  # 优雅停止
  kill -SIGINT $(ps aux | grep -v grep | grep demo-api | awk '{print $2}')
  ```


### Cli

#### 流程

`cmd/demo-cli/main.go` -> `internal/action/` [-> `internal/service/`]

- `cmd/demo-cli/main.go` 定义Cli路由, 按业务维度分两级.
- `internal/action/` 执行逻辑.

#### 使用

```
cd cmd/demo-cli
go build
RUNTIME_ENV=testing ./demo-cli <task> <action> [param]
```


### Cron

#### 流程

`cmd/demo-cron/main.go` -> `internal/cron/` [-> `internal/service/`]  

 - `cmd/demo-cron/main.go` 定义计划任务.
 - `internal/cron/` 执行逻辑.

#### 启动

```
cd cmd/demo-cron
go build
(RUNTIME_ENV=testing ./demo-cron &> /dev/nul &)
```


### Queue

#### 流程

`cmd/demo-queue/main.go` -> `internal/task/` [-> `internal/service/`]

- `cmd/demo-queue/main.go` 定义队列任务.
- `internal/task/` 执行逻辑.

#### 使用

- 启动Worker

  ```
  cd cmd/demo-queue
  go build
  (RUNTIME_ENV=testing ./demo-queue &> /dev/nul &)
  ```

- 优雅停止Worker

  ```
  kill -TERM $(ps aux | grep -v grep | grep demo-queue | awk '{print $2}')
  ```

- 发送Job

  消息队列按任务优先级分两个队列: 默认队列, 该队列分配了较多的系统资源, 任务一般发送至此队列; 低优先级队列, 该队列分配了较少的系统资源, 数据量大优先级低的任务发送至此队列

  默认队列: `service.QueueService.Enqueue()`, `service.QueueService.EnqueueIn()`; 低优先级队列: `service.QueueService.LowEnqueue()`, `service.QueueService.LowEnqueueIn()`


### Redis

#### 规范

<a href="https://developer.aliyun.com/article/531067" target="_blank">阿里云Redis开发规范</a>

#### 缓存

key统一在`config/consts/redis_key.go`中定义.

- 资源缓存

  以资源对象为单位, 使用旁路缓存策略.

  - `GET`资源时`service.CacheService.Get()`获取缓存(缓存不存在时会建立)
  - `PUT`/`DELETE`资源时`service.CacheService.Delete()`删除缓存

- 业务缓存

  针对业务设计的缓存, `service.CacheService.GetOrSet()`获取或设置业务缓存. 
