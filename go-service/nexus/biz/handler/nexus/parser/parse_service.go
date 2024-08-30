// Package parser @Author Adrian.Wang 2024/8/30 20:02:00
package parser

import (
	"bufio"
	"fmt"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/openai/openai-go"
	"os"
	"strings"
)

// extractServiceName 用于提取服务名称
func extractServiceName(path string) string {
	// Step 1: Split the path by "/"
	parts := strings.Split(path, "/")
	// Step 2: Get the last part
	fileName := parts[len(parts)-1]
	// Step 3: Split the file name by "."
	nameParts := strings.Split(fileName, ".")
	// Step 4: Get the first part
	serviceName := nameParts[0]
	return serviceName
}

// readFirstCommentLines 读取文件中最前面的以 // 开头的行
func readFirstCommentLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commentLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // 跳过空白行
		}
		if strings.HasPrefix(line, "//") {
			// 去掉开头的//
			line = strings.TrimPrefix(line, "//")
			commentLines = append(commentLines, line)
		} else {
			break // 遇到不以 // 开头且不是空白的行，停止读取
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commentLines, nil
}

func parseServiceFromPath(dir []string) (res []openai.ChatCompletionToolParam) {

	// 创建参数
	res = make([]openai.ChatCompletionToolParam, 0)

	// 只有服务所以是空参数，直接实例化一个就行了
	params := make(openai.FunctionParameters)

	for _, thriftFile := range dir {

		serviceName := extractServiceName(thriftFile)
		klog.Infof("serviceName: %s", serviceName)

		// 读取文件中的第一行
		descriptions, err := readFirstCommentLines(thriftFile)

		// 将所有开头的注释都拼接起来，就是该文件的解析
		description := strings.Join(descriptions, ",")

		if err != nil {
			fmt.Println(err)
		}

		toolParam := openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.String(serviceName), // 服务名称
				Description: openai.String(description), // 方法描述,通过 thrift 文件的注释获取
				Parameters:  openai.F(params),
			}),
		}

		res = append(res, toolParam)
	}

	return
}
