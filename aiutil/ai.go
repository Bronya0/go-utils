package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime/debug"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

// =================================================================================
// 1. 通用数据结构 (以 OpenAI Chat API 为例)
//    用户可以根据不同的 AI 提供商，定义自己的请求和响应结构体。
// =================================================================================

// ChatMessage 代表一次对话中的单条消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	// 可根据需要添加其他字段, 如 Name, ToolCalls 等
}

// ChatRequest 是我们封装的、通用的对话请求结构
type ChatRequest struct {
	Model            string        `json:"model"`
	Messages         []ChatMessage `json:"messages"`
	Stream           bool          `json:"stream,omitempty"`
	Temperature      float32       `json:"temperature,omitempty"`
	TopP             float32       `json:"top_p,omitempty"`
	MaxTokens        int           `json:"max_tokens,omitempty"`
	N                int           `json:"n,omitempty"`
	Stop             []string      `json:"stop,omitempty"`
	PresencePenalty  float32       `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32       `json:"frequency_penalty,omitempty"`
	// ... 其他官方支持的参数 ...

	// CustomParams 用于存放任何非官方、模型特定的参数。
	// 例如智谱AI的 'request_id' 或 'meta' 字段。
	CustomParams map[string]any `json:"-"` // 这个字段不直接参与序列化

	// RequestEndpoint 允许覆盖客户端配置中的默认端点。
	// 这使得同一个客户端可以调用不同的API，如 /chat/completions, /embeddings 等。
	RequestEndpoint string `json:"-"`
}

// ChatResponse 是同步模式的响应结构
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ChatStreamResponse 是流式模式下，每个数据块(chunk)的响应结构
type ChatStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// StreamEvent 封装了流式响应的数据或可能发生的错误
type StreamEvent struct {
	Data  ChatStreamResponse
	Error error
}

// =================================================================================
// 2. 客户端配置 (通用设计)
// =================================================================================

// Config 定义了 AI 客户端的配置
type Config struct {
	BaseURL          string
	DefaultEndpoint  string            // 默认 API 端点, e.g., "/chat/completions"
	DefaultHeaders   map[string]string // 用于设置认证、组织ID等
	HTTPClient       *http.Client
	Timeout          time.Duration
	MaxHistoryTokens int // 用于自动历史截断的最大Token数
}

// DefaultConfig 创建一个默认配置
// authToken 用于设置 OpenAI 兼容的 Bearer Token
func DefaultConfig(authToken string) Config {
	return Config{
		BaseURL:         "https://api.openai.com/v1",
		DefaultEndpoint: "/chat/completions",
		DefaultHeaders: map[string]string{
			"Authorization": "Bearer " + authToken,
			"Content-Type":  "application/json",
		},
		Timeout:          120 * time.Second,
		MaxHistoryTokens: 4096, // 默认保留4k token的历史上下文
	}
}

// Client 核心客户端 Client (仅依赖标准库)
// 使用示例
//
//	apiKey := os.Getenv("OPENAI_API_KEY")
//	if apiKey == "" {
//		log.Fatal("请设置环境变量 OPENAI_API_KEY")
//	}
//
//	// --- 客户端配置 ---
//	// 可以指向任何兼容 OpenAI API 的服务, 只需修改 BaseURL。
//	// 也可以通过修改 DefaultHeaders 支持不同认证方式。
//	config := DefaultConfig(apiKey)
//	// config.BaseURL = "https://api.groq.com/openai/v1" // 示例：切换到 Groq
//	// config.BaseURL = "https://open.bigmodel.cn/api/paas/v4" // 示例：切换到智谱AI
//	// config.DefaultHeaders["Authorization"] = "Bearer " + os.Getenv("ZHIPU_API_KEY")  // 自定义认证方式
//
//	config.MaxHistoryTokens = 2000 // 设置较小的历史记录，方便演示截断
//
//	client := NewClient(config)
//
//	// --- 示例1：同步调用，并使用自定义参数 ---
//	fmt.Println("--- 1. 同步调用 (Sync Call) ---")
//	syncRequest := ChatRequest{
//		Model: "gpt-4o-mini",
//		Messages: []ChatMessage{
//			{Role: "user", Content: "你好，请介绍一下自己。"},
//		},
//		Temperature: 0.7,
//		// 示例：为智谱AI添加自定义参数
//		// CustomParams: map[string]any{
//		// 	"request_id": fmt.Sprintf("my-app-%d", time.Now().Unix()),
//		// },
//	}
//
//	resp, err := client.CreateChatCompletion(context.Background(), syncRequest)
//	if err != nil {
//		log.Fatalf("同步调用失败: %v", err)
//	}
//	fmt.Printf("同步回复: %s\n\n", resp.Choices[0].Message.Content)
//
//	// --- 示例2：第二次同步调用，测试历史上下文 ---
//	fmt.Println("--- 2. 第二次同步调用 (Testing History) ---")
//	secondSyncRequest := ChatRequest{
//		Model: "gpt-4o-mini",
//		Messages: []ChatMessage{
//			{Role: "user", Content: "我刚才问了你什么问题？"},
//		},
//	}
//	resp, err = client.CreateChatCompletion(context.Background(), secondSyncRequest)
//	if err != nil {
//		log.Fatalf("第二次同步调用失败: %v", err)
//	}
//	fmt.Printf("带有历史上下文的回复: %s\n\n", resp.Choices[0].Message.Content)
//
//	fmt.Println("当前历史记录：")
//	for _, msg := range client.GetHistory() {
//		fmt.Printf("  - %s: %s\n", msg.Role, msg.Content)
//	}
//	fmt.Println()
//
//	client.ClearHistory() // 清理历史，准备流式示例
//
//	// --- 示例3：流式调用 ---
//	fmt.Println("--- 3. SSE 流式调用 (Stream Call) ---")
//	streamRequest := ChatRequest{
//		Model: "gpt-4o-mini",
//		Messages: []ChatMessage{
//			{Role: "user", Content: "用Go语言写一个经典的Hello World程序，并用Markdown代码块包裹起来。"},
//		},
//	}
//
//	streamChan, err := client.CreateChatCompletionSSEStream(context.Background(), streamRequest)
//	if err != nil {
//		log.Fatalf("流式调用失败: %v", err)
//	}
//
//	fmt.Print("流式回复: ")
//	for event := range streamChan {
//		if event.Error != nil {
//			log.Printf("流式处理中发生错误: %v", event.Error)
//			break
//		}
//		if len(event.Data.Choices) > 0 {
//			content := event.Data.Choices[0].Delta.Content
//			fmt.Print(content)
//		}
//	}
//	fmt.Println("\n\n流式调用结束。")
//
//	fmt.Println("当前历史记录：")
//	for _, msg := range client.GetHistory() {
//		content := msg.Content
//		if len(content) > 80 {
//			content = content[:80] + "..."
//		}
//		fmt.Printf("  - %s: %s\n", msg.Role, strings.ReplaceAll(content, "\n", " "))
//	}
//
//
//	// --- 示例4：WebSocket 流式调用 ---
//	// 注意：模型名称需要换成目标平台支持的，例如 "glm-4"
//	fmt.Println("--- 4. WebSocket 流式调用 (WebSocket Stream Call) ---")
//	wsRequest := ChatRequest{
//		Model: "glm-4",
//		Messages: []ChatMessage{
//			{Role: "user", Content: "请用 Python 写一个简单的 web 服务器，并用 Markdown 代码块包裹。"},
//		},
//	}
//
//	// 调用新增的 WebSocket 方法
//	wsStreamChan, err := client.CreateChatCompletionWebSocketStream(context.Background(), wsRequest)
//	if err != nil {
//		log.Fatalf("WebSocket 流式调用失败: %v", err)
//	}
//
//	fmt.Print("WebSocket 流式回复: ")
//	for event := range wsStreamChan {
//		if event.Error != nil {
//			log.Printf("\nWebSocket 流式处理中发生错误: %v", event.Error)
//			break
//		}
//		if len(event.Data.Choices) > 0 {
//			content := event.Data.Choices[0].Delta.Content
//			fmt.Print(content)
//		}
//	}
//	fmt.Println("\n\nWebSocket 流式调用结束。")
//
//	fmt.Println("当前历史记录：")
//	for _, msg := range client.GetHistory() {
//		content := msg.Content
//		if len(content) > 80 {
//			content = content[:80] + "..."
//		}
//		fmt.Printf("  - %s: %s\n", msg.Role, strings.ReplaceAll(content, "\n", " "))
//	}
type Client struct {
	config     Config
	httpClient *http.Client
	history    []ChatMessage
}

// NewClient 使用给定配置创建一个新的客户端
func NewClient(config Config) *Client {
	// 如果用户没有提供自定义的 http.Client，我们就创建一个
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	if config.Timeout > 0 {
		httpClient.Timeout = config.Timeout
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
		history:    make([]ChatMessage, 0),
	}
}

// =================================================================================
// 4. 核心 API 方法 (同步与流式)
// =================================================================================

// CreateChatCompletion 发起一个同步的对话请求
func (c *Client) CreateChatCompletion(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	request.Stream = false // 确保不是流式请求

	// 1. 准备消息（合并历史记录）
	finalMessages := c.pruneHistory(request.Messages)
	request.Messages = finalMessages

	// 2. 构建请求体和 HTTP 请求
	httpReq, err := c.buildHTTPRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 3. 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// 4. 检查响应状态码
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error: status=%s, body=%s", resp.Status, string(bodyBytes))
	}

	// 5. 解析响应
	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	// 6. 成功后，更新历史记录
	c.history = append(c.history, request.Messages[len(request.Messages)-1])
	if len(result.Choices) > 0 {
		c.history = append(c.history, result.Choices[0].Message)
	}

	return &result, nil
}

// CreateChatCompletionSSEStream 发起一个流式的对话请求
// 返回一个只读的 channel，用于接收流式事件 (数据或错误)
// 使用 SSE 协议
// SSE协议规定服务器发送的数据以 "data: " 前缀开始
// 数据结束标记使用 "[DONE]"
func (c *Client) CreateChatCompletionSSEStream(ctx context.Context, request ChatRequest) (<-chan StreamEvent, error) {
	request.Stream = true // 确保是流式请求

	// 1. 准备消息（合并历史记录）
	finalMessages := c.pruneHistory(request.Messages)
	request.Messages = finalMessages

	// 2. 构建请求体和 HTTP 请求
	httpReq, err := c.buildHTTPRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 3. 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	// 4. 检查初始响应状态码
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		bodyBytes, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close() // 别忘了关闭
		return nil, fmt.Errorf("api error on stream init: status=%s, body=%s", resp.Status, string(bodyBytes))
	}

	// 5. 创建 channel 并启动 goroutine 处理流
	streamChan := make(chan StreamEvent)
	go c.processStream(resp, streamChan, request.Messages)

	return streamChan, nil
}

// =================================================================================
// WebSocket 核心 API 方法
// =================================================================================

// CreateChatCompletionWebSocketStream 通过 WebSocket 发起一个流式的对话请求
func (c *Client) CreateChatCompletionWebSocketStream(ctx context.Context, request ChatRequest) (<-chan StreamEvent, error) {
	request.Stream = true
	finalMessages := c.pruneHistory(request.Messages)
	request.Messages = finalMessages

	// 1. 构建 WebSocket URL
	parsedURL, err := url.Parse(c.config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid BaseURL in config: %w", err)
	}

	switch parsedURL.Scheme {
	case "https":
		parsedURL.Scheme = "wss"
	case "http":
		parsedURL.Scheme = "ws"
	case "ws", "wss":
		// 已经是正确的 WebSocket 协议，无需改动
	default:
		return nil, fmt.Errorf("unsupported URL scheme for websocket: %q", parsedURL.Scheme)
	}

	endpoint := c.config.DefaultEndpoint
	if request.RequestEndpoint != "" {
		endpoint = request.RequestEndpoint
	}

	parsedURL.Path = path.Join(parsedURL.Path, endpoint)
	wsURL := parsedURL.String()

	// 2. 创建 WebSocket 配置，并附带请求头
	originURL := c.config.BaseURL
	wsConfig, err := websocket.NewConfig(wsURL, originURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create websocket config: %w", err)
	}
	for k, v := range c.config.DefaultHeaders {
		wsConfig.Header.Set(k, v)
	}

	// 3. 建立 WebSocket 连接
	conn, err := websocket.DialConfig(wsConfig)
	if err != nil {
		return nil, fmt.Errorf("websocket dial failed to %s: %w", wsURL, err)
	}

	// 4. 创建 channel 并启动 goroutine 处理 WebSocket 通信
	streamChan := make(chan StreamEvent)
	go c.processWebSocketStream(ctx, conn, request, streamChan)

	return streamChan, nil
}

// =================================================================================
// 5. 内部辅助方法
// =================================================================================

// buildPayload 将标准参数和自定义参数合并成最终的请求体
func (c *Client) buildPayload(request ChatRequest) ([]byte, error) {
	// 1. 将标准结构体转为 map
	var payload map[string]any
	b, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal base request: %w", err)
	}
	if err := json.Unmarshal(b, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal base request to map: %w", err)
	}

	// 2. 如果存在自定义参数，则合并
	if request.CustomParams != nil {
		for k, v := range request.CustomParams {
			payload[k] = v
		}
	}

	// 3. 重新序列化为最终的 JSON
	return json.Marshal(payload)
}

// buildHTTPRequest 构建一个标准的 http.Request
func (c *Client) buildHTTPRequest(ctx context.Context, request ChatRequest) (*http.Request, error) {
	// 1. 构建请求体
	payloadBytes, err := c.buildPayload(request)
	if err != nil {
		return nil, fmt.Errorf("failed to build payload: %w", err)
	}

	// 2. 确定 API 端点
	endpoint := c.config.DefaultEndpoint
	if request.RequestEndpoint != "" {
		endpoint = request.RequestEndpoint
	}
	_url := c.config.BaseURL + endpoint

	// 3. 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, _url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	// 4. 设置请求头
	for k, v := range c.config.DefaultHeaders {
		req.Header.Set(k, v)
	}

	return req, nil
}

// processStream 在一个单独的 goroutine 中处理流式响应
func (c *Client) processStream(resp *http.Response, streamChan chan<- StreamEvent, userMessages []ChatMessage) {
	// 确保无论如何都能关闭资源和 channel
	defer close(streamChan)
	defer resp.Body.Close()

	var fullResponseContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)

	// SSE协议规定服务器发送的数据以 "data: " 前缀开始
	// 数据结束标记使用 "[DONE]"，这也是SSE流结束的常见标识
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk ChatStreamResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			streamChan <- StreamEvent{Error: fmt.Errorf("error unmarshalling stream chunk: %w", err)}
			continue // 继续尝试处理下一行
		}

		if len(chunk.Choices) > 0 {
			fullResponseContent.WriteString(chunk.Choices[0].Delta.Content)
		}
		streamChan <- StreamEvent{Data: chunk}
	}

	if err := scanner.Err(); err != nil {
		streamChan <- StreamEvent{Error: fmt.Errorf("error reading stream: %w", err)}
	}

	// 流结束后，将用户消息和完整的AI回复添加到历史记录
	c.history = append(c.history, userMessages...)
	c.history = append(c.history, ChatMessage{
		Role:    "assistant",
		Content: fullResponseContent.String(),
	})
}

// processWebSocketStream 在一个 goroutine 中处理 WebSocket 通信
func (c *Client) processWebSocketStream(ctx context.Context, conn *websocket.Conn, request ChatRequest, streamChan chan<- StreamEvent) {
	// 1. 确保资源最终被清理
	defer func() {
		if r := recover(); r != nil {
			log.Printf("严重错误: processWebSocketStream 发生 panic: %v\n%s", r, debug.Stack())
			streamChan <- StreamEvent{Error: fmt.Errorf("panic recovered in websocket processing: %v", r)}
		}
		conn.Close()
		close(streamChan)
	}()

	// 2. 启动一个 goroutine 来监听 context 的取消信号
	// 当 context 被取消时，它会关闭连接，从而使下方的 Receive 调用立即返回错误
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	// 3. 发送初始请求数据
	payloadBytes, err := c.buildPayload(request)
	if err != nil {
		streamChan <- StreamEvent{Error: fmt.Errorf("failed to build websocket payload: %w", err)}
		return
	}
	if err := websocket.Message.Send(conn, payloadBytes); err != nil {
		streamChan <- StreamEvent{Error: fmt.Errorf("failed to send initial websocket message: %w", err)}
		return
	}

	// 4. 循环接收服务器的响应
	var fullResponseContent strings.Builder
	for {
		var chunk ChatStreamResponse
		// Receive 会阻塞，直到收到消息、连接关闭或发生错误
		if err := websocket.JSON.Receive(conn, &chunk); err != nil {
			// 如果错误是 io.EOF 或者与连接关闭相关，说明流正常结束
			if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
				break // 正常退出循环
			}
			// 其他错误则报告给调用者
			streamChan <- StreamEvent{Error: fmt.Errorf("error receiving websocket message: %w", err)}
			return
		}

		if len(chunk.Choices) > 0 {
			fullResponseContent.WriteString(chunk.Choices[0].Delta.Content)
		}
		streamChan <- StreamEvent{Data: chunk}
	}

	// 5. 流结束后，更新历史记录
	c.history = append(c.history, request.Messages...)
	c.history = append(c.history, ChatMessage{
		Role:    "assistant",
		Content: fullResponseContent.String(),
	})
}

// estimateTokens 是一个简单的 Token 估算函数。
// 注意：这只是一个粗略的估算，对于精确控制，建议使用 tiktoken 等官方库。
func estimateTokens(msg ChatMessage) int {
	return len(msg.Content)
}

// pruneHistory 根据 MaxHistoryTokens 截断历史消息
func (c *Client) pruneHistory(newMessages []ChatMessage) []ChatMessage {
	newTokens := 0
	for _, msg := range newMessages {
		newTokens += estimateTokens(msg)
	}

	currentTokenCount := newTokens
	startIndex := len(c.history)
	for i := len(c.history) - 1; i >= 0; i-- {
		msgTokens := estimateTokens(c.history[i])
		if currentTokenCount+msgTokens > c.config.MaxHistoryTokens {
			startIndex = i + 1
			break
		}
		currentTokenCount += msgTokens
		startIndex = i
	}

	finalMessages := make([]ChatMessage, 0)
	if startIndex < len(c.history) {
		finalMessages = append(finalMessages, c.history[startIndex:]...)
	}
	finalMessages = append(finalMessages, newMessages...)

	return finalMessages
}

// GetHistory 返回当前对话历史
func (c *Client) GetHistory() []ChatMessage {
	return c.history
}

// ClearHistory 清空对话历史
func (c *Client) ClearHistory() {
	c.history = make([]ChatMessage, 0)
}

// =================================================================================
// 6. Main 函数 - 使用示例
// =================================================================================

func main() {
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
	// config.DefaultHeaders["Authorization"] = "Bearer " + os.Getenv("ZHIPU_API_KEY")  // 自定义认证方式

	config.MaxHistoryTokens = 2000 // 设置较小的历史记录，方便演示截断

	client := NewClient(config)

	// --- 示例1：同步调用，并使用自定义参数 ---
	fmt.Println("--- 1. 同步调用 (Sync Call) ---")
	syncRequest := ChatRequest{
		Model: "gpt-4o-mini",
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
		Model: "gpt-4o-mini",
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
		Model: "gpt-4o-mini",
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
		Model: "glm-4",
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
