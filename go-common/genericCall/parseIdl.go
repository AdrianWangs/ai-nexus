// Package genericCall genericCall @Author Adrian.Wang 2024/8/6 下午5:18:00
package genericCall

import (
	"fmt"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/generic/descriptor"
)

// DataType 用于标识请求和响应的字段,主要用于解析 idl 文件过程中判断字段类型是否为请求或响应
type DataType string

const (
	REQUEST  DataType = "request"
	RESPONSE DataType = ""
	DEFAULT  DataType = "struct"
)

// getParams 用于获取请求和响应的字段
func getParams(functionDescriptor *descriptor.FunctionDescriptor) map[string]interface{} {

	// 获取请求和响应的字段列表的对象
	requestFields := functionDescriptor.Request.Struct.FieldsByName
	responseFields := functionDescriptor.Response.Struct.FieldsByName

	// 创建一个 map 用于存储请求和响应的字段
	params := make(map[string]interface{})

	params["Request"] = getParamsFromField(REQUEST, requestFields)
	params["Response"] = getParamsFromField(RESPONSE, responseFields)

	return params
}

// getParamsFromField 用于获取字段的详细信息
func getParamsFromField(dataType DataType, fieldsByName map[string]*descriptor.FieldDescriptor) map[string]interface{} {

	if dataType != DEFAULT {
		fieldsByName = fieldsByName[string(dataType)].Type.Struct.FieldsByName
	}

	fieldParams := make(map[string]interface{})

	for name, field := range fieldsByName {

		var fieldDetail interface{}

		fieldType := field.Type

		if fieldType == nil {
			continue
		}

		if fieldType.Type != descriptor.STRUCT {
			fieldDetail = field.Type.Name
		} else {
			fieldDetail = getParamsFromField(DEFAULT, field.Type.Struct.FieldsByName)
		}

		fieldParams[name] = fieldDetail

	}

	return fieldParams

}

// ParseGeneralFunction 将 idl 文件转化为泛化调用的参数结构体
func ParseGeneralFunction(idlProvider generic.DescriptorProvider) (description interface{}, err error) {

	// 从 provide 中获取到的是一个 channel，从 channel 中获取到的是一个 ServiceDescriptor
	serviceDescriptor := <-idlProvider.Provide()

	functions := serviceDescriptor.Functions

	var functionMap = make(map[string]interface{})

	for name, function := range functions {

		// 解析函数的参数，得到 rpc 调用的出入参数
		functionMap[name] = getParams(function)

	}

	if err != nil {
		fmt.Println("err:", err)
		return
	}

	_ = idlProvider.Close()

	return functionMap, nil
}
