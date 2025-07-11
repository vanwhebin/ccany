# CCany - Makefile
# 简化 Docker 部署和管理操作

.PHONY: help build build-go build-python up up-go up-python down clean logs test

# 默认目标
help:
	@echo "CCany - 可用命令："
	@echo ""
	@echo "构建相关："
	@echo "  build         构建所有镜像"
	@echo "  build-go      构建 Go 版本镜像"
	@echo ""
	@echo "运行相关："
	@echo "  up            启动 Go 版本服务"
	@echo "  up-go         启动 Go 版本服务"
	@echo "  up-admin      启动 Go 版本 + 管理界面"
	@echo "  up-full       启动完整服务栈"
	@echo ""
	@echo "管理相关："
	@echo "  down          停止并移除所有容器"
	@echo "  clean         清理所有资源（容器、镜像、数据卷）"
	@echo "  logs          查看服务日志"
	@echo "  logs-go       查看 Go 版本日志"
	@echo ""
	@echo "开发相关："
	@echo "  test          运行测试"
	@echo "  dev-go        启动 Go 版本开发模式"

# 构建相关
build: build-go

build-go:
	@echo "构建 Go 版本镜像..."
	docker-compose build ccany-go

# 运行相关
up: up-go

up-go:
	@echo "启动 Go 版本服务..."
	docker-compose up -d ccany-go


up-admin:
	@echo "启动 Go 版本 + 管理界面..."
	docker-compose --profile admin up -d

up-full:
	@echo "启动完整服务栈..."
	docker-compose --profile admin --profile nginx --profile monitoring --profile cache up -d

up-monitoring:
	@echo "启动监控服务..."
	docker-compose --profile monitoring up -d

up-cache:
	@echo "启动缓存服务..."
	docker-compose --profile cache up -d

# 管理相关
down:
	@echo "停止所有服务..."
	docker-compose --profile admin --profile nginx --profile monitoring --profile cache down

clean:
	@echo "清理所有资源..."
	docker-compose --profile admin --profile nginx --profile monitoring --profile cache down -v --rmi all
	docker system prune -f

# 日志相关
logs:
	docker-compose logs -f

logs-go:
	docker-compose logs -f ccany-go


logs-admin:
	docker-compose logs -f adminer

logs-nginx:
	docker-compose logs -f nginx

logs-monitoring:
	docker-compose logs -f prometheus grafana

# 开发相关
test:
	@echo "运行 Go 测试..."
	go test ./tests/... -v

dev-go:
	@echo "启动 Go 开发模式..."
	go run cmd/server/main.go

# 健康检查
health:
	@echo "检查服务健康状态..."
	@curl -f http://localhost:8082/health && echo "Go 版本：健康" || echo "Go 版本：异常"

# 备份和恢复
backup:
	@echo "备份数据..."
	docker run --rm -v ccany_go_proxy_data:/data -v $(PWD):/backup alpine tar czf /backup/go-backup-$(shell date +%Y%m%d-%H%M%S).tar.gz -C /data .

restore-go:
	@echo "恢复 Go 版本数据..."
	@read -p "请输入备份文件路径: " backup_file; \
	docker run --rm -v ccany_go_proxy_data:/data -v $(PWD):/backup alpine tar xzf /backup/$$backup_file -C /data

# 环境准备
env:
	@echo "准备环境文件..."
	@if [ ! -f .env ]; then cp .env.example .env; echo "已创建 .env 文件，请根据需要修改配置"; fi

# 版本信息
version:
	@echo "项目版本信息："
	@echo "Go 版本: $(shell go version)"
	@echo "Docker 版本: $(shell docker --version)"
	@echo "Docker Compose 版本: $(shell docker-compose --version)"

# 快速启动命令
quick-start: env build up-go
	@echo "快速启动完成！"
	@echo "服务地址: http://localhost:8082"
	@echo "设置页面: http://localhost:8082/setup"