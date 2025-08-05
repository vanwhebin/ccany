#!/bin/bash

# CCany 综合测试运行脚本

echo "🚀 CCany 综合测试运行器"
echo "========================"

# 检查Python是否安装
if ! command -v python3 &> /dev/null; then
    echo "❌ 错误: 需要安装 Python 3"
    echo "   请运行: sudo apt-get install python3 python3-pip"
    exit 1
fi

# 检查requests库是否安装
if ! python3 -c "import requests" &> /dev/null; then
    echo "📦 安装 requests 库..."
    pip3 install requests
fi

# 检查服务器是否运行
echo "🏥 检查CCany服务器状态..."
if ! curl -s http://localhost:8082/health > /dev/null; then
    echo "❌ CCany服务器未运行!"
    echo ""
    echo "请在另一个终端中启动服务器:"
    echo "  cd /home/czyt/code/go/ccany"
    echo "  go run cmd/server/main.go"
    echo ""
    echo "或使用Docker:"
    echo "  docker-compose up -d"
    echo ""
    read -p "服务器启动后，按Enter继续..."
fi

# 获取API密钥
API_KEY="${CCANY_API_KEY:-test-api-key}"

echo ""
echo "⚙️ 测试配置:"
echo "  - 服务器地址: http://localhost:8082"
echo "  - API密钥: ${API_KEY}"
echo ""
echo "📝 注意: 请确保已通过Web界面配置了至少一个API渠道"
echo "        访问 http://localhost:8082 进行配置"
echo ""

read -p "准备好开始测试了吗? (y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "测试已取消"
    exit 0
fi

# 运行测试
echo ""
echo "🧪 开始运行测试..."
echo "===================="

# 切换到tests目录
cd "$(dirname "$0")"

# 运行Python测试脚本
python3 ccany_comprehensive_test.py "$API_KEY"

echo ""
echo "✅ 测试完成!"
echo ""
echo "查看测试报告:"
echo "  ls -la test_report_*.json"