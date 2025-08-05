#!/usr/bin/env python3
"""
CCany Gemini API 格式转换测试
"""

import requests
import json
import time
from datetime import datetime

# 配置
CCANY_BASE_URL = "http://localhost:8082"
API_KEY = "test-api-key"

class GeminiTester:
    """Gemini API 格式转换测试类"""
    
    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url
        self.api_key = api_key
        self.session = requests.Session()
    
    def test_gemini_format(self):
        """测试 Gemini 格式的 API 调用"""
        print("\n🌟 测试 Gemini API 格式...")
        
        # Gemini 格式的请求
        gemini_request = {
            "contents": [
                {
                    "parts": [
                        {
                            "text": "请用一句话介绍人工智能。"
                        }
                    ]
                }
            ],
            "generationConfig": {
                "maxOutputTokens": 100,
                "temperature": 0.7,
                "topP": 0.9,
                "topK": 40
            }
        }
        
        # 测试不同的 Gemini 端点格式
        endpoints = [
            f"/v1beta/models/gemini-pro:generateContent?key={self.api_key}",
            f"/v1/models/gemini-pro:generateContent?key={self.api_key}",
            f"/gemini/v1beta/models/gemini-pro:generateContent?key={self.api_key}"
        ]
        
        for endpoint in endpoints:
            print(f"\n测试端点: {endpoint}")
            try:
                response = self.session.post(
                    f"{self.base_url}{endpoint}",
                    headers={"Content-Type": "application/json"},
                    json=gemini_request,
                    timeout=30
                )
                
                print(f"状态码: {response.status_code}")
                if response.status_code == 200:
                    result = response.json()
                    # 验证 Gemini 响应格式
                    if "candidates" in result:
                        print("✅ Gemini格式响应正确")
                        print(f"响应内容: {json.dumps(result, indent=2, ensure_ascii=False)[:500]}...")
                    else:
                        print("❌ 响应格式不符合Gemini规范")
                        print(f"响应: {json.dumps(result, indent=2, ensure_ascii=False)}")
                else:
                    print(f"❌ 请求失败: {response.text[:200]}")
                    
            except Exception as e:
                print(f"❌ 请求异常: {str(e)}")
    
    def test_gemini_streaming(self):
        """测试 Gemini 流式响应"""
        print("\n🌊 测试 Gemini 流式响应...")
        
        gemini_stream_request = {
            "contents": [
                {
                    "parts": [
                        {
                            "text": "从1数到5，每个数字一行。"
                        }
                    ]
                }
            ],
            "generationConfig": {
                "maxOutputTokens": 100,
                "temperature": 0.5
            }
        }
        
        endpoint = f"/v1beta/models/gemini-pro:streamGenerateContent?key={self.api_key}"
        
        try:
            response = self.session.post(
                f"{self.base_url}{endpoint}",
                headers={
                    "Content-Type": "application/json",
                    "Accept": "text/event-stream"
                },
                json=gemini_stream_request,
                stream=True,
                timeout=30
            )
            
            print(f"状态码: {response.status_code}")
            if response.status_code == 200:
                chunks_received = 0
                content_parts = []
                
                for line in response.iter_lines():
                    if line:
                        chunks_received += 1
                        try:
                            data = json.loads(line.decode('utf-8'))
                            # 提取内容
                            if "candidates" in data:
                                for candidate in data["candidates"]:
                                    if "content" in candidate:
                                        for part in candidate["content"].get("parts", []):
                                            if "text" in part:
                                                content_parts.append(part["text"])
                        except json.JSONDecodeError:
                            pass
                
                if chunks_received > 0:
                    print(f"✅ 成功接收 {chunks_received} 个数据块")
                    print(f"内容: {''.join(content_parts)}")
                else:
                    print("❌ 未接收到流式数据")
            else:
                print(f"❌ 请求失败: {response.text[:200]}")
                
        except Exception as e:
            print(f"❌ 请求异常: {str(e)}")
    
    def test_gemini_with_tools(self):
        """测试 Gemini 格式的工具调用"""
        print("\n🔧 测试 Gemini 工具调用格式...")
        
        # Gemini 格式的工具调用请求
        gemini_tool_request = {
            "contents": [
                {
                    "parts": [
                        {
                            "text": "今天北京的天气怎么样？"
                        }
                    ]
                }
            ],
            "tools": [
                {
                    "function_declarations": [
                        {
                            "name": "get_weather",
                            "description": "获取指定城市的天气信息",
                            "parameters": {
                                "type": "object",
                                "properties": {
                                    "location": {
                                        "type": "string",
                                        "description": "城市名称"
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
            ],
            "generationConfig": {
                "maxOutputTokens": 200,
                "temperature": 0.7
            }
        }
        
        endpoint = f"/v1beta/models/gemini-pro:generateContent?key={self.api_key}"
        
        try:
            response = self.session.post(
                f"{self.base_url}{endpoint}",
                headers={"Content-Type": "application/json"},
                json=gemini_tool_request,
                timeout=30
            )
            
            print(f"状态码: {response.status_code}")
            if response.status_code == 200:
                result = response.json()
                # 检查是否有函数调用
                has_function_call = False
                if "candidates" in result:
                    for candidate in result["candidates"]:
                        if "content" in candidate:
                            for part in candidate["content"].get("parts", []):
                                if "functionCall" in part:
                                    has_function_call = True
                                    print("✅ 包含函数调用:")
                                    print(f"  函数: {part['functionCall']['name']}")
                                    print(f"  参数: {json.dumps(part['functionCall']['args'], ensure_ascii=False)}")
                
                if not has_function_call:
                    print("⚠️ 响应中未包含函数调用")
                    print(f"响应: {json.dumps(result, indent=2, ensure_ascii=False)[:500]}...")
            else:
                print(f"❌ 请求失败: {response.text[:200]}")
                
        except Exception as e:
            print(f"❌ 请求异常: {str(e)}")
    
    def test_gemini_multimodal(self):
        """测试 Gemini 多模态输入"""
        print("\n🖼️ 测试 Gemini 多模态输入...")
        
        # 1x1 像素的蓝色 PNG 图片
        test_image_base64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="
        
        gemini_multimodal_request = {
            "contents": [
                {
                    "parts": [
                        {
                            "text": "这张图片主要是什么颜色？"
                        },
                        {
                            "inline_data": {
                                "mime_type": "image/png",
                                "data": test_image_base64
                            }
                        }
                    ]
                }
            ],
            "generationConfig": {
                "maxOutputTokens": 100,
                "temperature": 0.5
            }
        }
        
        endpoint = f"/v1beta/models/gemini-pro-vision:generateContent?key={self.api_key}"
        
        try:
            response = self.session.post(
                f"{self.base_url}{endpoint}",
                headers={"Content-Type": "application/json"},
                json=gemini_multimodal_request,
                timeout=30
            )
            
            print(f"状态码: {response.status_code}")
            if response.status_code == 200:
                result = response.json()
                print("✅ 多模态请求成功")
                print(f"响应: {json.dumps(result, indent=2, ensure_ascii=False)[:500]}...")
            else:
                print(f"❌ 请求失败: {response.text[:200]}")
                
        except Exception as e:
            print(f"❌ 请求异常: {str(e)}")
    
    def test_conversion_claude_to_gemini(self):
        """测试 Claude 格式转换为 Gemini 格式"""
        print("\n🔄 测试 Claude → Gemini 格式转换...")
        
        # 使用 Claude 格式的请求，但期望后端能转换为 Gemini
        claude_request = {
            "model": "claude-3-haiku-20240307",
            "max_tokens": 100,
            "messages": [
                {
                    "role": "user",
                    "content": "你好，请介绍一下你自己。"
                }
            ],
            "temperature": 0.7
        }
        
        # 如果配置了 Gemini 作为后端，这个请求应该被转换
        headers = {
            "x-api-key": self.api_key,
            "Content-Type": "application/json",
            "anthropic-version": "2023-06-01"
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/v1/messages",
                headers=headers,
                json=claude_request,
                timeout=30
            )
            
            print(f"状态码: {response.status_code}")
            if response.status_code == 200:
                result = response.json()
                print("✅ Claude格式请求成功（可能通过Gemini后端处理）")
                print(f"响应: {json.dumps(result, indent=2, ensure_ascii=False)[:500]}...")
            else:
                print(f"❌ 请求失败: {response.text[:200]}")
                
        except Exception as e:
            print(f"❌ 请求异常: {str(e)}")
    
    def run_all_tests(self):
        """运行所有 Gemini 相关测试"""
        print("\n🌟 开始 Gemini API 格式转换测试...")
        print("=" * 60)
        
        self.test_gemini_format()
        self.test_gemini_streaming()
        self.test_gemini_with_tools()
        self.test_gemini_multimodal()
        self.test_conversion_claude_to_gemini()
        
        print("\n" + "=" * 60)
        print("✅ Gemini 测试完成!")


def main():
    """主函数"""
    import sys
    
    api_key = API_KEY
    if len(sys.argv) > 1:
        api_key = sys.argv[1]
    
    print("🧪 CCany Gemini API 格式转换测试")
    print(f"服务器地址: {CCANY_BASE_URL}")
    print(f"API密钥: {'*' * (len(api_key) - 4) + api_key[-4:] if len(api_key) > 4 else '****'}")
    
    tester = GeminiTester(CCANY_BASE_URL, api_key)
    tester.run_all_tests()


if __name__ == "__main__":
    main()