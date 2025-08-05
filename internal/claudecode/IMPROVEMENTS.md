# Claude Code Router 改进实施方案

基于对 claude-code-router 项目的分析，我们为 ccany 项目实施了以下改进：

## 1. 增强的路由器配置系统

### 新增功能
- **RouterConfig** 结构体：提供了更灵活的路由配置选项
- **RouterConfigManager**：管理路由器配置，与数据库集成
- 配置通过管理后台进行管理，而非JSON文件
- 支持动态更新配置

### 配置选项
```go
type RouterConfig struct {
    Default          string // 默认模型
    Background       string // 后台任务模型（haiku）
    Think            string // 思考模式模型
    LongContext      string // 长上下文模型
    WebSearch        string // Web搜索模型
    LongContextThreshold int // 长上下文阈值
    EnableWebSearchDetection bool // 启用Web搜索检测
    EnableToolUseDetection   bool // 启用工具使用检测
    EnableDynamicRouting     bool // 启用动态路由
}
```

## 2. Web搜索工具检测

### 实现细节
- 新增 `WebSearchStrategy` 策略
- 自动检测 tools 中的 web_search 类型
- 为 Web 搜索任务分配专门的模型

```go
// 检测逻辑
- 检查 tool.Name == "web_search"
- 检查 tool type 字段
- 支持灵活的检测模式
```

## 3. 数据库驱动的配置管理

### 特性
- 所有配置存储在数据库中
- 通过管理后台 API 进行配置
- 支持实时更新，无需重启服务
- 配置加密存储（敏感信息）

### 配置键
```go
// Claude Code 配置键
KeyClaudeCodeUserID
KeyClaudeCodeNumStartups
KeyClaudeCodeAutoUpdaterStatus
KeyClaudeCodeHasCompletedOnboarding
KeyClaudeCodeLastOnboardingVersion
KeyClaudeCodeInstallationID
KeyClaudeCodeTelemetryEnabled
KeyClaudeCodeAnalyticsEnabled
KeyClaudeCodeCrashReportingEnabled

// Claude Code Router 配置键
KeyRouterDefault
KeyRouterBackground
KeyRouterThink
KeyRouterLongContext
KeyRouterWebSearch
KeyRouterLongContextThreshold
KeyRouterEnableWebSearchDetection
KeyRouterEnableToolUseDetection
KeyRouterEnableDynamicRouting
```

## 4. 增强的日志系统

### 特性
- 结构化日志（JSON 格式）
- 支持文件和控制台输出
- 环境变量控制
- 详细的路由决策日志
- 请求度量记录
- 工具执行日志

### 环境变量
- `CLAUDE_CODE_LOG=true` - 启用日志
- `CLAUDE_CODE_LOG_FILE=/path/to/log` - 日志文件路径
- `CLAUDE_CODE_LOG_LEVEL=info` - 日志级别

### 日志类型
1. **路由决策日志**
   - 原始模型
   - 路由后的模型
   - 路由原因
   - Token 计数
   - 其他元数据

2. **请求度量**
   - 请求 ID
   - 输入/输出 Token
   - 延迟
   - 状态
   - 缓存命中

3. **工具执行日志**
   - 工具名称
   - 执行时间
   - 成功/失败状态
   - 输入/输出大小

## 5. Token 计数优化

### 改进点
- 更准确的 Token 估算算法
- 支持工具定义的 Token 计算
- 支持复杂内容块（图片、工具使用、工具结果）

## 6. 策略优先级系统

### 路由策略优先级（从高到低）
1. 逗号分隔的模型列表（直接传递）
2. 长上下文处理（> 60K tokens）
3. Web 搜索检测
4. 后台模型（haiku）
5. 思考模式
6. 默认模型

## 使用示例

### 基本使用
```go
// 创建路由器
router := NewModelRouter(logger, "claude-3-5-sonnet", "claude-3-5-haiku")

// 路由请求
model := router.RouteModel(request)
```

### 通过管理后台配置
```bash
# 更新路由器配置
curl -X PUT http://localhost:8082/admin/config \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "router_default_model": "claude-3-5-sonnet-20241022",
    "router_background_model": "claude-3-5-haiku-20241022",
    "router_think_model": "claude-3-5-sonnet-20241022",
    "router_long_context_model": "claude-3-5-sonnet-20241022",
    "router_web_search_model": "claude-3-5-sonnet-20241022",
    "router_long_context_threshold": "60000",
    "router_enable_web_search_detection": "true",
    "router_enable_tool_use_detection": "true",
    "router_enable_dynamic_routing": "true"
  }'
```

## 架构优势

1. **简化的配置管理**
   - 移除了 JSON 文件依赖
   - 所有配置通过管理后台统一管理
   - 支持热更新

2. **更安全的扩展机制**
   - 移除了插件系统的安全风险
   - 路由逻辑内置，更加稳定
   - 通过策略模式实现扩展性

3. **更好的可维护性**
   - 代码结构更加清晰
   - 减少了外部依赖
   - 统一的配置管理接口

## 总结

通过借鉴 claude-code-router 的优秀设计，并结合 ccany 项目的实际需求，我们实现了：
- 更灵活的路由配置（通过数据库管理）
- 更智能的模型选择
- 更完善的监控日志
- 更安全的架构设计

这些改进让 ccany 的 Claude Code 支持更加完善、安全和易于管理。