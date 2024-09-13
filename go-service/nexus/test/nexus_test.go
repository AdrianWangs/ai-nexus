// @Author Adrian.Wang 2024/8/30 19:50:00
package test

import (
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/handler/nexus"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kr/pretty"
	"os"
	"testing"
)

func TestGetServicesFromThrift(t *testing.T) {

	/** 设置当前路径为项目根目录 */
	if err := os.Chdir("../"); err != nil {
		klog.Error("设置当前路径为项目根目录失败")
		os.Exit(1)
	}

	res := nexus.GetServicesFromThrift()
	pretty.Println(res)
}
