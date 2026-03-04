# 系统架构设计

## 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                         用户 / 客户端                         │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ HTTP/HTTPS
                         │
┌────────────────────────▼────────────────────────────────────┐
│                      Nginx (可选)                            │
│                    反向代理 + 负载均衡                        │
└────────────────────────┬────────────────────────────────────┘
                         │
         ┌───────────────┴───────────────┐
         │                               │
┌────────▼─────────┐         ┌──────────▼──────────┐
│   Web UI (React)  │         │   LLM Gateway       │
│   :8080 (前端)    │         │   :3000 (后端)       │
│                   │         │                     │
│  - 模型配置       │         │  ┌─────────────────┐ │
│  - Key 管理       │         │  │  HTTP Handler   │ │
│  - 用量统计       │         │  └─────────────────┘ │
│  - 计费配置       │         │          │            │
│                   │         │          ▼            │
└───────────────────┘         │  ┌─────────────────┐ │
                             │  │  Router         │ │
                             │  └─────────────────┘ │
                             │          │            │
             ┌───────────────┼──────────┴────────────┐ │
             │               │                       │ │
┌────────────▼─────┐ ┌───────▼──────┐   ┌──────────▼────────┐
│   Middleware    │ │    Service   │   │    Proxy         │
├─────────────────┤ ├──────────────┤   ├──────────────────┤
│ - Auth          │ │ - Model      │   │ - OpenAI         │
│ - Rate Limit    │ │ - Key        │   │ - Anthropic      │
│ - Logging       │ │ - Usage      │   │ - 智谱 GLM       │
│ - Recovery      │ │ - Billing    │   │ - 通义千问       │
└─────────────────┘ │ - Alert      │   │ - DeepSeek       │
                    │ - Admin      │   └──────────────────┘
                    └──────────────┘           │
                                ┌──────────────┴──────────────┐
                                │                              │
                        ┌───────▼──────┐            ┌──────────▼──────────┐
                        │ PostgreSQL   │            │   Redis (可选)       │
                        │   :5432      │            │   :6379            │
                        │              │            │  - 缓存             │
                        │ - models     │            │  - 限流             │
                        │ - keys       │            │  - 会话             │
                        │ - usage      │            │                     │
                        │ - billing    │            │                     │
                        │ - alerts     │            │                     │
                        └──────────────┘            └─────────────────────┘
```

---

## 后端架构（Go）

### 目录结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── config/
│   │   ├── config.go           # 配置管理
│   │   └── database.go         # 数据库配置
│   ├── model/
│   │   ├── model.go            # 数据模型定义
│   │   └── dto.go              # 请求/响应 DTO
│   ├── repository/
│   │   ├── model_repository.go
│   │   ├── key_repository.go
│   │   ├── usage_repository.go
│   │   ├── billing_repository.go
│   │   └── alert_repository.go
│   ├── service/
│   │   ├── model_service.go
│   │   ├── key_service.go
│   │   ├── usage_service.go
│   │   ├── billing_service.go
│   │   ├── alert_service.go
│   │   └── proxy_service.go    # API 代理核心
│   ├── handler/
│   │   ├── admin_handler.go    # 管理端 API
│   │   ├── proxy_handler.go    # 代理端 API
│   │   └── middleware.go       # 中间件
│   ├── provider/
│   │   ├── provider.go         # Provider 接口
│   │   ├── openai.go           # OpenAI 实现
│   │   ├── anthropic.go        # Anthropic 实现
│   │   ├── zhipu.go            # 智谱 GLM 实现
│   │   └── qwen.go             # 通义千问实现
│   └── util/
│       ├── jwt.go              # JWT 工具
│       ├── crypto.go           # 加密工具
│       ├── logger.go           # 日志工具
│       └── validator.go        # 验证工具
├── migrations/
│   └── 001_init_schema.up.sql  # 数据库迁移
├── go.mod
└── go.sum
```

---

## 核心模块设计

### 1. Provider 接口（统一 LLM Provider）

```go
type Provider interface {
    // 获取 Provider 名称
    Name() string

    // 聊天补全
    ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

    // 嵌入
    Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

    // 计算 token 数
    CountTokens(text string) int

    // 获取定价
    GetPricing(model string) (*Pricing, error)
}
```

### 2. 代理服务（Proxy Service）

```go
type ProxyService struct {
    providers map[string]Provider
    keyRepo   repository.KeyRepository
    usageRepo repository.UsageRepository
    modelRepo repository.ModelRepository
}

func (s *ProxyService) HandleRequest(
    ctx context.Context,
    key string,
    req *ProxyRequest,
) (*ProxyResponse, error) {
    // 1. 验证 Key
    keyRecord, err := s.keyRepo.GetByKey(key)
    if err != nil {
        return nil, ErrInvalidKey
    }

    // 2. 检查额度
    if keyRecord.Used >= keyRecord.Quota {
        return nil, ErrQuotaExceeded
    }

    // 3. 获取 Provider
    provider, err := s.getProvider(req.Model)
    if err != nil {
        return nil, err
    }

    // 4. 转发请求
    resp, err := provider.ChatCompletion(ctx, req.ChatRequest)
    if err != nil {
        // 记录失败
        s.recordUsage(ctx, keyRecord.ID, req, nil, err)
        return nil, err
    }

    // 5. 记录用量
    s.recordUsage(ctx, keyRecord.ID, req, resp, nil)

    // 6. 更新 Key 使用量
    s.keyRepo.IncrementUsage(keyRecord.ID, resp.Usage.TotalTokens)

    return resp, nil
}
```

