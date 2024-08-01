# 错误记录文档

## go mod tidy 报错包找不到

### 问题描述

当在项目的 module 中执行 `go mod tidy` 时，报错找不到包，但是在项目中确实存在。

报错信息如下：

```shell
go: github.com/AdrianWangs/ai-nexus/go-common/middleware imports
        github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/model: module github.com/AdrianWangs/ai-nexus@latest found (v0.0.0-20240801104342-3285310785ec), but does not contain package github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/model
go: github.com/AdrianWangs/ai-nexus/go-common/middleware imports
        github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/mysql: module github.com/AdrianWangs/ai-nexus@latest found (v0.0.0-20240801104342-3285310785ec), but does not contain package github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/mysql
```

### 报错原因

因为 go mod tidy 会检查项目中的所有依赖包，但是是会去下载最新的包，但是由于我们项目的结构是多个 module，所以下载包一定是找不到的（go 默认将一个项目地址看做一个 module，跟我们的一项目多 module 的结构不一样），他会将仓库的地址当做一整个 module，所以会报错找不到包。

### 解决方法

将项目中每个 module 的 go.mod 文件中添加 replace 项，将项目中的 module 替换为本地的 module。

```text
// 需要替换为本地module 的路径，不然会找不到
replace github.com/AdrianWangs/ai-nexus/go-service/user => ../go-service/user
```