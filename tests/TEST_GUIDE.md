# CCany 测试指南

本指南说明如何测试 CCany 的工具调用和 API 格式转换功能。

## 测试准备

### 1. 环境要求

- Go 1.21 或更高版本
- Python 3.6 或更高版本
- requests 库（Python）

```bash
# 安装 Python 依赖
pip3 install requests
```

### 2. 启动 CCany 服务器

在项目根目录运行：

```bash
# 方式1：直接运行
go run cmd/server/main.go

# 方式2：使用 Docker
docker-compose up -d

# 方式3：使用部署脚本
./scripts/deploy.sh start
```

### 3. 配置 API 渠道

1. 访问 http://localhost:8082
2. 使用管理员账户登录（如果是首次运行，访问 /setup 创建账户）
3. 在管理界面中配置至少一个 API 渠道：
   - **OpenAI 渠道**：配置 OpenAI API 密钥和基础 URL
   - **Claude 渠道**：配置 Claude API 密钥（如果有）
   - **本地模型**：配置 Ollama 或其他本地模型

### 4. 设置测试 API 密钥

在渠道配置中设置自定义密钥（例如：`test-api-key`），测试脚本将使用此密钥进行认证。

## 运行测试

### 方式1：使用测试运行脚本（推荐）

```bash
cd tests
./run_comprehensive_test.sh
```

### 方式2：直接运行 Python 脚本

```bash
# 综合测试
cd tests
python3 ccany_comprehensive_test.py test-api-key

# Gemini 专项测试
python3 test_gemini_conversion.py test-api-key
```

### 方式3：运行 Go 测试

```bash
# 运行所有测试
go test ./tests/...

# 运行特定测试
go test -v -run TestToolUse ./tests/
```

## 测试内容说明

### 综合测试 (ccany_comprehensive_test.py)

该脚本测试以下功能：

1. **健康检查**
   - 验证服务器是否正常运行

2. **API 格式转换**
   - Claude → OpenAI 格式转换
   - OpenAI → Claude 格式转换

3. **工具调用**
   - 基本工具调用（单个工具）
   - 复杂工具调用（多个工具）

4. **流式响应**
   - 验证 SSE 流式响应是否正确

5. **多模态输入**
   - 测试图像输入的处理

6. **错误处理**
   - 无效 JSON
   - 缺少必填字段
   - 无效 API 密钥

### Gemini 测试 (test_gemini_conversion.py)

专门测试 Gemini API 相关功能：

1. **Gemini 格式请求**
   - 标准 Gemini API 格式
   - 不同端点变体

2. **Gemini 流式响应**
   - streamGenerateContent 端点

3. **Gemini 工具调用**
   - 函数声明格式

4. **Gemini 多模态**
   - 图像输入处理

5. **格式转换**
   - Claude → Gemini 转换

## 测试结果

测试完成后会生成：

1. **控制台输出**：实时显示每个测试的结果
2. **测试报告**：JSON 格式的详细报告，保存为 `test_report_YYYYMMDD_HHMMSS.json`

## 常见问题

### 1. 服务器未运行

错误信息：`❌ CCany服务器未运行!`

解决方法：
- 确保服务器已启动
- 检查端口 8082 是否被占用
- 查看服务器日志是否有错误

### 2. API 密钥无效

错误信息：`401 Unauthorized` 或 `403 Forbidden`

解决方法：
- 检查是否已配置渠道
- 确认自定义密钥是否正确
- 验证后端 API 密钥是否有效

### 3. 请求超时

错误信息：`请求异常: timeout`

解决方法：
- 检查网络连接
- 增加超时时间
- 确认后端 API 是否可访问

### 4. 格式转换失败

错误信息：`响应格式不符合规范`

解决方法：
- 检查模型映射配置
- 确认请求格式是否正确
- 查看服务器日志获取详细错误信息

## 测试技巧

1. **分步测试**：先运行健康检查，确保服务器正常后再进行其他测试

2. **查看日志**：测试失败时，查看服务器日志获取更多信息：
   ```bash
   tail -f logs/app.log
   ```

3. **使用调试模式**：设置环境变量启用调试日志：
   ```bash
   LOG_LEVEL=debug go run cmd/server/main.go
   ```

4. **测试单个功能**：可以注释掉测试脚本中的部分测试，专注于特定功能

5. **使用 Postman**：对于复杂的测试场景，可以使用 Postman 或类似工具进行手动测试

## 扩展测试

如需添加更多测试用例，可以：

1. 修改现有测试脚本，添加新的测试方法
2. 创建新的测试脚本，专注于特定功能
3. 使用 Go 编写单元测试或集成测试

## 相关文档

- [README.md](../README.md) - 项目主文档
- [USAGE_GUIDE.md](../USAGE_GUIDE.md) - 使用指南
- [docs/DEPLOYMENT_GUIDE.md](../docs/DEPLOYMENT_GUIDE.md) - 部署指南