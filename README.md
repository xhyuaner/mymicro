# MyMicro

### 1 介绍 **🚩**

**MyMicro** 是一个基于 **gRPC** 和 **Gin** 封装的基础微服务框架，并且集成了 **OpenTelemetry** 来实现对各项服务的链路追踪，便于快速排错

### 2 功能 🔎

- 支持 RPC/Http 协议
- 服务注册发现
- 服务负载均衡
- 配置集中式管理
- 集成 RPC 服务的多种拦截器
- 支持 Basic、Cache 和 JWT 权限认证

### 3 优势💡

- 遵从三层代码结构规范，降低代码耦合度
- 全新日志包设计与统一错误码管理
- 通过 AST 语法树自动解析并生成代码
- 组件化的设计模式，用接口对功能进行抽象，实现组件可插拔
- 运用函数选项模式，配置化驱动组件
- 保证业务方能够以统一的调用方式启动各种服务
- 实现微服务的可观测、可治理

### 4 快速开始 ✒️

```shell
// 安装项目脚手架
go get -u github.com/xhyuaner/mymicro
go install github.com/xhyuaner/mymicro/tools

// 修改服务配置和路由配置
......

// 启动服务
go run cmd/my-service/main.go
```



