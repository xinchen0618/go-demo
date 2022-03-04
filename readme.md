## Golang

### 技术栈

- API           Gin          https://github.com/gin-gonic/gin
- MySQL         GoRose       https://github.com/gohouse/gorose
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
    - db.go             db服务
    - logger.go         日志服务
    - queue.go          消息队列服务
    - redis.go          redis服务
    - worker_pool.go    Goroutine池服务
  - config.go           配置实现
  - common.go           公共配置
  - prod.go             生产环境配置
  - testing.go          测试环境配置
- internal/             内部应用代码
  - action/             命令行action
  - cron/               计划任务  
  - controller/         API控制器
  - router/             API路由
  - middleware/         API中间件  
  - task/               消息队列任务 
  - service/            公共业务逻辑
    - cache.go          资源缓存
    - queue.go          消息队列 
- pkg/                  外部应用可以使用的库代码
  - dbx/                db操作封装. MySQL增删改查操作封装
  - ginx/               gin增强方法. 此包中出现error会向客户端输出4xx/500错误, 调用时捕获到error直接结束业务逻辑即可
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
  
  环境配置 <RUNTIME_ENV>.go, 同键名环境配置覆盖公共配置. dev环境配置不参与版本控制.

- 使用

  `config.Get()`, `config.GetInt()`, `config.GetString()`, `config.GetBool()`, `config.GetIntSlice()`, `config.GetStringSlice()`

### 日志

- 记录日志

  `zap.L(),Error()`, `zap.L().Warn()`, `zap.L().Info()`, `zap.L().Debug()`或者`di.Logger().Error()`, `di.Logger().Warn()`, `di.Logger().Info()`, `di.Logger().Debug()`. 错误日志会记录栈信息.

  日志文件路径通过`config/`中`error_log`项配置, 注意文件需要读写权限. 

- SQL日志

  SQL日志文件路径通过`config/`中`sql_log`项配置, 缺省或为空时不记录日志, 注意文件需要读写权限.

### WorkerPool 

使用WorkerPool(Goroutine池)旨在解决两个问题 

- Goroutine使用资源上限 
- 优雅处理Goroutine中panic
 
#### 使用

- 公共Goroutine池

  ```
  # go func
  for i := 0; i < 10; i++ {
    di.WorkerPool().Submit(func)
  }
  
  # Wait Group
  wpg := di.WorkerPool().Group()
  for i := 0; i < 10; i++ {
    wpg.Submit(func)
  }
  wpg.Wait()
  ```

- 独享Goroutine池

  独享Goroutine池通常起到类似分页处理或者限流的作用  

  ```
  # go func
  wps := di.WorkerPoolSeparate(100)
  for i := 0; i < 10000; i++ {
    wps.Submit(func)
  }
  
  # Wait Group
  wpsg := di.WorkerPool(100).Group()
  for i := 0; i < 10000; i++ {
    wpsg.Submit(func)
  }
  wpsg.Wait()  
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
  - 以 `jwt:user:<userId>:<JwtSignature>`的格式记录入Redis白名单
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

  实际上会提前编译好, 在机器上直接部署可执行文件. 若使用Docker并且没有CGO的需求, 使用`scratch`或`alpine`作为底层镜像, 编译时注意设置`CGO_ENABLED=0`. 

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

  默认队列: 及时消息`service.Queue.Enqueue()`, 延时消息`service.Queue.EnqueueIn()`; 低优先级队列: 及时消息`service.Queue.LowEnqueue()`, 延时消息`service.Queue.LowEnqueueIn()`

### MySQL

`dbx/` 提供以`map[string]interface{}`类型操作和读取数据库的封装 

#### 数据类型映射

- 读操作, MySQL到Golang做统一数据类型映射, 方便操作. MySQL整型(包括无符号)统一映射为`int64`, 浮点型统一映射为`float64`, 其他类型统一映射为`string` 

```
MySQL => Golang 数据类型映射:
    bigint/int/smallint/tinyint => int64,
    float/double => float64,
    varchar/char/longtext/text/mediumtext/tinytext/decimal/datetime/timestamp/date/time => string,
```

- 写操作, Golang到MySQL对数据类型没有强制要求

#### 操作封装

- `FetchAll()` 获取多行记录
- `FetchOne()` 获取一行记录
- `FetchValue()` 获取一个值
- `FetchColumn()` 获取一列值
- `Slice2in()` Slice转IN条件
- `Insert()` 新增记录
- `Update()` 更新记录
- `Delete()` 删除记录
- `Execute()` 执行原生SQL
- `Begin()` 手动开始事务. 事务应优先考虑Transaction()闭包操作是否会更加方便
- `Commit()` 手动提交事务
- `Rollback()` 手动回滚事务

### Redis

key统一在`config/consts/redis_key.go`中定义.

#### 规范

<a href="https://developer.aliyun.com/article/531067" target="_blank">阿里云Redis开发规范</a>

#### 缓存

- 资源缓存

  以资源对象为单位, 使用旁路缓存策略.

  - `GET`资源时`service.Cache.Get()`获取缓存(缓存不存在时会建立)
  - `PUT`/`DELETE`资源时`service.Cache.Delete()`删除缓存

- 业务缓存

  - 自定义缓存, `service.Cache.GetOrSet()`获取或设置自定义缓存
  - 针对API业务设计的缓存, `ginx.GetOrSetCache()`获取或设置API业务缓存