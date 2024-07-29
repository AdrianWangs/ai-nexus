> ## 更新计划
> - [ ] go 用户微服务的编写
> - [ ] go 用户部署的配置文件
> - [ ] go 配置中心的代码编写
> - [ ] 编写thrift模板代码来配置自己的模板文件
> - [ ] java demo 的编写

# docker-compose 本地开发环境部署

## 一、部署环境
> 使用 dockcer-compose 部署本地开发环境
> ```bash
> docker-compose up -d
> ```

## 二、运行服务
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