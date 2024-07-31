> ## 更新计划
> - [x] go 配置中心的代码编写
> - [ ] go 用户微服务的编写
>   - [x] 登录
>   - [x] 注册
>   - [x] 用户信息修改
>   - [x] 用户信息查询
>   - [ ] 第三方登录(后期计划)
>   - [x] JWT 鉴权
> - [ ] ai 智核微服务的编写
>   - [ ] 通用 ai 的接口编写
>   - [ ] 反射机制调用函数
>   - [ ] 工作流
> - [ ] 前端的编写
> - [ ] go 用户部署的配置文件
> - [ ] 编写thrift模板代码来配置自己的模板文件
> - [ ] java demo 的编写


# 目录结构
```text
.
├── README.md
├── api-test // 请求接口测试
├── compose.yaml // docker-compose 配置文件
├── go-common // 存放公共的go代码
│   ├── conf // 配置文件
│   ├── error_code // 错误码
│   ├── go.mod // go 依赖包
│   ├── middleware // 中间件
│   │   └── jwt.go // jwt中间件
│   ├── nacos // nacos配置，含配置中心和服务发现
│   └── utils // 工具类
├── go-service // 存放go的微服务
│   ├── run.sh // 一键运行所有微服务
│   └── user // 用户微服务
│       ├── biz //业务逻辑相关的代码，主要修改这里
│       │   ├── dal // 数据访问层，用于初始化数据库和数据库相关的业务逻辑
│       │   ├── handler //（http 相关的）相当于 mvc 中的 controller，用于处理请求和返回响应
│       │   │   └── user_microservice
│       │   │       └── user_service.go
│       │   ├── router //路由相关的代码，用于初始化路由和中间件
│       │   │   ├── register.go
│       │   │   └── user_microservice
│       │   │       ├── middleware.go
│       │   │       └── user-service.go // 这种文件不要修改，因为每次生成代码都会覆盖
│       │   └── service //业务逻辑层，用于处理业务逻辑
│       ├── build.sh // 一键编译当前微服务
│       ├── conf // 配置文件，是单个微服务的配置文件，其中 nacos 相关配置是公共的，不在这里配置
│       │   ├── conf.go
│       │   ├── dev // 开发环境配置
│       │   ├── online // 线上环境配置
│       │   └── test // 测试环境配置
│       ├── docker-compose.yaml // 构建当前微服务所需的docker环境
│       ├── go.mod // go 依赖包
│       ├── handler.go //业务逻辑入口，更新时会覆盖
│       ├── hex_trans_handler.go // 业务逻辑入口，更新时不会覆盖
│       ├── kitex_gen // 生成的代码，这里不要修改
│       ├── main.go // 主函数
│       ├── model // 数据模型
│       ├── readme.md // 当前微服务说明文档
│       ├── script // 脚本文件
├── go.work  //存放go的工作目录
├── idl //存放用于生成代码的thrift文件
├── k8s-config //k8s的配置文件
│   ├── auth.yaml //创建服务账号和角色
│   ├── cluster-config.yaml //创建服务
│   ├── database.yaml //创建数据库
│   ├── ingress.yaml //创建ingress规则
│   └── namespace.yaml //创建命名空间
├── redis // redis 数据存放的目录，用于持久化
└── some-think.md // 项目说明文档，整个项目的构思

```


# docker-compose 本地开发环境部署

## 一、部署环境
> 使用 dockcer-compose 部署本地开发环境
> ```bash
> docker-compose up -d
> ```

## 二、运行服务
### 1. 一键运行（不推荐）
进入到 go-service 目录下，运行如下命令
```bash
bash run.sh
```
这个命令会自动运行 go 的所有微服务，如果需要查看日志，可以使用如下命令
```bash
tail -f  微服务名称/微服务名称.log
```
> 但是这里有个问题，就是每次关闭得手动关闭，使用以下步骤：
> 1. 查看所有相关端口的进程:
> ```bash
> lsof -i:端口号
> ```
> 2. 杀掉进程
> ```bash
> kill -9 进程号
> ```

### 2. 单个运行
进入到 go-service 目录下，这个目录下全部都是微服务的代码，可以进入到对应的微服务目录下，运行如下命令
```bash
go run .
```
> 这个命令会自动运行 go 的所有微服务，如果需要查看日志，可以使用如下命令

# k8s线上负载均衡环境部署

## 一、编写go代码

## 二、使用Dockerfile更新镜像

### 手动构建
1. 将微服务目录的Dockerfile拷贝到根目录下
2. 构建镜像
```bash
docker build -t 微服务名称:版本号 .
```
3. (可选)推送镜像

### 运行脚本
或者其实也可以使用我编写的脚本来自动构建镜像
```bash
sh deploy.sh
```
windows电脑点击deploy.bat即可运行

## 三、使用k8s部署微服务

### 先配置nginx-ingress进行端口转发
1. 安装helm，参考在线教程
2. 安装nginx-ingress
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
```
```bash
helm repo update
```
```bash
helm install ingress-nginx ingress-nginx/ingress-nginx
```
3. 运行`k8s-config/auth.yaml`文件，创建服务账号和角色，以保证服务的pod可以正确被创建
```bash
kubectl create namespace go-zero-demo
```
```bash
kubectl apply -f k8s-config/auth.yaml 
```

4. 运行`k8s-config/cluster-config.yaml`文件，将服务跑起来
> 这里需要注意一下，使用配置中的镜像名称，需要修改配置文件中的镜像名称
```bash
kubectl apply -f k8s-config/cluster-config.yaml
```
5. 运行`k8s-config/ingress.yaml`文件，创建ingress规则，以保证服务可以被外部访问
```bash
kubectl apply -f k8s-config/ingress.yaml
```
> 可以运行如下命令查看服务是否正常运行
```bash
kubectl get pods --all-namespaces
```
> mac 电脑使用 minikube 发现不能访问时，可以使用如下命令查看是否是因为ip地址不对导致的
```bash
minikube tunnel
```