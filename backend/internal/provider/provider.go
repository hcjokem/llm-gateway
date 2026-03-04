package provider

import "context"

// Pricing 模型定价
type Pricing struct {
	InputPrice  float64 `json:"input_price"`  // 每 1K tokens 价格
	OutputPrice float64 `json:"output_price"` // 每 1K tokens 价格
	Currency    string  `json:"currency"`     // 货币
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Usage 用量
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// EmbeddingRequest 嵌入请求
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input string   `json:"input"`
}

// EmbeddingResponse 嵌入响应
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}

// Embedding 嵌入
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// Provider LLM Provider 接口
type Provider interface {
	// Name 返回 Provider 名称
	Name() string

	// ChatCompletion 聊天补全
	ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// Embedding 嵌入
	Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// GetPricing 获取定价
	GetPricing(model string) (*Pricing, error)
}
