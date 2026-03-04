# 数据库设计

## 表结构

### 1. models - 模型配置表

```sql
CREATE TABLE models (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(200) NOT NULL,
    provider VARCHAR(50) NOT NULL, -- openai, anthropic, zhipu, qwen
    type VARCHAR(20) NOT NULL, -- chat, embedding, image
    context_length INTEGER NOT NULL DEFAULT 4096,
    api_key TEXT NOT NULL, -- 加密存储
    api_base VARCHAR(500) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_provider (provider),
    INDEX idx_enabled (enabled)
);
```

### 2. model_pricing - 模型定价表

```sql
CREATE TABLE model_pricing (
    id SERIAL PRIMARY KEY,
    model_id INTEGER NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    input_price DECIMAL(10, 6) NOT NULL, -- 每 1K tokens 价格
    output_price DECIMAL(10, 6) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    effective_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (model_id, effective_date),
    INDEX idx_model_id (model_id)
);
```

### 3. keys - 虚拟 Key 表

```sql
CREATE TABLE keys (
    id SERIAL PRIMARY KEY,
    key_value VARCHAR(100) NOT NULL UNIQUE, -- sk-gateway-xxx
    name VARCHAR(200) NOT NULL,
    models TEXT[] NOT NULL, -- 允许使用的模型列表
    quota BIGINT NOT NULL DEFAULT 0, -- 总额度（tokens）
    used BIGINT NOT NULL DEFAULT 0, -- 已使用额度
    ip_whitelist TEXT[] DEFAULT ARRAY[]::TEXT[], -- IP 白名单
    expires_at TIMESTAMP, -- 过期时间
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, disabled, expired
    package_id INTEGER REFERENCES packages(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_key_value (key_value),
    INDEX idx_status (status),
    INDEX idx_package_id (package_id)
);
```

### 4. packages - 套餐表

```sql
CREATE TABLE packages (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    duration INTEGER NOT NULL, -- 天数
    quota BIGINT NOT NULL, -- 包含的 token 数量
    models TEXT[] NOT NULL, -- 包含的模型列表
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_enabled (enabled)
);
```

### 5. usage_records - 用量记录表

```sql
CREATE TABLE usage_records (
    id BIGSERIAL PRIMARY KEY,
    key_id INTEGER NOT NULL REFERENCES keys(id) ON DELETE CASCADE,
    model_id INTEGER NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    request_id VARCHAR(100) NOT NULL,
    request_type VARCHAR(20) NOT NULL, -- chat, embedding, image
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    cost DECIMAL(10, 6) NOT NULL,
    status VARCHAR(20) NOT NULL, -- success, failed
    error_message TEXT,
    ip_address INET,
    user_agent TEXT,
    request_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    response_time INTEGER, -- 响应时间（毫秒）
    INDEX idx_key_id (key_id),
    INDEX idx_model_id (model_id),
    INDEX idx_request_time (request_time),
    INDEX idx_status (status)
);
```

### 6. billing_config - 计费配置表

```sql
CREATE TABLE billing_config (
    id SERIAL PRIMARY KEY,
    billing_mode VARCHAR(20) NOT NULL DEFAULT 'pay_as_you_go', -- pay_as_you_go, package
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    tax_rate DECIMAL(5, 4) NOT NULL DEFAULT 0.0,
    default_package_id INTEGER REFERENCES packages(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 7. alert_rules - 告警规则表

```sql
CREATE TABLE alert_rules (
    id SERIAL PRIMARY KEY,
    type VARCHAR(30) NOT NULL, -- quota_warning, cost_warning, key_expiry
    threshold DECIMAL(5, 2) NOT NULL, -- 百分比或金额
    enabled BOOLEAN NOT NULL DEFAULT true,
    notification_email TEXT[] DEFAULT ARRAY[]::TEXT[],
    notification_webhook VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_type (type),
    INDEX idx_enabled (enabled)
);
```

### 8. admin_users - 管理员表

```sql
CREATE TABLE admin_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    role VARCHAR(20) NOT NULL DEFAULT 'admin', -- admin, viewer
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
);
```

### 9. api_logs - API 请求日志表

```sql
CREATE TABLE api_logs (
    id BIGSERIAL PRIMARY KEY,
    method VARCHAR(10) NOT NULL,
    path VARCHAR(500) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time INTEGER, -- 毫秒
    ip_address INET,
    user_agent TEXT,
    request_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_request_time (request_time),
    INDEX idx_status_code (status_code)
);
```

### 10. key_renewals - Key 续费记录表

```sql
CREATE TABLE key_renewals (
    id SERIAL PRIMARY KEY,
    key_id INTEGER NOT NULL REFERENCES keys(id) ON DELETE CASCADE,
    package_id INTEGER NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
    old_quota BIGINT NOT NULL,
    new_quota BIGINT NOT NULL,
    price_paid DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_key_id (key_id),
    INDEX idx_created_at (created_at)
);
```

---

## 视图

### 1. key_usage_view - Key 用量汇总视图

```sql
CREATE VIEW key_usage_view AS
SELECT
    k.id,
    k.key_value,
    k.name,
    k.quota,
    k.used,
    k.quota - k.used AS remaining,
    k.status,
    k.expires_at,
    COUNT(ur.id) AS total_requests,
    SUM(ur.total_tokens) AS total_tokens,
    SUM(ur.cost) AS total_cost,
    MAX(ur.request_time) AS last_used_at
FROM keys k
LEFT JOIN usage_records ur ON k.id = ur.key_id
GROUP BY k.id;
```

### 2. model_usage_view - 模型用量汇总视图

```sql
CREATE VIEW model_usage_view AS
SELECT
    m.id,
    m.name,
    m.display_name,
    m.provider,
    COUNT(ur.id) AS total_requests,
    SUM(ur.total_tokens) AS total_tokens,
    SUM(ur.cost) AS total_cost,
    AVG(ur.response_time) AS avg_response_time
FROM models m
LEFT JOIN usage_records ur ON m.id = ur.model_id
GROUP BY m.id;
```

---

## 索引优化

```sql
-- 复合索引：用于按时间范围查询用量
CREATE INDEX idx_usage_records_key_time ON usage_records(key_id, request_time);
CREATE INDEX idx_usage_records_model_time ON usage_records(model_id, request_time);

-- 复合索引：用于状态查询
CREATE INDEX idx_keys_status_quotas ON keys(status, quota, used);
```

---

## 初始化数据

### 1. 插入默认管理员

```sql
INSERT INTO admin_users (username, password_hash, email, role)
VALUES ('admin', '$2a$10$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx', 'admin@example.com', 'admin');
```

### 2. 插入默认计费配置

```sql
INSERT INTO billing_config (billing_mode, currency, tax_rate)
VALUES ('pay_as_you_go', 'USD', 0.0);
```

### 3. 插入默认告警规则

```sql
INSERT INTO alert_rules (type, threshold, enabled)
VALUES
    ('quota_warning', 80.0, true),
    ('cost_warning', 100.0, true),
    ('key_expiry', 7.0, true);
```

---

## 备份策略

- 每日凌晨全量备份
- 保留最近 7 天的备份
- 每周清理过期备份

```bash
# 备份脚本
pg_dump -h localhost -U postgres -d llm_gateway > backup_$(date +%Y%m%d).sql
```
