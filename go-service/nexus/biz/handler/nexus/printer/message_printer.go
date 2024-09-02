// Package printer @Author Adrian.Wang 2024/9/2 15:13:00
// 主要用于开发过程中调试，格式化输出消息
package printer

import (
	"fmt"
	"github.com/openai/openai-go"
)

func PrintMessages(messages []openai.ChatCompletionMessageParamUnion) {

	fmt.Print("=========================================")

	for _, message := range messages {
		openai_message, ok := message.(openai.ChatCompletionMessage)
		if ok {
			fmt.Println(openai_message.Role, ":", openai_message.Content)
			jsonStr, _ := openai_message.MarshalJSON()
			fmt.Println("json:", string(jsonStr))
			fmt.Println("---------------------")
		}
	}

	fmt.Print("=========================================")
}
