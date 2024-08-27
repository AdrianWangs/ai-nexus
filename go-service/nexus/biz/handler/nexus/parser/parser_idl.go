// Package parser @Author Adrian.Wang 2024/8/27 下午5:38:00
package parser

import (
	"github.com/AdrianWangs/ai-nexus/go-common/idlParser"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kr/pretty"
	"github.com/openai/openai-go"
)

// parseIdlFromPath 使用路径解析 idl 文件
func parseIdlFromPath(idlPath string, includeDirs ...string) (res []openai.ChatCompletionToolParam) {

	var p idlParser.IdlParser
	p = idlParser.NewThriftIdlParser()

	_, err := p.ParseIdlFromPath(idlPath, includeDirs...)

	if err != nil {
		klog.Info("err:", err)
		res = nil
		return
	}

	return nil

	//description, err := parser.ParseIdlFromPath(idlPath, includeDirs...)
	//
	//if err != nil {
	//	klog.Info("err:", err)
	//	res = nil
	//	return
	//}
	//
	//return parseIdlFromDescription(description)

}

// parseIdlFromDescription 从描述中解析 idl 文件
func parseIdlFromDescription(description interface{}) (res []openai.ChatCompletionToolParam) {
	pretty.Println(description)

	return nil

}
