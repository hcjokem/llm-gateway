package model

import (
	"time"
)

// Model 表示 LLM 模型配置
type Model struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	DisplayName   string     `json:"display_name"`
	Provider      string     `json:"provider"`
	Type          string     `json:"type"`
	ContextLength int        `json:"context_length"`
	APIKey        string     `json:"api_key"`
	APIBase       string     `json:"api_base"`
	Enabled       bool       `json:"enabled"`
	Priority      int        `json:"priority"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ModelPricing 表示模型定价
type ModelPricing struct {
	ID            int       `json:"id"`
	ModelID       int       `json:"model_id"`
	InputPrice    float64   `json:"input_price"`
	OutputPrice   float64   `json:"output_price"`
	Currency      string    `json:"currency"`
	EffectiveDate time.Time `json:"effective_date"`
}

// Key 表示虚拟 API Key
type Key struct {
	ID         int       `json:"id"`
	KeyValue   string    `json:"key"`
	Name       string    `json:"name"`
	Models     []string  `json:"models"`
	Quota      int64     `json:"quota"`
	Used       int64     `json:"used"`
	IPWhitelist []string `json:"ip_whitelist"`
	ExpiresAt  *time.Time `json:"expires_at"`
	Status     string    `json:"status"`
	PackageID  *int      `json:"package_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Package 表示套餐
type Package struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Currency    string    `json:"currency"`
	Duration    int       `json:"duration"`
	Quota       int64     `json:"quota"`
	Models      []string  `json:"models"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UsageRecord 表示用量记录
type UsageRecord struct {
	ID              int        `json:"id"`
	KeyID           int        `json:"key_id"`
	ModelID         int        `json:"model_id"`
	RequestID       string     `json:"request_id"`
	RequestType     string     `json:"request_type"`
	PromptTokens    int        `json:"prompt_tokens"`
	CompletionTokens int        `json:"completion_tokens"`
	TotalTokens     int        `json:"total_tokens"`
	Cost            float64    `json:"cost"`
	Status          string     `json:"status"`
	ErrorMessage    string     `json:"error_message"`
	IPAddress       string     `json:"ip_address"`
	UserAgent       string     `json:"user_agent"`
	RequestTime     time.Time  `json:"request_time"`
	ResponseTime    int        `json:"response_time"`
}

// BillingConfig 表示计费配置
type BillingConfig struct {
	ID              int       `json:"id"`
	BillingMode     string    `json:"billing_mode"`
	Currency        string    `json:"currency"`
	TaxRate         float64   `json:"tax_rate"`
	DefaultPackageID *int      `json:"default_package_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AlertRule 表示告警规则
type AlertRule struct {
	ID                   int       `json:"id"`
	Type                 string    `json:"type"`
	Threshold            float64   `json:"threshold"`
	Enabled              bool      `json:"enabled"`
	NotificationEmail    []string  `json:"notification_email"`
	NotificationWebhook  string    `json:"notification_webhook"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// AdminUser 表示管理员用户
type AdminUser struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"password_hash"`
	Email        string     `json:"email"`
	Role         string     `json:"role"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// UsageStats 表示用量统计
type UsageStats struct {
	Summary struct {
		TotalTokens   int64   `json:"total_tokens"`
		TotalRequests int64   `json:"total_requests"`
		TotalCost     float64 `json:"total_cost"`
	} `json:"summary"`
	ByModel []ModelUsage `json:"by_model"`
	Timeline []TimeUsage `json:"timeline"`
}

// ModelUsage 表示模型用量
type ModelUsage struct {
	Model      string  `json:"model"`
	Tokens     int64   `json:"tokens"`
	Requests   int64   `json:"requests"`
	Cost       float64 `json:"cost"`
}

// TimeUsage 表示时间维度用量
type TimeUsage struct {
	Date   string  `json:"date"`
	Tokens int64   `json:"tokens"`
	Cost   float64 `json:"cost"`
}

// Response 统一响应结构
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}, message string) *Response {
	return &Response{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(error, code string) *Response {
	return &Response{
		Success: false,
		Error:   error,
		Code:    code,
	}
}
