# LLM Gateway - API 代理与用量监控系统

## 项目概述

一个实用的 LLM API 代理和用量监控工具，支持统一代理多个 LLM、用量统计、计费配置、虚拟 Key 分配等功能。

## 核心功能

### 1. API 代理
- 统一转发多个 LLM 请求（OpenAI、Anthropic、智谱、通义千问等）
- 支持 stream 模式
- 自动负载均衡
- 失败重试机制

### 2. 用量监控
- 实时统计 token 数量
- 调用次数统计
- 成本计算
- 按时间维度聚合（小时/天/周/月）

### 3. 计费配置
- 按量计费（按 token 数计费）
- 套餐计费（月费+免费额度）
- 灵活切换计费模式
- 多种模型定价配置

### 4. 虚拟 Key 管理
- 生成虚拟 API Key
- 配置额度、过期时间、IP 白名单
- 独立的 base URL
- 用量限制

### 5. Web UI
- 模型配置管理
- 用量统计图表
- 告警配置
- 计费规则配置

## 技术栈

- **后端：** Go 1.22+
- **前端：** React 18 + Ant Design 5
- **数据库：** PostgreSQL 15+
- **部署：** Docker / 单文件部署

## 快速开始

```bash
# 克隆仓库
git clone https://github.com/hcjokem/llm-gateway.git
cd llm-gateway

# 启动服务
docker-compose up -d

# 访问 Web UI
open http://localhost:8080
```

## 项目结构

```
llm-gateway/
├── backend/           # Go 后端
│   ├── cmd/          # 主程序入口
│   ├── internal/     # 内部实现
│   ├── api/          # API 处理器
│   ├── service/      # 业务逻辑
│   ├── model/        # 数据模型
│   ├── middleware/   # 中间件
│   └── config/       # 配置
├── frontend/         # React 前端
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── api/
│   │   └── utils/
├── scripts/          # 脚本
└── docs/            # 文档
```

## API 文档

### 认证

所有 API 请求需要携带 Bearer Token：
```
Authorization: Bearer <your-admin-token>
```

### 核心 API

#### 模型管理
- `GET /api/v1/models` - 获取模型列表
- `POST /api/v1/models` - 创建模型
- `PUT /api/v1/models/:id` - 更新模型
- `DELETE /api/v1/models/:id` - 删除模型

#### 虚拟 Key 管理
- `GET /api/v1/keys` - 获取 Key 列表
- `POST /api/v1/keys` - 创建 Key
- `PUT /api/v1/keys/:id` - 更新 Key
- `DELETE /api/v1/keys/:id` - 删除 Key

#### 用量统计
- `GET /api/v1/usage` - 获取用量统计
- `GET /api/v1/usage/:key` - 获取指定 Key 的用量

#### 计费配置
- `GET /api/v1/pricing` - 获取计费规则
- `PUT /api/v1/pricing` - 更新计费规则
- `POST /api/v1/packages` - 创建套餐
- `GET /api/v1/packages` - 获取套餐列表

#### 代理接口
- `POST /v1/chat/completions` - 聊天补全（兼容 OpenAI 格式）
- `POST /v1/embeddings` - 嵌入（兼容 OpenAI 格式）

## 参考项目

- [OneAPI](https://github.com/songquanpeng/one-api) - LLM API 管理与分发系统
- [LiteLLM](https://github.com/BerriAI/litellm) - Python SDK + AI Gateway

## License

MIT License
