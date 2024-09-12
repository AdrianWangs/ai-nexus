// Package function_call
// @Author Adrian.Wang 2024/9/12 18:54:00
// 泛化调用主要负责对于微服务进行泛化调用
package function_call

import (
	"context"
	"fmt"
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

	microServiceName := splitName[0]
	//serviceName := splitName[1]
	functionName = splitName[2]

	// 本地文件 idl 解析
	// YOUR_IDL_PATH thrift 文件路径: 举例 ./idl/example.thrift
	// includeDirs: 指定 include 路径，默认用当前文件的相对路径寻找 include
	p, err := generic.NewThriftFileProvider(fmt.Sprintf("./resources/idl/%s.thrift", microServiceName))
	if err != nil {
		klog.Error("调用失败:", err.Error())
		return "调用失败:" + err.Error()
	}

	// 构造 map 类型的泛化调用
	g, err := generic.MapThriftGeneric(p)
	if err != nil {
		klog.Error("调用失败:", err.Error())
		return "调用失败:" + err.Error()
	}
	cli, err := genericclient.NewClient(microServiceName, g)
	if err != nil {
		klog.Error("调用失败:", err.Error())
		return "调用失败:" + err.Error()
	}
	// 'ExampleMethod' 方法名必须包含在 idl 定义中
	// resp 类型为 map[string]interface{}
	resp, err := cli.GenericCall(context.Background(), functionName, params)

	return pretty.Sprint(resp)
}
