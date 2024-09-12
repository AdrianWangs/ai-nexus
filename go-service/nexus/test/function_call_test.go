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

	/** 设置当前路径为项目根目录 */
	if err := os.Chdir("../"); err != nil {
		klog.Error("设置当前路径为项目根目录失败")
		os.Exit(1)
	}

	res := function_call.GeneralizationCall("plan-ScheduleService-createEvent", `{"event":{"title":"苏州经典旅游行程计划","description":"安排苏州的经典旅游景点","startTime":"2023-10-01T09:00:00","endTime":"2023-10-01T18:00:00","location":"苏州"}}`)

	fmt.Print(res)

}
