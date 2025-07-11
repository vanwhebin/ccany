# 部署指南

本文档提供了CCany Go版本的详细部署指南。

## 系统要求

### 最低要求
- Go 1.21 或更高版本
- 2GB RAM
- 1GB 可用磁盘空间
- Linux/macOS/Windows

### 推荐要求
- Go 1.24 或更高版本
- 4GB RAM
- 5GB 可用磁盘空间
- Linux (Ubuntu 20.04+/CentOS 8+)

## 部署方式

### 1. 直接运行

#### 安装和配置
```bash
# 克隆项目
git clone https://github.com/yourusername/ccany.git
cd ccany

# 安装依赖
go mod tidy

# 可选：复制配置文件设置系统环境变量
cp .env.example .env
# 编辑 .env 文件设置数据存储目录等（可选）

# 构建项目
go build -o ccany cmd/server/main.go

# 运行项目
./ccany
```

#### 初始化设置
```bash
# 方式1：使用脚本初始化管理员
go run scripts/init_admin.go

# 方式2：通过Web界面初始化
# 访问 http://localhost:8082/setup
# 创建管理员账户并配置API密钥
```

### 2. 使用systemd服务（Linux）

#### 创建服务文件
```bash
sudo nano /etc/systemd/system/ccany.service
```

#### 服务配置
```ini
[Unit]
Description=CCany Go Server
After=network.target

[Service]
Type=simple
User=ccany
Group=ccany
WorkingDirectory=/opt/ccany
ExecStart=/opt/ccany/ccany
Restart=always
RestartSec=10
Environment=GIN_MODE=release
EnvironmentFile=/opt/ccany/.env

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/ccany/data

[Install]
WantedBy=multi-user.target
```

#### 部署步骤
```bash
# 创建用户
sudo useradd -r -s /bin/false ccany

# 创建目录
sudo mkdir -p /opt/ccany/data

# 复制文件
sudo cp ccany /opt/ccany/
sudo cp .env /opt/ccany/
sudo cp -r web /opt/ccany/

# 设置权限
sudo chown -R ccany:ccany /opt/ccany
sudo chmod +x /opt/ccany/ccany

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable ccany
sudo systemctl start ccany
```

### 3. 使用Docker

#### Dockerfile
```dockerfile
FROM golang:1.24-alpine AS builder

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o ccany cmd/server/main.go

# 运行阶段
FROM alpine:latest

# 安装依赖
RUN apk --no-cache add ca-certificates sqlite

# 创建用户
RUN addgroup -g 1001 -S claude && adduser -u 1001 -S claude -G claude

# 创建目录
RUN mkdir -p /app/data /app/web

# 复制文件
COPY --from=builder /app/ccany /app/
COPY --from=builder /app/web /app/web/

# 设置权限
RUN chown -R claude:claude /app

# 切换用户
USER claude

# 设置工作目录
WORKDIR /app

# 暴露端口
EXPOSE 8082

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# 启动应用
CMD ["./ccany"]
```

#### docker-compose.yml
```yaml
version: '3.8'

services:
  ccany:
    build: .
    ports:
      - "8082:8082"
    environment:
      - GIN_MODE=release
      - HOST=0.0.0.0
      - PORT=8082
      - DATABASE_URL=sqlite3://./data/claude_proxy.db
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - OPENAI_BASE_URL=${OPENAI_BASE_URL:-https://api.openai.com/v1}
      - BIG_MODEL=${BIG_MODEL:-gpt-4o}
      - SMALL_MODEL=${SMALL_MODEL:-gpt-4o-mini}
    volumes:
      - ./data:/app/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8082/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

#### 运行Docker容器
```bash
# 构建和运行
docker-compose up -d

# 查看日志
docker-compose logs -f ccany

# 停止服务
docker-compose down
```

### 4. 使用Kubernetes

#### deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ccany
  labels:
    app: ccany
spec:
  replicas: 2
  selector:
    matchLabels:
      app: ccany
  template:
    metadata:
      labels:
        app: ccany
    spec:
      containers:
      - name: ccany
        image: ccany:latest
        ports:
        - containerPort: 8082
        env:
        - name: GIN_MODE
          value: "release"
        - name: HOST
          value: "0.0.0.0"
        - name: PORT
          value: "8082"
        - name: DATABASE_URL
          value: "sqlite3://./data/claude_proxy.db"
        - name: CLAUDE_PROXY_MASTER_KEY
          valueFrom:
            secretKeyRef:
              name: ccany-secret
              key: master-key
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: ccany-secret
              key: jwt-secret
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8082
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: data
          mountPath: /app/data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: ccany-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: ccany-service
spec:
  selector:
    app: ccany
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8082
  type: LoadBalancer
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ccany-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
---
apiVersion: v1
kind: Secret
metadata:
  name: ccany-secret
type: Opaque
stringData:
  master-key: "your-secure-master-key-here"
  jwt-secret: "your-jwt-secret-key-here"
  jwt-secret: "your-jwt-secret-key-here"
```

#### 部署到Kubernetes
```bash
# 应用配置
kubectl apply -f deployment.yaml

# 检查部署状态
kubectl get pods -l app=ccany
kubectl get svc ccany-service

# 查看日志
kubectl logs -l app=ccany -f
```

## 环境变量配置

### 系统环境变量（可选）
```env
# 数据存储目录
CLAUDE_PROXY_DATA_PATH=./data

# 主密钥（生产环境建议设置）
CLAUDE_PROXY_MASTER_KEY=your-secure-master-key

# JWT密钥（生产环境建议设置）
JWT_SECRET=your-jwt-secret-key

# Go运行时配置
GIN_MODE=release
GOGC=100
```

