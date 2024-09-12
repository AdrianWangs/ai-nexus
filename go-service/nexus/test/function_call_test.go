// @Author Adrian.Wang 2024/9/12 19:44:00
package test

import (
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/handler/nexus/function_call"
	"github.com/cloudwego/kitex/pkg/klog"
	"os"
	"testing"
)

func TestFunctionCall(t *testing.T) {

	klog.SetLevel(klog.LevelDebug)

	/** 设置当前路径为项目根目录 */
	if err := os.Chdir("../"); err != nil {
		klog.Error("设置当前路径为项目根目录失败")
		os.Exit(1)
	}

	res := function_call.GeneralizationCall("user-service-UserService-GetUser", `{"request":{"UserId":1}}`)

	fmt.Print(res)

}
