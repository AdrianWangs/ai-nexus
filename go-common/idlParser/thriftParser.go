// Package idlParser @Author Adrian.Wang 2024/8/27 下午7:56:00
// 专门用来解析 idl 文件的包
package idlParser

import (
	"github.com/cloudwego/thriftgo/parser"
)

type ThriftIdlParser struct{}

func NewThriftIdlParser() *ThriftIdlParser {
	return &ThriftIdlParser{}
}

func (tp *ThriftIdlParser) ParseIdlFromPath(idlPath string, includeDirs ...string) (*parser.Thrift, error) {
	tree, err := parser.ParseFile(idlPath, includeDirs, true)
	if err != nil {
		return nil, err
	}
	return tree, nil
}