### 3. 用量服务（Usage Service）

```go
type UsageService struct {
    usageRepo   repository.UsageRepository
    alertRepo   repository.AlertRepository
}

func (s *UsageService) GetUsageStats(
    ctx context.Context,
    filter *UsageFilter,
) (*UsageStats, error) {
    return s.usageRepo.GetStats(ctx, filter)
}

func (s *UsageService) CheckAlerts(ctx context.Context) error {
    // 1. 获取所有活跃的告警规则
    rules, err := s.alertRepo.GetActiveAlerts(ctx)
    if err != nil {
        return err
    }

    // 2. 检查每个规则
    for _, rule := range rules {
        switch rule.Type {
        case "quota_warning":
            s.checkQuotaWarning(ctx, rule)
        case "cost_warning":
            s.checkCostWarning(ctx, rule)
        case "key_expiry":
            s.checkKeyExpiry(ctx, rule)
        }
    }

    return nil
}
```

---

## 前端架构（React）

### 目录结构

```
frontend/
├── src/
│   ├── api/
│   │   ├── client.ts            # API 客户端
│   │   ├── models.ts            # 模型 API
│   │   ├── keys.ts              # Key API
│   │   ├── usage.ts             # 用量 API
│   │   ├── billing.ts           # 计费 API
│   │   └── alerts.ts            # 告警 API
│   ├── components/
│   │   ├── common/
│   │   │   ├── Layout.tsx       # 布局组件
│   │   │   ├── Navbar.tsx       # 导航栏
│   │   │   └── Sidebar.tsx      # 侧边栏
│   │   ├── models/
│   │   │   ├── ModelList.tsx
│   │   │   ├── ModelForm.tsx
│   │   │   └── ModelCard.tsx
│   │   ├── keys/
│   │   │   ├── KeyList.tsx
│   │   │   ├── KeyForm.tsx
│   │   │   └── KeyCard.tsx
│   │   ├── usage/
│   │   │   ├── UsageChart.tsx
│   │   │   ├── UsageTable.tsx
│   │   │   └── UsageSummary.tsx
│   │   └── alerts/
│   │       ├── AlertList.tsx
│   │       └── AlertForm.tsx
│   ├── pages/
│   │   ├── Dashboard.tsx        # 仪表盘
│   │   ├── Models.tsx           # 模型管理
│   │   ├── Keys.tsx             # Key 管理
│   │   ├── Usage.tsx            # 用量统计
│   │   ├── Billing.tsx          # 计费配置
│   │   └── Alerts.tsx           # 告警配置
│   ├── router/
│   │   └── index.tsx            # 路由配置
│   ├── store/
│   │   ├── index.ts             # Redux store
│   │   └── slices/              # Redux slices
│   ├── styles/
│   │   ├── global.css
│   │   └── variables.css
│   ├── utils/
│   │   ├── format.ts            # 格式化工具
│   │   ├── chart.ts             # 图表工具
│   │   └── constants.ts         # 常量
│   ├── App.tsx
│   ├── main.tsx
│   └── vite-env.d.ts
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

---

## 数据流设计

### 1. API 代理请求流程

```
客户端请求
    │
    ▼
Middleware (Auth, Rate Limit)
    │
    ▼
Proxy Handler
    │
    ├─► 验证 Key
    │   ├─► KeyRepository.GetByKey()
    │   └─► 检查额度、IP、模型权限
    │
    ├─► 获取 Provider
    │   └─► ModelRepository.GetModel()
    │
    ├─► 转发请求
    │   ├─► Provider.ChatCompletion()
    │   └─► 处理 stream
    │
    ├─► 记录用量
    │   └─► UsageRepository.Create()
    │
    └─► 返回响应
```

### 2. 用量统计查询流程

```
前端请求
    │
    ▼
Handler
    │
    ├─► 验证权限
    │
    ├─► 查询数据库
    │   └─► UsageRepository.GetStats()
    │
    ├─► 聚合数据
    │   ├─► 按模型聚合
    │   ├─► 按时间聚合
    │   └─► 计算总计
    │
    └─► 返回数据
        │
        ▼
前端图表展示
```

---

## 部署架构

### Docker Compose 部署

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: llm-gateway-db
    environment:
      POSTGRES_DB: llm_gateway
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    container_name: llm-gateway-redis
    ports:
      - "6379:6379"

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: llm-gateway-backend
    ports:
      - "3000:3000"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=llm_gateway
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_SECRET=your-jwt-secret
    depends_on:
      - postgres
      - redis

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: llm-gateway-frontend
    ports:
      - "8080:80"
    depends_on:
      - backend

volumes:
  postgres_data:
```

### 单文件部署

```bash
# 构建
cd backend && go build -o llm-gateway ./cmd/server

# 运行
./llm-gateway --config config.yaml
```

---

## 性能优化

### 1. 数据库优化
- 合理使用索引
- 使用视图减少查询复杂度
- 定期归档历史数据

### 2. 缓存策略
- Redis 缓存热点数据（模型列表、Key 信息）
- 本地缓存 Provider 实例

### 3. 并发处理
- 使用 goroutine 处理并发请求
- 连接池复用数据库连接

### 4. 日志优化
- 异步日志写入
- 日志分级管理

---

## 安全设计

### 1. 认证与授权
- JWT Token 认证
- 角色权限控制（admin/viewer）

### 2. 数据加密
- API Key 加密存储（AES-256）
- 敏感配置加密

### 3. 限流与防护
- IP 白名单
- 请求频率限制
- 额度限制

### 4. 审计日志
- 记录所有管理操作
- 记录所有 API 调用
