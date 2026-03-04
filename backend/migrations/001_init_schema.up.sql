-- LLM Gateway 数据库初始化脚本

-- 1. models 表 - 模型配置表
CREATE TABLE IF NOT EXISTS models (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(200) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'chat',
    context_length INTEGER NOT NULL DEFAULT 4096,
    api_key TEXT NOT NULL,
    api_base VARCHAR(500) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_models_provider ON models(provider);
CREATE INDEX IF NOT EXISTS idx_models_enabled ON models(enabled);

-- 2. model_pricing 表 - 模型定价表
CREATE TABLE IF NOT EXISTS model_pricing (
    id SERIAL PRIMARY KEY,
    model_id INTEGER NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    input_price DECIMAL(10, 6) NOT NULL,
    output_price DECIMAL(10, 6) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    effective_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (model_id, effective_date)
);

CREATE INDEX IF NOT EXISTS idx_model_pricing_model_id ON model_pricing(model_id);

-- 3. keys 表 - 虚拟 Key 表
CREATE TABLE IF NOT EXISTS keys (
    id SERIAL PRIMARY KEY,
    key_value VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    models TEXT[] NOT NULL,
    quota BIGINT NOT NULL DEFAULT 0,
    used BIGINT NOT NULL DEFAULT 0,
    ip_whitelist TEXT[] DEFAULT ARRAY[]::TEXT[],
    expires_at TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    package_id INTEGER REFERENCES packages(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_keys_key_value ON keys(key_value);
CREATE INDEX IF NOT EXISTS idx_keys_status ON keys(status);
CREATE INDEX IF NOT EXISTS idx_keys_package_id ON keys(package_id);

-- 4. packages 表 - 套餐表
CREATE TABLE IF NOT EXISTS packages (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    duration INTEGER NOT NULL,
    quota BIGINT NOT NULL,
    models TEXT[] NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_packages_enabled ON packages(enabled);

-- 5. usage_records 表 - 用量记录表
CREATE TABLE IF NOT EXISTS usage_records (
    id BIGSERIAL PRIMARY KEY,
    key_id INTEGER NOT NULL REFERENCES keys(id) ON DELETE CASCADE,
    model_id INTEGER NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    request_id VARCHAR(100) NOT NULL,
    request_type VARCHAR(20) NOT NULL,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    cost DECIMAL(10, 6) NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    ip_address INET,
    user_agent TEXT,
    request_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    response_time INTEGER
);

CREATE INDEX IF NOT EXISTS idx_usage_records_key_id ON usage_records(key_id);
CREATE INDEX IF NOT EXISTS idx_usage_records_model_id ON usage_records(model_id);
CREATE INDEX IF NOT EXISTS idx_usage_records_request_time ON usage_records(request_time);
CREATE INDEX IF NOT EXISTS idx_usage_records_status ON usage_records(status);
CREATE INDEX IF NOT EXISTS idx_usage_records_key_time ON usage_records(key_id, request_time);
CREATE INDEX IF NOT EXISTS idx_usage_records_model_time ON usage_records(model_id, request_time);

-- 6. billing_config 表 - 计费配置表
CREATE TABLE IF NOT EXISTS billing_config (
    id SERIAL PRIMARY KEY,
    billing_mode VARCHAR(20) NOT NULL DEFAULT 'pay_as_you_go',
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    tax_rate DECIMAL(5, 4) NOT NULL DEFAULT 0.0,
    default_package_id INTEGER REFERENCES packages(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 7. alert_rules 表 - 告警规则表
CREATE TABLE IF NOT EXISTS alert_rules (
    id SERIAL PRIMARY KEY,
    type VARCHAR(30) NOT NULL,
    threshold DECIMAL(5, 2) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    notification_email TEXT[] DEFAULT ARRAY[]::TEXT[],
    notification_webhook VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_rules_type ON alert_rules(type);
CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled);

-- 8. admin_users 表 - 管理员表
CREATE TABLE IF NOT EXISTS admin_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    role VARCHAR(20) NOT NULL DEFAULT 'admin',
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_users_username ON admin_users(username);
CREATE INDEX IF NOT EXISTS idx_admin_users_email ON admin_users(email);

-- 9. api_logs 表 - API 请求日志表
CREATE TABLE IF NOT EXISTS api_logs (
    id BIGSERIAL PRIMARY KEY,
    method VARCHAR(10) NOT NULL,
    path VARCHAR(500) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time INTEGER,
    ip_address INET,
    user_agent TEXT,
    request_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_api_logs_request_time ON api_logs(request_time);
CREATE INDEX IF NOT EXISTS idx_api_logs_status_code ON api_logs(status_code);

-- 10. key_renewals 表 - Key 续费记录表
CREATE TABLE IF NOT EXISTS key_renewals (
    id SERIAL PRIMARY KEY,
    key_id INTEGER NOT NULL REFERENCES keys(id) ON DELETE CASCADE,
    package_id INTEGER NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
    old_quota BIGINT NOT NULL,
    new_quota BIGINT NOT NULL,
    price_paid DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_key_renewals_key_id ON key_renewals(key_id);
CREATE INDEX IF NOT EXISTS idx_key_renewals_created_at ON key_renewals(created_at);

-- 11. 创建视图

-- key_usage_view - Key 用量汇总视图
CREATE OR REPLACE VIEW key_usage_view AS
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

-- model_usage_view - 模型用量汇总视图
CREATE OR REPLACE VIEW model_usage_view AS
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

-- 12. 插入初始数据

-- 默认管理员用户 (密码: admin123)
INSERT INTO admin_users (username, password_hash, email, role)
VALUES ('admin', '240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9', 'admin@llmgateway.com', 'admin')
ON CONFLICT (username) DO NOTHING;

-- 默认计费配置
INSERT INTO billing_config (billing_mode, currency, tax_rate)
VALUES ('pay_as_you_go', 'USD', 0.0)
ON CONFLICT DO NOTHING;

-- 默认告警规则
INSERT INTO alert_rules (type, threshold, enabled)
VALUES
    ('quota_warning', 80.0, true),
    ('cost_warning', 100.0, true),
    ('key_expiry', 7.0, true)
ON CONFLICT DO NOTHING;
