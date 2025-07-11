# 发布说明

## 自动构建和发布

本项目配置了GitHub Actions工作流，当创建新的Release时会自动构建和发布Docker镜像。

### 功能特性

- 🚀 **自动构建**: 创建Release时自动触发构建
- 🐳 **多平台支持**: 支持 `linux/amd64` 和 `linux/arm64` 架构
- 📦 **双平台发布**: 同时发布到 GitHub Container Registry 和 Docker Hub
- 🏷️ **版本标签**: 自动生成多种版本标签
- 🔧 **版本信息**: 在应用中显示版本号和构建时间

### 配置要求

#### GitHub Secrets

需要在GitHub仓库中配置以下Secrets：

1. **DOCKERHUB_USERNAME**: Docker Hub用户名
2. **DOCKERHUB_TOKEN**: Docker Hub访问令牌（推荐使用Access Token而非密码）

#### 获取Docker Hub Access Token

1. 登录到 [Docker Hub](https://hub.docker.com/)
2. 前往 Account Settings → Security
3. 点击 "New Access Token"
4. 输入描述并选择权限（推荐Read, Write, Delete）
5. 复制生成的Token并添加到GitHub Secrets

### 使用方法

#### 1. 创建Release

```bash
# 使用GitHub CLI
gh release create v1.0.0 --title "Release v1.0.0" --notes "发布说明"

# 或者在GitHub网页上创建Release
# 前往 Releases → Create a new release
```

#### 2. 自动构建

创建Release后，GitHub Actions会自动：

1. 检出代码
2. 设置Docker Buildx
3. 登录到容器注册表
4. 构建多架构镜像
5. 推送到GitHub Container Registry和Docker Hub

#### 3. 镜像标签

自动生成的标签包括：

- `latest` - 最新版本
- `v1.0.0` - 完整版本号
- `v1.0` - 主版本号+次版本号
- `v1` - 主版本号
- `sha-abc1234` - Git提交SHA

### 镜像仓库

构建完成后，镜像可从以下地址获取：

#### GitHub Container Registry
```bash
docker pull ghcr.io/your-username/your-repo:latest
```

#### Docker Hub
```bash
docker pull your-dockerhub-username/ccany:latest
```

### 版本信息显示

应用启动时会显示版本信息：

```
🚀 Claude-to-OpenAI API Proxy v1.0.0
🏗️  Built at: 2024-01-01T12:00:00Z
✅ Configuration loaded from database
```

Web界面中也会显示版本信息：
- 管理面板仪表板
- 首次设置页面

### 本地测试

可以本地构建和测试镜像：

```bash
# 构建镜像
docker build -t claude-proxy:dev \
  --build-arg VERSION=dev \
  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  .

# 运行容器
docker run -d --name claude-proxy \
  -p 8082:8082 \
  -e MASTER_KEY=your-master-key \
  claude-proxy:dev
```

### 故障排除

#### 构建失败

1. 检查Docker Hub凭据是否正确
2. 确认仓库权限设置
3. 查看GitHub Actions日志

#### 权限问题

确保GitHub Actions有足够权限：
- `contents: read` - 读取仓库代码
- `packages: write` - 写入GitHub Container Registry

#### 多架构构建问题

如果多架构构建失败，可能是由于：
- 依赖项不支持目标架构
- 构建时间过长导致超时

### 手动发布

如果需要手动发布，可以使用以下命令：

```bash
# 构建和推送到GitHub Container Registry
docker buildx build --platform linux/amd64,linux/arm64 \
  --build-arg VERSION=v1.0.0 \
  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  -t ghcr.io/your-username/your-repo:v1.0.0 \
  --push .

# 构建和推送到Docker Hub
docker buildx build --platform linux/amd64,linux/arm64 \
  --build-arg VERSION=v1.0.0 \
  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  -t your-dockerhub-username/ccany:v1.0.0 \
  --push .
```

### 注意事项

1. **版本号格式**: 推荐使用语义版本号（如v1.0.0）
2. **构建时间**: 构建可能需要10-20分钟，请耐心等待
3. **存储空间**: 多架构镜像占用更多存储空间
4. **网络限制**: 确保网络环境允许访问Docker Hub和GitHub

### 更新日志

- v1.0.0: 初始版本，支持自动构建和发布
- 支持多架构构建（amd64, arm64）
- 版本信息显示功能
- 自动更新Docker Hub描述