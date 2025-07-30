#!/usr/bin/env python3
"""
CCany 综合测试脚本
测试工具调用和多种API格式转换功能
"""

import requests
import json
import time
import base64
import sys
from typing import Dict, Any, List, Optional
from datetime import datetime

# 配置
CCANY_BASE_URL = "http://localhost:8082"
API_KEY = "test-api-key"  # 需要替换为实际配置的API密钥

# 测试结果存储
test_results = []

class CCanyTester:
    """CCany 功能测试类"""
    
    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url
        self.api_key = api_key
        self.session = requests.Session()
        
    def log_result(self, test_name: str, success: bool, message: str, details: Optional[Dict] = None):
        """记录测试结果"""
        result = {
            "test_name": test_name,
            "success": success,
            "message": message,
            "timestamp": datetime.now().isoformat(),
            "details": details or {}
        }
        test_results.append(result)
        
        # 打印结果
        status = "✅ 成功" if success else "❌ 失败"
        print(f"\n{status} - {test_name}")
        print(f"   {message}")
        if details and not success:
            print(f"   详情: {json.dumps(details, indent=2, ensure_ascii=False)}")
    
    def test_health_check(self):
        """测试健康检查端点"""
        print("\n🏥 测试健康检查...")
        
        try:
            response = self.session.get(f"{self.base_url}/health", timeout=10)
            if response.status_code == 200:
                self.log_result("健康检查", True, "服务器运行正常", response.json())
                return True
            else:
                self.log_result("健康检查", False, f"服务器响应异常: {response.status_code}", {"response": response.text})
                return False
        except Exception as e:
            self.log_result("健康检查", False, f"无法连接服务器: {str(e)}")
            return False
    
    def test_claude_to_openai_conversion(self):
        """测试 Claude 到 OpenAI 的 API 格式转换"""
        print("\n🔄 测试 Claude → OpenAI 格式转换...")
        
        headers = {
            "x-api-key": self.api_key,
            "Content-Type": "application/json",
            "anthropic-version": "2023-06-01"
        }
        
        # Claude 格式的请求
        claude_request = {
            "model": "claude-3-haiku-20240307",
            "max_tokens": 100,
            "messages": [
                {
                    "role": "user",
                    "content": "请用一句话介绍你自己。"
                }
            ],
            "temperature": 0.7
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/v1/messages",
                headers=headers,
                json=claude_request,
                timeout=30
            )
            
            if response.status_code == 200:
                result = response.json()
                # 验证响应格式是否符合 Claude 规范
                if all(key in result for key in ["id", "type", "role", "content", "model", "usage"]):
                    self.log_result("Claude到OpenAI转换", True, "格式转换成功", result)
                else:
                    self.log_result("Claude到OpenAI转换", False, "响应格式不符合Claude规范", result)
            else:
                self.log_result("Claude到OpenAI转换", False, f"请求失败: {response.status_code}", {"response": response.text})
                
        except Exception as e:
            self.log_result("Claude到OpenAI转换", False, f"请求异常: {str(e)}")
    
    def test_openai_to_claude_conversion(self):
        """测试 OpenAI 到 Claude 的 API 格式转换"""
        print("\n🔄 测试 OpenAI → Claude 格式转换...")
        
        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json"
        }
        
        # OpenAI 格式的请求
        openai_request = {
            "model": "gpt-3.5-turbo",
            "messages": [
                {
                    "role": "system",
                    "content": "你是一个有帮助的助手。"
                },
                {
                    "role": "user",
                    "content": "你好，请简单介绍一下自己。"
                }
            ],
            "max_tokens": 100,
            "temperature": 0.7
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/v1/chat/completions",
                headers=headers,
                json=openai_request,
                timeout=30
            )
            
            if response.status_code == 200:
                result = response.json()
                # 验证响应格式是否符合 OpenAI 规范
                if all(key in result for key in ["id", "object", "created", "model", "choices", "usage"]):
                    self.log_result("OpenAI到Claude转换", True, "格式转换成功", result)
                else:
                    self.log_result("OpenAI到Claude转换", False, "响应格式不符合OpenAI规范", result)
            else:
                self.log_result("OpenAI到Claude转换", False, f"请求失败: {response.status_code}", {"response": response.text})
                
        except Exception as e:
            self.log_result("OpenAI到Claude转换", False, f"请求异常: {str(e)}")
    
    def test_tool_calling_basic(self):
        """测试基本工具调用功能"""
        print("\n🔧 测试基本工具调用...")
        
        headers = {
            "x-api-key": self.api_key,
            "Content-Type": "application/json",
            "anthropic-version": "2023-06-01"
        }
        
        # 包含工具定义的 Claude 请求
        tool_request = {
            "model": "claude-3-haiku-20240307",
            "max_tokens": 200,
            "messages": [
                {
                    "role": "user",
                    "content": "现在的天气怎么样？请使用天气查询工具。"
                }
            ],
            "tools": [
                {
                    "name": "get_weather",
                    "description": "获取指定地点的天气信息",
                    "input_schema": {
                        "type": "object",
                        "properties": {
                            "location": {
                                "type": "string",
                                "description": "需要查询天气的地点"
                            },
                            "unit": {
                                "type": "string",
                                "enum": ["celsius", "fahrenheit"],
                                "description": "温度单位"
                            }
                        },
                        "required": ["location"]
                    }
                }
            ]
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/v1/messages",
                headers=headers,
                json=tool_request,
                timeout=30
            )
            
            if response.status_code == 200:
                result = response.json()
                # 检查是否有工具调用
                has_tool_use = False
                if "content" in result:
                    for content_block in result.get("content", []):
                        if content_block.get("type") == "tool_use":
                            has_tool_use = True
                            break
                
                if has_tool_use:
                    self.log_result("基本工具调用", True, "工具调用成功", result)
                else:
                    self.log_result("基本工具调用", False, "响应中未包含工具调用", result)
            else:
                self.log_result("基本工具调用", False, f"请求失败: {response.status_code}", {"response": response.text})
                
        except Exception as e:
            self.log_result("基本工具调用", False, f"请求异常: {str(e)}")
    
    def test_tool_calling_complex(self):
        """测试复杂工具调用（多工具）"""
        print("\n🔧 测试复杂工具调用...")
        
        headers = {
            "x-api-key": self.api_key,
            "Content-Type": "application/json",
            "anthropic-version": "2023-06-01"
        }
        
        # 包含多个工具定义的请求
        multi_tool_request = {
            "model": "claude-3-haiku-20240307",
            "max_tokens": 300,
            "messages": [
                {
                    "role": "user",
                    "content": "请帮我查询北京的天气，并计算如果温度是摄氏25度，转换成华氏度是多少。"
                }
            ],
            "tools": [
                {
                    "name": "get_weather",
                    "description": "获取指定地点的天气信息",
                    "input_schema": {
                        "type": "object",
                        "properties": {
                            "location": {
                                "type": "string",
                                "description": "需要查询天气的地点"
                            }
                        },
                        "required": ["location"]
                    }
                },
                {
                    "name": "convert_temperature",
                    "description": "在摄氏度和华氏度之间转换温度",
                    "input_schema": {
                        "type": "object",
                        "properties": {
                            "temperature": {
                                "type": "number",
                                "description": "要转换的温度值"
                            },
                            "from_unit": {
                                "type": "string",
                                "enum": ["celsius", "fahrenheit"],
                                "description": "源温度单位"
                            },
                            "to_unit": {
                                "type": "string",
                                "enum": ["celsius", "fahrenheit"],
                                "description": "目标温度单位"
                            }
                        },
                        "required": ["temperature", "from_unit", "to_unit"]
                    }
                }
            ]
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/v1/messages",
                headers=headers,
                json=multi_tool_request,
                timeout=30
            )
            
            if response.status_code == 200:
                result = response.json()
                # 统计工具调用次数
                tool_calls = 0
                if "content" in result:
                    for content_block in result.get("content", []):
                        if content_block.get("type") == "tool_use":
                            tool_calls += 1
                
                if tool_calls >= 1:
                    self.log_result("复杂工具调用", True, f"成功调用了 {tool_calls} 个工具", result)
                else:
                    self.log_result("复杂工具调用", False, "响应中未包含工具调用", result)
            else:
                self.log_result("复杂工具调用", False, f"请求失败: {response.status_code}", {"response": response.text})
                
        except Exception as e:
            self.log_result("复杂工具调用", False, f"请求异常: {str(e)}")
    
    def test_streaming_response(self):
        """测试流式响应的格式转换"""
        print("\n🌊 测试流式响应...")
        
        headers = {
            "x-api-key": self.api_key,
            "Content-Type": "application/json",
            "anthropic-version": "2023-06-01",
            "Accept": "text/event-stream"
        }
        
        # 流式请求
        stream_request = {
            "model": "claude-3-haiku-20240307",
            "max_tokens": 100,
            "stream": True,
            "messages": [
                {
                    "role": "user",
                    "content": "请从1数到5，每个数字单独一行。"
                }
            ]
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/v1/messages",
                headers=headers,
                json=stream_request,
                stream=True,
                timeout=30
            )
            
            if response.status_code == 200:
                events_received = 0
                content_parts = []
                
                for line in response.iter_lines():
                    if line:
                        line_str = line.decode('utf-8')
                        if line_str.startswith('data: '):
                            events_received += 1
                            try:
                                data = json.loads(line_str[6:])
                                # 收集内容
                                if data.get("type") == "content_block_delta":
                                    delta = data.get("delta", {})
                                    if "text" in delta:
                                        content_parts.append(delta["text"])
                            except json.JSONDecodeError:
                                pass
                
                if events_received > 0:
                    self.log_result("流式响应", True, f"成功接收 {events_received} 个事件", {
                        "events": events_received,
                        "content": "".join(content_parts)
                    })
                else:
                    self.log_result("流式响应", False, "未接收到流式事件")
            else:
                self.log_result("流式响应", False, f"请求失败: {response.status_code}", {"response": response.text})
                
        except Exception as e:
            self.log_result("流式响应", False, f"请求异常: {str(e)}")
    
    def test_multimodal_input(self):
        """测试多模态输入的格式转换"""
        print("\n🖼️ 测试多模态输入...")
        
        headers = {
            "x-api-key": self.api_key,
            "Content-Type": "application/json",
            "anthropic-version": "2023-06-01"
        }
        
        # 创建一个简单的测试图片（1x1像素的红色PNG）
        test_image_base64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
        
        # 多模态请求
        multimodal_request = {
            "model": "claude-3-haiku-20240307",
            "max_tokens": 100,
            "messages": [
                {
                    "role": "user",
                    "content": [
                        {
                            "type": "text",
                            "text": "这张图片是什么颜色？"
                        },
                        {
                            "type": "image",
                            "source": {
                                "type": "base64",
                                "media_type": "image/png",
                                "data": test_image_base64
                            }
                        }
                    ]
                }
            ]
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/v1/messages",
                headers=headers,
                json=multimodal_request,
                timeout=30
            )
            
            if response.status_code == 200:
                result = response.json()
                self.log_result("多模态输入", True, "多模态请求成功", result)
            else:
                self.log_result("多模态输入", False, f"请求失败: {response.status_code}", {"response": response.text})
                
        except Exception as e:
            self.log_result("多模态输入", False, f"请求异常: {str(e)}")
    
    def test_error_handling(self):
        """测试错误处理和边界情况"""
        print("\n⚠️ 测试错误处理...")
        
        # 测试1: 无效的JSON
        print("  - 测试无效JSON...")
        try:
            response = self.session.post(
                f"{self.base_url}/v1/messages",
                headers={"Content-Type": "application/json"},
                data="invalid json",
                timeout=10
            )
            
            if response.status_code == 400:
                self.log_result("错误处理-无效JSON", True, "正确返回400错误")
            else:
                self.log_result("错误处理-无效JSON", False, f"期望400，实际返回{response.status_code}")
        except Exception as e:
            self.log_result("错误处理-无效JSON", False, f"请求异常: {str(e)}")
        
        # 测试2: 缺少必填字段
        print("  - 测试缺少必填字段...")
        headers = {
            "x-api-key": self.api_key,
            "Content-Type": "application/json"
        }
        
        invalid_request = {
            # 缺少 model 字段
            "messages": [{"role": "user", "content": "test"}]
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/v1/messages",
                headers=headers,
                json=invalid_request,
                timeout=10
            )
            
            if response.status_code == 400:
                self.log_result("错误处理-缺少字段", True, "正确返回400错误")
            else:
                self.log_result("错误处理-缺少字段", False, f"期望400，实际返回{response.status_code}")
        except Exception as e:
            self.log_result("错误处理-缺少字段", False, f"请求异常: {str(e)}")
        
        # 测试3: 无效的API密钥
        print("  - 测试无效API密钥...")
        invalid_headers = {
            "x-api-key": "invalid-key",
            "Content-Type": "application/json"
        }
        
        valid_request = {
            "model": "claude-3-haiku-20240307",
            "max_tokens": 10,
            "messages": [{"role": "user", "content": "test"}]
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/v1/messages",
                headers=invalid_headers,
                json=valid_request,
                timeout=10
            )
            
            if response.status_code in [401, 403]:
                self.log_result("错误处理-无效密钥", True, f"正确返回{response.status_code}错误")
            else:
                self.log_result("错误处理-无效密钥", False, f"期望401/403，实际返回{response.status_code}")
        except Exception as e:
            self.log_result("错误处理-无效密钥", False, f"请求异常: {str(e)}")
    
    def run_all_tests(self):
        """运行所有测试"""
        print("\n🚀 开始运行CCany综合测试...")
        print("=" * 60)
        
        # 首先检查服务器健康状态
        if not self.test_health_check():
            print("\n❌ 服务器未运行，请先启动CCany服务器")
            return
        
        # 运行所有测试
        self.test_claude_to_openai_conversion()
        self.test_openai_to_claude_conversion()
        self.test_tool_calling_basic()
        self.test_tool_calling_complex()
        self.test_streaming_response()
        self.test_multimodal_input()
        self.test_error_handling()
        
        # 生成测试报告
        self.generate_report()
    
    def generate_report(self):
        """生成测试报告"""
        print("\n" + "=" * 60)
        print("📊 测试报告")
        print("=" * 60)
        
        total_tests = len(test_results)
        passed_tests = sum(1 for r in test_results if r["success"])
        failed_tests = total_tests - passed_tests
        
        print(f"\n总测试数: {total_tests}")
        print(f"✅ 通过: {passed_tests}")
        print(f"❌ 失败: {failed_tests}")
        print(f"成功率: {passed_tests/total_tests*100:.1f}%")
        
        if failed_tests > 0:
            print("\n失败的测试:")
            for result in test_results:
                if not result["success"]:
                    print(f"  - {result['test_name']}: {result['message']}")
        
        # 保存详细报告
        report_filename = f"test_report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        with open(report_filename, 'w', encoding='utf-8') as f:
            json.dump({
                "summary": {
                    "total": total_tests,
                    "passed": passed_tests,
                    "failed": failed_tests,
                    "success_rate": f"{passed_tests/total_tests*100:.1f}%"
                },
                "details": test_results
            }, f, indent=2, ensure_ascii=False)
        
        print(f"\n详细报告已保存到: {report_filename}")


def main():
    """主函数"""
    print("🧪 CCany 综合功能测试")
    print("测试工具调用和API格式转换功能")
    print("-" * 60)
    
    # 从命令行参数或环境变量获取API密钥
    api_key = API_KEY
    if len(sys.argv) > 1:
        api_key = sys.argv[1]
    
    print(f"服务器地址: {CCANY_BASE_URL}")
    print(f"API密钥: {'*' * (len(api_key) - 4) + api_key[-4:] if len(api_key) > 4 else '****'}")
    
    # 创建测试器并运行测试
    tester = CCanyTester(CCANY_BASE_URL, api_key)
    tester.run_all_tests()


if __name__ == "__main__":
    main()