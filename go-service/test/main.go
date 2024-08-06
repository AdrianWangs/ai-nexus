// @Author Adrian.Wang 2024/8/2 下午3:59:00
package main

import (
	"context"
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-common/genericCall"
	"github.com/AdrianWangs/ai-nexus/go-common/nacos"
	"github.com/cloudwego/hertz/pkg/common/json"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/kitex-contrib/registry-nacos/resolver"
	"time"
)

func main() {

	idlPath := "./../../idl/user-service.thrift"

	p, err := generic.NewThriftFileProvider(idlPath)

	if err != nil {
		fmt.Println("err:", err)
		return
	}

	var parser genericCall.IdlParser
	parser = &genericCall.ThriftIdlParser{}

	descriptior, err := parser.ParseGeneralFunction(p)

	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println("解析 idl 文件得到的rpc 调用:")
	// 将 descriptior 转化为 json 字符串
	descriptiorJson, err := json.Marshal(descriptior)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	fmt.Println("descriptior:", string(descriptiorJson))

	p, err = generic.NewThriftFileProvider(idlPath)

	// 获取 nacos 配置中心的客户端
	configClient, err := nacos.GetNacosConfigClient()

	if err != nil {
		fmt.Println("err:", err)
		return
	}

	// 创建 nacos 的服务发现客户端
	r := resolver.NewNacosResolver(configClient)

	// 生成泛化调用的参数
	thriftGeneric, err := generic.JSONThriftGeneric(p)

	if err != nil {
		fmt.Println("err:", err)
	}

	// 创建泛化调用的客户端
	cli, err := genericclient.NewClient(
		"user-service",
		thriftGeneric,
		client.WithResolver(r),
		client.WithRPCTimeout(3*time.Second),
	)

	if err != nil {
		fmt.Println("err:", err)
		return
	}

	req := map[string]interface{}{
		"UserId": int64(1),
	}

	// req 转化为 json 字符串
	reqJson, err := json.Marshal(req)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	fmt.Println("reqJson:", string(reqJson))

	resp, err := cli.GenericCall(context.Background(), "GetUser", "{\"UserId\":1}")

	if err != nil {
		fmt.Println("请求失败")
		fmt.Println("err:", err)
		return
	}

	fmt.Println("resp:", resp)

}
