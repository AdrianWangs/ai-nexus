// Package parser @Author Adrian.Wang 2024/8/27 下午4:29:00
// 这个文件主要包含一些将 thrift 文件解析为 openai 的 Functions 的方法
package parser

import (
	"github.com/openai/openai-go"
)

// parseThriftIdlFromPath 从路径解析 thrift 文件
func parseThriftIdlFromPath(idlPath string) (res []openai.ChatCompletionToolParam, err error) {

	// 调用通用解析方法来解析 idl 文件
	return parseIdlFromPath(idlPath), nil
}
