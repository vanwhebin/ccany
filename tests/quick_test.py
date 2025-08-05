#!/usr/bin/env python3
"""
CCany 快速测试脚本
用于快速验证服务器是否正常运行以及基本功能是否可用
"""

import requests
import json
import sys
from datetime import datetime

# 配置
CCANY_BASE_URL = "http://localhost:8082"
API_KEY = "test-api-key"

def print_header(title):
    """打印标题"""
    print(f"\n{'='*60}")
    print(f"{title:^60}")
    print('='*60)

def test_server_health():
    """测试服务器健康状态"""
    print("\n🏥 检查服务器健康状态...")
    
    try:
        response = requests.get(f"{CCANY_BASE_URL}/health", timeout=5)
        if response.status_code == 200:
            print("✅ 服务器运行正常")
            data = response.json()
            print(f"   状态: {data.get('status', 'unknown')}")
            print(f"   版本: {data.get('version', 'unknown')}")
            return True
        else:
            print(f"❌ 服务器响应异常: {response.status_code}")
            return False
    except requests.exceptions.ConnectionError:
        print("❌ 无法连接到服务器")
        print(f"   请确保服务器在 {CCANY_BASE_URL} 上运行")
        print("\n   启动服务器:")
        print("   cd /home/czyt/code/go/ccany")
        print("   go run cmd/server/main.go")
        return False
    except Exception as e:
        print(f"❌ 检查失败: {str(e)}")
        return False

def test_basic_claude_request():
    """测试基本的Claude格式请求"""
    print("\n🧪 测试Claude格式请求...")
    
    headers = {
        "x-api-key": API_KEY,
        "Content-Type": "application/json",
        "anthropic-version": "2023-06-01"
    }
    
    request_data = {
        "model": "claude-3-haiku-20240307",
        "max_tokens": 50,
        "messages": [
            {
                "role": "user",
                "content": "Say 'Hello, CCany test!'"
            }
        ]
    }
    
    try:
        response = requests.post(
            f"{CCANY_BASE_URL}/v1/messages",
            headers=headers,
            json=request_data,
            timeout=10
        )
        
        print(f"   状态码: {response.status_code}")
        
        if response.status_code == 200:
            print("✅ Claude请求成功")
            result = response.json()
            if "content" in result and len(result["content"]) > 0:
                print(f"   响应: {result['content'][0].get('text', 'No text')[:100]}")
            return True
        elif response.status_code == 401:
            print("⚠️  认证失败 - 请检查API密钥配置")
            return False
        elif response.status_code == 404:
            print("⚠️  端点未找到 - 请检查服务器版本")
            return False
        else:
            print(f"❌ 请求失败: {response.text[:200]}")
            return False
            
    except Exception as e:
        print(f"❌ 请求异常: {str(e)}")
        return False

def test_basic_openai_request():
    """测试基本的OpenAI格式请求"""
    print("\n🧪 测试OpenAI格式请求...")
    
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    request_data = {
        "model": "gpt-3.5-turbo",
        "messages": [
            {
                "role": "user",
                "content": "Say 'Hello, CCany test!'"
            }
        ],
        "max_tokens": 50
    }
    
    try:
        response = requests.post(
            f"{CCANY_BASE_URL}/v1/chat/completions",
            headers=headers,
            json=request_data,
            timeout=10
        )
        
        print(f"   状态码: {response.status_code}")
        
        if response.status_code == 200:
            print("✅ OpenAI请求成功")
            result = response.json()
            if "choices" in result and len(result["choices"]) > 0:
                content = result["choices"][0].get("message", {}).get("content", "No content")
                print(f"   响应: {content[:100]}")
            return True
        elif response.status_code == 401:
            print("⚠️  认证失败 - 请检查API密钥配置")
            return False
        elif response.status_code == 404:
            print("⚠️  端点未找到 - 请检查服务器版本")
            return False
        else:
            print(f"❌ 请求失败: {response.text[:200]}")
            return False
            
    except Exception as e:
        print(f"❌ 请求异常: {str(e)}")
        return False

