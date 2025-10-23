package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

func TestClient(t *testing.T) {

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置环境变量 OPENAI_API_KEY")
	}

	// --- 客户端配置 ---
	// 可以指向任何兼容 OpenAI API 的服务, 只需修改 BaseURL。
	// 也可以通过修改 DefaultHeaders 支持不同认证方式。
	config := DefaultConfig(apiKey)
	// config.BaseURL = "https://api.groq.com/openai/v1" // 示例：切换到 Groq
	// config.BaseURL = "https://open.bigmodel.cn/api/paas/v4" // 示例：切换到智谱AI
	config.BaseURL = "https://api.ppinfra.com/v3/openai"
	//config.DefaultHeaders["Authorization"] = "Bearer " + os.Getenv("ZHIPU_API_KEY")

	config.MaxHistoryTokens = 2000 // 设置较小的历史记录，方便演示截断

	client := NewClient(config)

	// --- 示例1：同步调用，并使用自定义参数 ---
	fmt.Println("--- 1. 同步调用 (Sync Call) ---")
	syncRequest := ChatRequest{
		Model: "deepseek/deepseek-v3.2-exp",
		Messages: []ChatMessage{
			{Role: "user", Content: "你好，请介绍一下自己。"},
		},
		Temperature: 0.7,
		// 示例：为智谱AI添加自定义参数
		// CustomParams: map[string]any{
		// 	"request_id": fmt.Sprintf("my-app-%d", time.Now().Unix()),
		// },
	}

	resp, err := client.CreateChatCompletion(context.Background(), syncRequest)
	if err != nil {
		log.Fatalf("同步调用失败: %v", err)
	}
	fmt.Printf("同步回复: %s\n\n", resp.Choices[0].Message.Content)

	// --- 示例2：第二次同步调用，测试历史上下文 ---
	fmt.Println("--- 2. 第二次同步调用 (Testing History) ---")
	secondSyncRequest := ChatRequest{
		Model: "deepseek/deepseek-v3.2-exp",
		Messages: []ChatMessage{
			{Role: "user", Content: "我刚才问了你什么问题？"},
		},
	}
	resp, err = client.CreateChatCompletion(context.Background(), secondSyncRequest)
	if err != nil {
		log.Fatalf("第二次同步调用失败: %v", err)
	}
	fmt.Printf("带有历史上下文的回复: %s\n\n", resp.Choices[0].Message.Content)

	fmt.Println("当前历史记录：")
	for _, msg := range client.GetHistory() {
		fmt.Printf("  - %s: %s\n", msg.Role, msg.Content)
	}
	fmt.Println()

	client.ClearHistory() // 清理历史，准备流式示例

	// --- 示例3：流式调用 ---
	fmt.Println("--- 3. SSE 流式调用 (Stream Call) ---")
	streamRequest := ChatRequest{
		Model: "deepseek/deepseek-v3.2-exp",
		Messages: []ChatMessage{
			{Role: "user", Content: "用Go语言写一个经典的Hello World程序，并用Markdown代码块包裹起来。"},
		},
	}

	streamChan, err := client.CreateChatCompletionSSEStream(context.Background(), streamRequest)
	if err != nil {
		log.Fatalf("流式调用失败: %v", err)
	}

	fmt.Print("流式回复: ")
	for event := range streamChan {
		if event.Error != nil {
			log.Printf("流式处理中发生错误: %v", event.Error)
			break
		}
		if len(event.Data.Choices) > 0 {
			content := event.Data.Choices[0].Delta.Content
			fmt.Print(content)
		}
	}
	fmt.Println("\n\n流式调用结束。")

	fmt.Println("当前历史记录：")
	for _, msg := range client.GetHistory() {
		content := msg.Content
		if len(content) > 80 {
			content = content[:80] + "..."
		}
		fmt.Printf("  - %s: %s\n", msg.Role, strings.ReplaceAll(content, "\n", " "))
	}

	// --- 示例4：WebSocket 流式调用 ---
	// 注意：模型名称需要换成目标平台支持的，例如 "glm-4"
	fmt.Println("--- 4. WebSocket 流式调用 (WebSocket Stream Call) ---")
	wsRequest := ChatRequest{
		Model: "deepseek/deepseek-v3.2-exp",
		Messages: []ChatMessage{
			{Role: "user", Content: "请用 Python 写一个简单的 web 服务器，并用 Markdown 代码块包裹。"},
		},
	}

	// 调用新增的 WebSocket 方法
	wsStreamChan, err := client.CreateChatCompletionWebSocketStream(context.Background(), wsRequest)
	if err != nil {
		log.Fatalf("WebSocket 流式调用失败: %v", err)
	}

	fmt.Print("WebSocket 流式回复: ")
	for event := range wsStreamChan {
		if event.Error != nil {
			log.Printf("\nWebSocket 流式处理中发生错误: %v", event.Error)
			break
		}
		if len(event.Data.Choices) > 0 {
			content := event.Data.Choices[0].Delta.Content
			fmt.Print(content)
		}
	}
	fmt.Println("\n\nWebSocket 流式调用结束。")

	fmt.Println("当前历史记录：")
	for _, msg := range client.GetHistory() {
		content := msg.Content
		if len(content) > 80 {
			content = content[:80] + "..."
		}
		fmt.Printf("  - %s: %s\n", msg.Role, strings.ReplaceAll(content, "\n", " "))
	}
}
