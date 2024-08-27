// @Author Adrian.Wang 2024/8/27 下午5:45:00
package parser

import (
	"testing"
)

func TestThriftParser(t *testing.T) {

	// 测试解析 thrift 文件
	res, err := parseThriftIdlFromPath("./../../../resources/idl/user-service.thrift")
	if err != nil {
		t.Error(err)
		t.Error("解析 thrift 文件失败")
	}
	if res == nil {
		t.Error("解析结果为空")
	}
	t.Log("解析结果:", res)
}