### 后台配置管理
所有API和服务配置现在都通过Web管理界面进行配置，包括：

**API配置:**
- OpenAI API密钥和基础URL
- Claude API密钥和基础URL
- Azure API版本

**模型配置:**
- 大模型设置
- 小模型设置

**服务器配置:**
- 服务器主机和端口
- 日志级别

**性能配置:**
- Token限制
- 请求超时
- 重试次数
- 温度参数

**注意:** 首次部署后，访问 `/setup` 页面进行初始化配置。

## 安全配置

### 1. 数据库安全
```bash
# 设置数据库文件权限
chmod 600 data/claude_proxy.db
chown ccany:ccany data/claude_proxy.db
```

### 2. API密钥安全
```bash
# 使用环境变量文件
chmod 600 .env
chown ccany:ccany .env

# 或使用系统环境变量
export OPENAI_API_KEY="your-key-here"
```

### 3. 网络安全
```bash
# 使用防火墙限制访问
ufw allow from trusted-ip to any port 8082
ufw deny 8082
```

### 4. HTTPS配置
#### 使用Nginx反向代理
```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    location /v1/messages {
        proxy_pass http://localhost:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 300;
        proxy_connect_timeout 300;
        proxy_send_timeout 300;
    }
}
```

## 监控和日志

### 1. 日志配置
```bash
# 设置日志轮转
cat > /etc/logrotate.d/ccany << EOF
/var/log/ccany/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 ccany ccany
    postrotate
        systemctl reload ccany
    endscript
}
EOF
```

### 2. 系统监控
```bash
# 添加监控脚本
cat > /usr/local/bin/ccany-monitor.sh << 'EOF'
#!/bin/bash
# 检查服务状态
if ! systemctl is-active --quiet ccany; then
    echo "Claude Proxy服务未运行"
    systemctl start ccany
fi

# 检查端口
if ! netstat -tlnp | grep -q ":8082"; then
    echo "端口8082未监听"
fi

# 检查健康状态
if ! curl -f http://localhost:8082/health > /dev/null 2>&1; then
    echo "健康检查失败"
fi
EOF

chmod +x /usr/local/bin/ccany-monitor.sh

# 添加到crontab
echo "*/5 * * * * /usr/local/bin/ccany-monitor.sh" | crontab -
```

### 3. 性能监控
```bash
# 使用htop监控系统资源
htop -u ccany

# 监控应用日志
tail -f /var/log/ccany/app.log

# 监控数据库文件大小
du -sh /opt/ccany/data/claude_proxy.db
```

## 故障排除

### 1. 常见问题

#### 服务启动失败
```bash
# 检查日志
journalctl -u ccany -f

# 检查配置文件
go run cmd/server/main.go --check-config

# 检查端口占用
netstat -tlnp | grep 8082
```

#### 数据库连接失败
```bash
# 检查数据库文件权限
ls -la data/claude_proxy.db

# 检查数据库文件完整性
sqlite3 data/claude_proxy.db ".schema"
```

#### API请求失败
```bash
# 测试API连接
curl -X POST http://localhost:8082/v1/messages/count_tokens \
  -H "Content-Type: application/json" \
  -d '{"model": "claude-3-5-sonnet-20241022", "messages": [{"role": "user", "content": "test"}]}'
```

### 2. 性能调优

#### 内存优化
```bash
# 设置Go垃圾回收
export GOGC=100

# 限制内存使用
ulimit -v 2097152  # 2GB
```

#### 并发优化
```bash
# 调整文件描述符限制
ulimit -n 65536

# 在systemd服务中设置
echo "LimitNOFILE=65536" >> /etc/systemd/system/ccany.service
```

### 3. 备份和恢复

#### 数据备份
```bash
# 备份数据库
cp data/claude_proxy.db data/claude_proxy.db.backup.$(date +%Y%m%d)

# 自动备份脚本
cat > /usr/local/bin/ccany-backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/backup/ccany"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR
cp /opt/ccany/data/claude_proxy.db $BACKUP_DIR/claude_proxy_$DATE.db

# 保留最近7天的备份
find $BACKUP_DIR -name "claude_proxy_*.db" -mtime +7 -delete
EOF

chmod +x /usr/local/bin/ccany-backup.sh

# 添加到crontab
echo "0 2 * * * /usr/local/bin/ccany-backup.sh" | crontab -
```

#### 恢复数据
```bash
# 停止服务
systemctl stop ccany

# 恢复数据库
cp data/claude_proxy.db.backup.YYYYMMDD data/claude_proxy.db

# 启动服务
systemctl start ccany
```

## 升级指南

### 1. 准备升级
```bash
# 备份数据
/usr/local/bin/ccany-backup.sh

# 停止服务
systemctl stop ccany
```

### 2. 执行升级
```bash
# 下载新版本
wget https://github.com/yourusername/ccany/releases/latest/download/ccany

# 替换可执行文件
mv ccany /opt/ccany/ccany.new
chmod +x /opt/ccany/ccany.new
mv /opt/ccany/ccany /opt/ccany/ccany.old
mv /opt/ccany/ccany.new /opt/ccany/ccany

# 更新Web文件
cp -r web/* /opt/ccany/web/
```

### 3. 完成升级
```bash
# 启动服务
systemctl start ccany

# 检查状态
systemctl status ccany
curl http://localhost:8082/health
```

## 联系支持

如果在部署过程中遇到问题，请：

1. 查看日志文件
2. 检查GitHub Issues
3. 提交新的Issue并包含详细的错误信息
4. 加入讨论社区

---

更多详细信息请参考项目文档：[README.md](README.md)