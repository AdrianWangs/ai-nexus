package main

import (
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	client := openai.NewClient(
		option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1/"),
		option.WithAPIKey("sk-8285fe317edc44ef95f029be9b7cfe94"), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Say this is a test"),
		}),
		Model: openai.F("qwen-plus-0806"),
	})
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(chatCompletion)

}
