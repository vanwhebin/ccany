# 🌟 ccany 多 API 转换系统使用指南

## 📋 目录
1. [系统启动](#系统启动)
2. [配置渠道](#配置渠道)
3. [使用方式](#使用方式)
4. [API 端点](#api-端点)
5. [实际使用示例](#实际使用示例)
6. [高级功能](#高级功能)

## 🚀 系统启动

### 1. 启动服务器
```bash
go run cmd/server/main.go
```

服务器将在 `http://localhost:8082` 启动

### 2. 访问 Web 界面
打开浏览器访问: `http://localhost:8082`
- 配置 API 密钥
- 管理渠道
- 测试 API 转换

## ⚙️ 配置渠道

### 方式一：通过 Web 界面配置
1. 访问 `http://localhost:8082`
2. 点击"渠道管理"
3. 添加新渠道，配置：
   - **渠道名称**: 如 "OpenAI-GPT4"
   - **提供商**: openai/anthropic/gemini
   - **API 密钥**: 您的真实 API 密钥
   - **自定义密钥**: 客户端使用的密钥
   - **模型映射**: 配置模型转换规则

### 方式二：直接数据库配置
```sql
INSERT INTO channels (name, provider, api_key, custom_key, base_url, models_mapping) 
VALUES (
  'Claude-API', 
  'anthropic', 
  'sk-ant-your-real-key', 
  'ccany-claude-123', 
  'https://api.anthropic.com',
  '{"gpt-4": "claude-3-sonnet-20240229", "gpt-3.5-turbo": "claude-3-haiku-20240307"}'
);
```

## 🔗 使用方式

### 核心概念
- **源格式**: 客户端发送的 API 格式
- **目标提供商**: 实际处理请求的 AI 服务商
- **自动转换**: 系统自动处理格式转换

### 统一端点
系统提供三种标准 API 端点，客户端可以使用任意格式：

#### 1. OpenAI 兼容端点
```
POST /v1/chat/completions
Authorization: Bearer ccany-your-custom-key
```

#### 2. Anthropic 兼容端点
```
POST /v1/messages
x-api-key: ccany-your-custom-key
```

#### 3. Gemini 兼容端点
```
POST /v1beta/models/{model}:generateContent?key=ccany-your-custom-key
POST /v1beta/models/{model}:streamGenerateContent?alt=sse&key=ccany-your-custom-key
```

## 💻 实际使用示例

### 示例 1: OpenAI 客户端调用 Claude API

**配置渠道:**
```json
{
  "name": "Claude-Backend",
  "provider": "anthropic",
  "api_key": "sk-ant-your-real-claude-key",
  "custom_key": "ccany-claude-proxy",
  "models_mapping": {
    "gpt-4": "claude-3-sonnet-20240229",
    "gpt-3.5-turbo": "claude-3-haiku-20240307"
  }
}
```

**客户端代码:**
```python
import openai

# 配置客户端指向 ccany 服务器
client = openai.OpenAI(
    api_key="ccany-claude-proxy",  # 使用自定义密钥
    base_url="http://localhost:8082/v1"  # 指向 ccany 服务器
)

# 正常使用 OpenAI SDK，实际调用 Claude API
response = client.chat.completions.create(
    model="gpt-4",  # 会自动转换为 claude-3-sonnet-20240229
    messages=[
        {"role": "user", "content": "Hello, how are you?"}
    ],
    temperature=0.7,
    max_tokens=150
)

print(response.choices[0].message.content)
```

### 示例 2: Anthropic 客户端调用 Gemini API

**配置渠道:**
```json
{
  "name": "Gemini-Backend", 
  "provider": "gemini",
  "api_key": "your-gemini-api-key",
  "custom_key": "ccany-gemini-proxy",
  "base_url": "https://generativelanguage.googleapis.com/v1beta"
}
```

**客户端代码:**
```python
import anthropic

# 配置客户端指向 ccany 服务器
client = anthropic.Anthropic(
    api_key="ccany-gemini-proxy",  # 使用自定义密钥
    base_url="http://localhost:8082"  # 指向 ccany 服务器
)

# 使用 Anthropic SDK，实际调用 Gemini API
response = client.messages.create(
    model="claude-3-sonnet-20240229",  # 会转换为合适的 Gemini 模型
    max_tokens=100,
    messages=[
        {"role": "user", "content": "Explain quantum computing"}
    ]
)

print(response.content[0].text)
```

### 示例 3: 流式调用

**OpenAI 流式调用 → Claude 后端:**
```python
import openai

client = openai.OpenAI(
    api_key="ccany-claude-proxy",
    base_url="http://localhost:8082/v1"
)

# 流式调用
stream = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Write a poem about AI"}],
    stream=True,
    max_tokens=200
)

for chunk in stream:
    if chunk.choices[0].delta.content is not None:
        print(chunk.choices[0].delta.content, end="")
```

### 示例 4: 工具调用转换

**OpenAI 工具调用 → Anthropic 后端:**
```python
tools = [
    {
        "type": "function",
        "function": {
            "name": "get_weather",
            "description": "Get weather information",
            "parameters": {
                "type": "object",
                "properties": {
                    "location": {"type": "string", "description": "City name"}
                },
                "required": ["location"]
            }
        }
    }
]

response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "What's the weather in Tokyo?"}],
    tools=tools,
    tool_choice="auto"
)

# 系统自动转换 OpenAI 工具格式为 Claude 工具格式
```

## 🎯 高级功能

### 1. 多渠道负载均衡
```json
{
  "load_balancing": {
    "strategy": "round_robin",
    "health_check": true,
    "timeout": "30s"
  }
}
```

### 2. 自动故障转移
```json
{
  "failover": {
    "enabled": true,
    "retry_attempts": 3,
    "fallback_provider": "openai"
  }
}
```

### 3. 请求日志和监控
- 访问 `/monitoring` 查看系统状态
- 查看 `/logs` 了解请求详情
- 监控 API 使用统计

### 4. 模型映射管理
```json
{
  "models_mapping": {
    "gpt-4": "claude-3-opus-20240229",
    "gpt-4-turbo": "claude-3-sonnet-20240229", 
    "gpt-3.5-turbo": "claude-3-haiku-20240307",
    "gemini-pro": "gemini-1.5-pro",
    "gemini-flash": "gemini-1.5-flash"
  }
}
```

## 🔍 调试和故障排除

### 查看日志
```bash
# 启动时开启详细日志
LOG_LEVEL=debug go run cmd/server/main.go
```

### 测试 API 连通性
```bash
# 测试渠道健康状态
curl http://localhost:8082/health

# 测试特定渠道
curl -X POST http://localhost:8082/api/channels/test \
  -H "Content-Type: application/json" \
  -d '{"channel_id": "your-channel-id"}'
```

### 常见问题

1. **API 密钥错误**: 检查渠道配置中的 `api_key` 是否正确
2. **格式转换失败**: 查看日志了解具体转换错误
3. **超时问题**: 调整渠道的 `timeout` 配置
4. **模型不存在**: 检查 `models_mapping` 配置

## 📊 性能优化

1. **连接池**: 系统自动管理 HTTP 连接池
2. **请求缓存**: 可配置响应缓存
3. **并发限制**: 防止 API 限流
4. **智能重试**: 自动处理临时错误

## 🌟 总结

ccany 系统让您可以：
- ✅ 使用任意 AI SDK 调用任意 AI 服务
- ✅ 无需修改现有代码
- ✅ 统一管理多个 AI 提供商
- ✅ 自动处理格式转换和协议差异
- ✅ 获得完整的监控和日志功能

现在您可以享受真正的多 AI 提供商统一接入体验！🎉