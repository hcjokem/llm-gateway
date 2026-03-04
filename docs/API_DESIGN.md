# API 接口设计

## 认证说明

- 管理端 API：使用管理员 Token（Bearer Token）
- 代理端 API：使用虚拟 Key（API Key）

## 响应格式

### 成功响应
```json
{
  "success": true,
  "data": {},
  "message": "操作成功"
}
```

### 错误响应
```json
{
  "success": false,
  "error": "错误信息",
  "code": "ERROR_CODE"
}
```

---

## 管理端 API

### 1. 模型管理

#### 1.1 获取模型列表
```
GET /api/v1/admin/models?page=1&limit=20&provider=openai
```

响应：
```json
{
  "success": true,
  "data": {
    "total": 10,
    "items": [
      {
        "id": "1",
        "name": "gpt-4",
        "display_name": "GPT-4",
        "provider": "openai",
        "type": "chat",
        "context_length": 8192,
        "enabled": true,
        "pricing": {
          "input_price": 0.03,
          "output_price": 0.06,
          "currency": "USD"
        },
        "created_at": "2026-03-05T00:00:00Z"
      }
    ]
  }
}
```

#### 1.2 创建模型
```
POST /api/v1/admin/models
Content-Type: application/json

{
  "name": "gpt-4",
  "display_name": "GPT-4",
  "provider": "openai",
  "type": "chat",
  "context_length": 8192,
  "api_key": "sk-xxx",
  "api_base": "https://api.openai.com/v1",
  "enabled": true,
  "pricing": {
    "input_price": 0.03,
    "output_price": 0.06,
    "currency": "USD"
  }
}
```

#### 1.3 更新模型
```
PUT /api/v1/admin/models/:id
Content-Type: application/json

{
  "enabled": false,
  "pricing": {
    "input_price": 0.02,
    "output_price": 0.05
  }
}
```

#### 1.4 删除模型
```
DELETE /api/v1/admin/models/:id
```

---

### 2. 虚拟 Key 管理

#### 2.1 获取 Key 列表
```
GET /api/v1/admin/keys?page=1&limit=20&status=active
```

响应：
```json
{
  "success": true,
  "data": {
    "total": 5,
    "items": [
      {
        "id": "1",
        "key": "sk-gateway-xxx",
        "name": "项目 A Key",
        "models": ["gpt-4", "gpt-3.5-turbo"],
        "quota": 1000000,
        "used": 50000,
        "remaining": 950000,
        "ip_whitelist": ["1.1.1.1"],
        "expires_at": null,
        "status": "active",
        "created_at": "2026-03-05T00:00:00Z"
      }
    ]
  }
}
```

#### 2.2 创建 Key
```
POST /api/v1/admin/keys
Content-Type: application/json

{
  "name": "项目 A Key",
  "models": ["gpt-4", "gpt-3.5-turbo"],
  "quota": 1000000,
  "ip_whitelist": ["1.1.1.1"],
  "expires_at": null
}
```

#### 2.3 更新 Key
```
PUT /api/v1/admin/keys/:id
Content-Type: application/json

{
  "quota": 2000000,
  "status": "disabled"
}
```

#### 2.4 删除 Key
```
DELETE /api/v1/admin/keys/:id
```

---

### 3. 用量统计

#### 3.1 获取总用量统计
```
GET /api/v1/admin/usage?from=2026-03-01&to=2026-03-05&group_by=model
```

响应：
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_tokens": 1000000,
      "total_requests": 500,
      "total_cost": 30.50
    },
    "by_model": [
      {
        "model": "gpt-4",
        "tokens": 600000,
        "requests": 300,
        "cost": 20.00
      },
      {
        "model": "gpt-3.5-turbo",
        "tokens": 400000,
        "requests": 200,
        "cost": 10.50
      }
    ],
    "timeline": [
      {
        "date": "2026-03-01",
        "tokens": 200000,
        "cost": 5.00
      }
    ]
  }
}
```

#### 3.2 获取指定 Key 的用量
```
GET /api/v1/admin/usage/keys/:key?from=2026-03-01&to=2026-03-05
```

#### 3.3 获取实时用量
```
GET /api/v1/admin/usage/realtime
```

---

### 4. 计费配置

#### 4.1 获取计费规则
```
GET /api/v1/admin/pricing
```

响应：
```json
{
  "success": true,
  "data": {
    "billing_mode": "pay_as_you_go", // pay_as_you_go | package
    "currency": "USD",
    "tax_rate": 0.0,
    "models": {
      "gpt-4": {
        "input_price": 0.03,
        "output_price": 0.06,
        "unit": "token"
      }
    },
    "packages": [
      {
        "id": "1",
        "name": "标准套餐",
        "price": 29.99,
        "currency": "USD",
        "duration": 30,
        "quota": 5000000
      }
    ]
  }
}
```

#### 4.2 更新计费规则
```
PUT /api/v1/admin/pricing
Content-Type: application/json

{
  "billing_mode": "package",
  "currency": "CNY",
  "tax_rate": 0.06
}
```

#### 4.3 创建套餐
```
POST /api/v1/admin/packages
Content-Type: application/json

{
  "name": "标准套餐",
  "price": 199.99,
  "currency": "CNY",
  "duration": 30,
  "quota": 10000000,
  "models": ["gpt-4", "gpt-3.5-turbo"]
}
```

---

### 5. 告警配置

#### 5.1 获取告警规则
```
GET /api/v1/admin/alerts
```

响应：
```json
{
  "success": true,
  "data": {
    "rules": [
      {
        "id": "1",
        "type": "quota_warning", // quota_warning | cost_warning | key_expiry
        "threshold": 80,
        "enabled": true,
        "notification": {
          "email": ["admin@example.com"],
          "webhook": "https://example.com/webhook"
        }
      }
    ]
  }
}
```

#### 5.2 更新告警规则
```
PUT /api/v1/admin/alerts/:id
Content-Type: application/json

{
  "threshold": 90,
  "enabled": true,
  "notification": {
    "email": ["admin@example.com"],
    "webhook": "https://example.com/webhook"
  }
}
```

---

## 代理端 API（兼容 OpenAI 格式）

### 1. 聊天补全
```
POST /v1/chat/completions
Authorization: Bearer <virtual-key>
Content-Type: application/json

{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "stream": true
}
```

响应：
```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1699999999,
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello!"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 20,
    "total_tokens": 30
  }
}
```

### 2. 嵌入
```
POST /v1/embeddings
Authorization: Bearer <virtual-key>
Content-Type: application/json

{
  "model": "text-embedding-ada-002",
  "input": "Hello world"
}
```

---

## 状态码

- `200` - 成功
- `201` - 创建成功
- `400` - 请求参数错误
- `401` - 未授权
- `403` - 权限不足
- `404` - 资源不存在
- `429` - 超出额度
- `500` - 服务器错误
