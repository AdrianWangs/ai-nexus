// @Author Adrian.Wang 2024/8/2 下午3:59:00
package main

import (
	"context"
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-common/nacos"
	"github.com/cloudwego/hertz/pkg/common/json"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/genericclient"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/kitex-contrib/registry-nacos/resolver"
	"time"
)

func main() {

	configClient, err := nacos.GetNacosConfigClient()

	if err != nil {
		fmt.Println("err:", err)
		return
	}

	r := resolver.NewNacosResolver(configClient)

	p, err := generic.NewThriftFileProvider("./../../idl/user-service.thrift")

	// 从 provide 中获取到的是一个 channel，从 channel 中获取到的是一个 ServiceDescriptor
	serviceDescriptor := <-p.Provide()

	fmt.Println("serviceDescriptor:", serviceDescriptor)
	fmt.Println("serviceDescriptor.Name:", serviceDescriptor.Name)
	functions := serviceDescriptor.Functions

	go func() {
		for name, function := range functions {
			println("************************" + name + "************************")
			fmt.Println("function.Name:", function.Name)
			request := function.Request
			requestStruct := request.Struct
			fieldsByName := requestStruct.FieldsByName

			for _, field := range fieldsByName {

				fieldType := field.Type
				println("%%%%%%%%%%%%%" + fieldType.Name + "%%%%%%%%%%%%%%%")

				fieldStruct := fieldType.Struct

				fmt.Println("请求结构体名：", fieldStruct.Name)

				queryFields := fieldStruct.FieldsByName

				for name, field := range queryFields {

					fmt.Println("名称:", name)
					fmt.Println("字段名称（可能有别名）:", field.FieldName())
					fmt.Println("类型", field.Type.Name)
					fmt.Println("field.HTTPMapping:", field.HTTPMapping)
					fmt.Println("field.ValueMapping:", field.ValueMapping)

				}

				fmt.Println("fieldStruct.RequiredFields:", fieldStruct.RequiredFields)
				fmt.Println("fieldStruct.DefaultFields:", fieldStruct.DefaultFields)

				fmt.Println("fieldType.IsRequestBase:", fieldType.IsRequestBase)

				println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
			}

			fmt.Println("function.Response:", function.Response)

			println("************************" + name + "************************")

		}
	}()

	if err != nil {
		fmt.Println("err:", err)
		return
	}

	thriftGeneric, err := generic.JSONThriftGeneric(p)

	if err != nil {
		fmt.Println("err:", err)
	}

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
