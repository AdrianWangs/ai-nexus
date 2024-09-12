// Package parser @Author Adrian.Wang 2024/8/27 下午5:38:00
package parser

import (
	"github.com/AdrianWangs/ai-nexus/go-common/idlParser"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/kr/pretty"
	"github.com/openai/openai-go"
)

// parseIdlFromPath 使用路径解析 idl 文件
func parseIdlFromPath(idlPath string, includeDirs ...string) (res []openai.ChatCompletionToolParam) {

	var p idlParser.IdlParser
	p = idlParser.NewThriftIdlParser()

	thrift, err := p.ParseIdlFromPath(idlPath, includeDirs...)

	if err != nil {
		klog.Info("err:", err)
		res = nil
		return
	}

	return parseIdlFromThrift(thrift)

}

// parseIdlFromThrift 使用 thrift 解析 idl 文件
/**
返回结构大概如下：
[
        {
            "function": {
                "description": "用于安排计划清单",
                "name": "make_plan",
                "parameters": {
                    "properties": {
						"base_example": {
                            "description": "这是一个基本数据类型的例子",
                            "type": "string"
                        },
						"object_example": {
                            "description": "这是一个对象数据类型的例子'",
                            "type": "object",
                            "properties": {
                               "base_example": {
									"description": "这是一个基本数据类型的例子",
									"type": "string"
								},...
                            }
                        },
                        "list_example": {
                            "description": "这是一个列表类型的例子'",
                            "items": {
								"description": "这是一个对象数据类型的例子'",
								"type": "object",
								"properties": {
								   "base_example": {
										"description": "这是一个基本数据类型的例子",
										"type": "string"
									},...
								}
							},
                            "type": "array"
                        },
                    },
                    "required": [
                        "location"
                    ],
                    "type": "object"
                }
            },
            "type": "function"
        }
    ]
*/
func parseIdlFromThrift(thrift *parser.Thrift) (res []openai.ChatCompletionToolParam) {

	// 结构体字典，因为结构体之间有相互引用关系，所以存放在字典里方便解析结构体类型
	structsMap := struct2Map(thrift.Structs)

	// 创建数组来存放解析结果
	res = make([]openai.ChatCompletionToolParam, 0)

	// 服务列表
	services := thrift.Services

	for _, service := range services {

		// 服务名称
		//fmt.Println("\t|", service.Name)

		// 服务方法列表
		functions := service.Functions

		for _, function := range functions {

			// 方法名称
			functionName := service.Name + "-" + function.Name

			// 方法参数列表
			arguments := function.Arguments

			// 将参数解析成 openai 格式的参数列表
			params, err := arguments2OpenaiParams(arguments, structsMap)

			if err != nil {
				klog.Error("解析参数列表失败")
				return nil
			}

			// openai 格式的参数列表
			toolParam := openai.ChatCompletionToolParam{
				Type: openai.F(openai.ChatCompletionToolTypeFunction),
				Function: openai.F(openai.FunctionDefinitionParam{
					Name:        openai.String(functionName),              // 方法名称
					Description: openai.String(function.ReservedComments), // 方法描述,通过 thrift 文件的注释获取
					Parameters:  openai.F(params),
				}),
			}

			// 将解析结果添加到结果列表
			res = append(res, toolParam)

		}

	}

	return res

}

// arguments2OpenaiParams 将 thrift 的参数列表转换成 openai 的参数列表
// 说白了就是解析字段类型
/**
这是解析出来的例子：
 {
		"properties": {
			"base_example": {
				"description": "这是一个基本数据类型的例子",
				"type": "string"
			},
			"object_example": {
				"description": "这是一个对象数据类型的例子'",
				"type": "object",
				"properties": {
				   "base_example": {
						"description": "这是一个基本数据类型的例子",
						"type": "string"
					},...
				}
			},
			"list_example": {
				"description": "这是一个列表类型的例子'",
				"items": {
					"description": "这是一个对象数据类型的例子'",
					"type": "object",
					"properties": {
					   "base_example": {
							"description": "这是一个基本数据类型的例子",
							"type": "string"
						},...
					}
				},
				"type": "array"
			},
		},
		"required": [
			"location"
		],
		"type": "object"
	}
}
*/
func arguments2OpenaiParams(arguments []*parser.Field, structsMap map[string]*parser.StructLike) (openai.FunctionParameters, error) {

	required := make([]string, 0)

	params := make(map[string]interface{})
	params["type"] = "object"

	// 参数清单
	properties := make(map[string]interface{})

	for _, argument := range arguments {

		property, err := argument2OpenaiProperty(argument.Type, argument.ReservedComments, structsMap)

		if err != nil {
			klog.Error("解析参数失败")
			return nil, err
		}

		properties[argument.Name] = property

		// 如果字段必须
		if argument.Requiredness == parser.FieldType_Required {
			required = append(required, argument.Name)
		}
	}

	// 必须字段
	params["required"] = required

	// 参数列表
	params["properties"] = properties

	return params, nil

}

