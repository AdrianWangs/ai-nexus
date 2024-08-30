// @Author Adrian.Wang 2024/8/30 19:50:00
package nexus

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kr/pretty"
	"os"
	"testing"
)

func TestGetServicesFromThrift(t *testing.T) {

	/** 设置当前路径为项目根目录 */
	if err := os.Chdir("../../.."); err != nil {
		klog.Error("设置当前路径为项目根目录失败")
		os.Exit(1)
	}

	res := GetServicesFromThrift()
	pretty.Println(res)
}
