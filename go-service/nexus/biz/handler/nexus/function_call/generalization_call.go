// Package function_call
// @Author Adrian.Wang 2024/9/12 18:54:00
// 泛化调用主要负责对于微服务进行泛化调用
package function_call

import (
	"context"
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-common/nacos"
	"github.com/cloudwego/hertz/pkg/common/json"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kr/pretty"
	"strings"
)

func GeneralizationCall(functionName string, params string) string {

	// 将functionName 使用-分割，正常会有三个部分，第一个是服务名称（nacos）
	// 第二个是服务名称（thrift 的名称），第三个是方法名称

	splitName := strings.Split(functionName, "-")

	// 用来组合，thrift 的服务名称和方法名称不可能包含-，所以最后两个一定是
	splitNum := len(splitName)

	microServiceName := ""
	serviceName := ""
	functionName = ""

	for i := 0; i < splitNum-2; i++ {
		microServiceName += splitName[i] + "-"
	}
	microServiceName = microServiceName[:len(microServiceName)-1]
	serviceName = splitName[splitNum-2]
	functionName = splitName[splitNum-1]

	klog.Debug("microServiceName:", microServiceName)
	klog.Debug("serviceName:", serviceName)
	klog.Debug("functionName:", functionName)

	// 本地文件 idl 解析
	// YOUR_IDL_PATH thrift 文件路径: 举例 ./idl/example.thrift
	// includeDirs: 指定 include 路径，默认用当前文件的相对路径寻找 include
	p, err := generic.NewThriftFileProvider(fmt.Sprintf("./resources/idl/%s.thrift", microServiceName))
	if err != nil {
		klog.Error("调用失败:", err.Error())
		return "调用失败:" + err.Error()
	}

	// 构造 map 类型的泛化调用
	g, err := generic.JSONThriftGeneric(p)
	if err != nil {
		klog.Error("调用失败:", err.Error())
		return "调用失败:" + err.Error()
	}

	// nacos 注册中心
	r := nacos.GetNacosResolver()

	cli, err := genericclient.NewClient(microServiceName, g, client.WithResolver(r))
	if err != nil {
		klog.Error("调用失败:", err.Error())
		return "调用失败:" + err.Error()
	}

	// 将 params 转化为 map 类型
	paramMap := make(map[string]interface{})

	err = json.Unmarshal([]byte(params), &paramMap)
	if err != nil {
		klog.Error("调用失败:", err.Error())
		return "调用失败:" + err.Error()
	}

	// 取出 map 中的第一个参数
	var request interface{}
	for _, v := range paramMap {
		request = v
		break
	}

	// 将 request 转化json
	req, err := json.Marshal(request)

	if err != nil {
		klog.Error("调用失败:", err.Error())
		return "调用失败:" + err.Error()
	}

	fmt.Println("req:", string(req))

	// 调用函数
	resp, err := cli.GenericCall(context.Background(), functionName, string(req))

	if err != nil {
		klog.Error("调用失败:", err.Error())
		return "调用失败:" + err.Error()
	}

	return pretty.Sprint(resp)
}
