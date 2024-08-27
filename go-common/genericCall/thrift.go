// @Author Adrian.Wang 2024/8/6 下午5:33:00
package genericCall

import (
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/generic/descriptor"
)

type ThriftIdlParser struct{}

// ParseGeneralFunction 将 idl 文件转化为泛化调用的参数结构体
func (tp *ThriftIdlParser) ParseGeneralFunction(idlProvider generic.DescriptorProvider) (description interface{}, err error) {

	// 从 provide 中获取到的是一个 channel，从 channel 中获取到的是一个 ServiceDescriptor
	serviceDescriptor := <-idlProvider.Provide()

	functions := serviceDescriptor.Functions

	var functionMap = make(map[string]interface{})

	for name, function := range functions {

		// 解析函数的参数，得到 rpc 调用的出入参数
		functionMap[name] = tp.getParams(function)

	}

	_ = idlProvider.Close()

	return functionMap, nil
}

// getParams 用于获取请求和响应的字段
func (tp *ThriftIdlParser) getParams(functionDescriptor *descriptor.FunctionDescriptor) map[string]interface{} {

	// 获取请求和响应的字段列表的对象
	requestFields := functionDescriptor.Request.Struct.FieldsByName
	responseFields := functionDescriptor.Response.Struct.FieldsByName

	// 创建一个 map 用于存储请求和响应的字段
	params := make(map[string]interface{})

	params["Request"] = tp.getParamsFromField(REQUEST, requestFields)
	params["Response"] = tp.getParamsFromField(RESPONSE, responseFields)

	return params
}

// getParamsFromField 用于获取字段的详细信息
func (tp *ThriftIdlParser) getParamsFromField(dataType DataType, fieldsByName map[string]*descriptor.FieldDescriptor) map[string]interface{} {

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
			fieldDetail = tp.getParamsFromField(DEFAULT, field.Type.Struct.FieldsByName)
		}

		fieldParams[name] = fieldDetail

	}

	return fieldParams

}