// argument2OpenaiProperty 将 thrift 的参数转换成 openai 的参数
/**
这是解析出来的三个例子：
1. 基本数据类型：
 {
	"description": "这是一个基本数据类型的例子",
	"type": "string"
},
2. 对象类型：
{
	"description": "这是一个对象数据类型的例子'",
	"type": "object",
	"properties": {
	   "base_example": {
			"description": "这是一个基本数据类型的例子",
			"type": "string"
		},...
	}
},
3. 列表类型：
{
	"description": "这是一个列表类型的例子'",
	"items": {
		"description": "这是一个对象数据类型的例子'",
		"type": "object",
		"properties": {
		   "base_example": {
				"description": "这是一个基本数据类型的例子",
				"type": "string"
			},...
		}
	},
	"type": "array"
}
设置另一个例子：

*/
func argument2OpenaiProperty(argument *parser.Type, comments string, structsMap map[string]*parser.StructLike) (map[string]interface{}, error) {

	property := map[string]interface{}{}

	// 直接使用注释作为描述符
	property["description"] = comments

	property["type"] = argumentType2openaiType(argument.Name)

	//如果是对象类型，则还需要
	if property["type"] == "object" {
		properties, err := struct2OpenaiProperty(structsMap[argument.Name], structsMap)

		if err != nil {
			klog.Error("解析参数列表(object)失败")
			return nil, err
		}
		property["properties"] = properties
	}

	if property["type"] == "array" {

		items, err := argument2OpenaiProperty(argument.ValueType, "", structsMap)

		if err != nil {
			klog.Error("解析参数列表(array)失败")
			return nil, err
		}

		property["items"] = items

	}

	return property, nil
}

// struct2OpenaiProperty 将结构体转换成 openai 的对象参数
/**
这是解析出来的一个例子：
{
   "base_example": {
		"description": "这是一个基本数据类型的例子",
		"type": "string"
	}
}
*/
func struct2OpenaiProperty(structLike *parser.StructLike, structsMap map[string]*parser.StructLike) (map[string]interface{}, error) {

	// 参数列表
	properties := map[string]interface{}{}

	// 遍历转化参数列表
	for _, argument := range structLike.Fields {

		property, err := argument2OpenaiProperty(argument.Type, argument.ReservedComments, structsMap)

		if err != nil {
			klog.Error("解析参数失败")
			return nil, err
		}

		properties[argument.Name] = property

	}

	return properties, nil

}

// struct2Map 将结构体列表转化成 map，方便查找
func struct2Map(structs []*parser.StructLike) map[string]*parser.StructLike {

	structsMap := make(map[string]*parser.StructLike)

	for _, structLike := range structs {

		if structsMap[structLike.Name] != nil {
			continue
		}

		structsMap[structLike.Name] = structLike

		/**
		fmt.Println()

		// 字段解释
		fmt.Println("\t|", structLike.ReservedComments)
		// 结构体名称
		fmt.Println("\t|", structLike.Name)

		// 结构体字段
		fields := structLike.Fields

		for _, field := range fields {
			fmt.Println()

			// 字段解释
			fmt.Println("\t\t|", field.ReservedComments)
			// 字段 id
			fmt.Println("\t\t|", field.ID)
			// 字段名称
			fmt.Println("\t\t|", field.Name)
			// 字段是否必须
			fmt.Println("\t\t|", field.Requiredness)
			// 字段类型
			fmt.Println("\t\t|", field.Type)
		}
		**/

	}

	return structsMap
}

func argumentType2openaiType(argumentType string) (openaiType string) {
	switch argumentType {
	case "bool":
		openaiType = "boolean"
	case "byte":
	case "i16":
	case "i32":
	case "i64":
		openaiType = "integer"
	case "double":
		openaiType = "number"
	case "string":
		openaiType = "string"
	case "list":
		openaiType = "array"
	default:
		openaiType = "object"
	}

	return
}

// Deprecated: 默认的 description  不包含解释(无法解释注释和注解)，属于一个不完整的 thrift 解析方法，因此废弃
// parseIdlFromDescription 从描述中解析 idl 文件
func parseIdlFromDescription(description interface{}) (res []openai.ChatCompletionToolParam) {
	pretty.Println(description)

	return nil

}
