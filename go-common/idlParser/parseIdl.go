// Package genericCall genericCall @Author Adrian.Wang 2024/8/6 下午5:18:00
package idlParser

import (
	"github.com/cloudwego/thriftgo/parser"
)

// DataType 用于标识请求和响应的字段,主要用于解析 idl 文件过程中判断字段类型是否为请求或响应
type DataType string

const (
	REQUEST  DataType = "request"
	RESPONSE DataType = ""
	DEFAULT  DataType = "struct"
)

type IdlParser interface {
	ParseIdlFromPath(idlPath string, includeDirs ...string) (*parser.Thrift, error)
}