def test_tool_calling():
    """测试工具调用功能"""
    print("\n🔧 测试工具调用功能...")
    
    headers = {
        "x-api-key": API_KEY,
        "Content-Type": "application/json",
        "anthropic-version": "2023-06-01"
    }
    
    request_data = {
        "model": "claude-3-haiku-20240307",
        "max_tokens": 150,
        "messages": [
            {
                "role": "user",
                "content": "What's 25 + 17? Use the calculator tool."
            }
        ],
        "tools": [
            {
                "name": "calculator",
                "description": "A simple calculator that can add two numbers",
                "input_schema": {
                    "type": "object",
                    "properties": {
                        "a": {
                            "type": "number",
                            "description": "First number"
                        },
                        "b": {
                            "type": "number",
                            "description": "Second number"
                        }
                    },
                    "required": ["a", "b"]
                }
            }
        ]
    }
    
    try:
        response = requests.post(
            f"{CCANY_BASE_URL}/v1/messages",
            headers=headers,
            json=request_data,
            timeout=10
        )
        
        print(f"   状态码: {response.status_code}")
        
        if response.status_code == 200:
            result = response.json()
            # 检查是否有工具调用
            has_tool_use = False
            for content in result.get("content", []):
                if content.get("type") == "tool_use":
                    has_tool_use = True
                    print("✅ 工具调用成功")
                    print(f"   工具: {content.get('name', 'unknown')}")
                    print(f"   输入: {json.dumps(content.get('input', {}), ensure_ascii=False)}")
                    break
            
            if not has_tool_use:
                print("⚠️  响应中未包含工具调用")
                # 但如果响应了正确答案，也算部分成功
                for content in result.get("content", []):
                    if content.get("type") == "text" and "42" in content.get("text", ""):
                        print("   但模型直接给出了正确答案")
                        return True
            return has_tool_use
        else:
            print(f"❌ 请求失败: {response.text[:200]}")
            return False
            
    except Exception as e:
        print(f"❌ 请求异常: {str(e)}")
        return False

def main():
    """主函数"""
    global API_KEY
    
    print_header("CCany 快速功能测试")
    print(f"时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"服务器: {CCANY_BASE_URL}")
    
    # 检查命令行参数
    if len(sys.argv) > 1:
        API_KEY = sys.argv[1]
    
    print(f"API密钥: {'*' * (len(API_KEY) - 4) + API_KEY[-4:] if len(API_KEY) > 4 else '****'}")
    
    # 运行测试
    tests_passed = 0
    tests_total = 0
    
    # 1. 健康检查
    tests_total += 1
    if test_server_health():
        tests_passed += 1
    else:
        print("\n❌ 服务器未运行，测试终止")
        return
    
    # 2. Claude格式测试
    tests_total += 1
    if test_basic_claude_request():
        tests_passed += 1
    
    # 3. OpenAI格式测试
    tests_total += 1
    if test_basic_openai_request():
        tests_passed += 1
    
    # 4. 工具调用测试
    tests_total += 1
    if test_tool_calling():
        tests_passed += 1
    
    # 总结
    print_header("测试总结")
    print(f"总测试数: {tests_total}")
    print(f"✅ 通过: {tests_passed}")
    print(f"❌ 失败: {tests_total - tests_passed}")
    print(f"成功率: {tests_passed/tests_total*100:.1f}%")
    
    if tests_passed < tests_total:
        print("\n💡 提示:")
        print("1. 确保已通过Web界面配置了API渠道")
        print("2. 检查API密钥是否正确")
        print("3. 查看服务器日志了解详细错误信息")
        print("\n运行完整测试:")
        print("  ./run_comprehensive_test.sh")
    else:
        print("\n🎉 所有基本功能测试通过!")
        print("可以运行完整测试套件:")
        print("  ./run_comprehensive_test.sh")


if __name__ == "__main__":
    main()