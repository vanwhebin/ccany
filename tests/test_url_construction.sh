#!/bin/bash

echo "🧪 Testing Universal URL Construction Logic"
echo "=========================================="

BASE_URL="http://localhost:8082"

test_urls=(
    "https://ark.cn-beijing.volces.com/api/v3/"
    "https://api.x.ai"
    "https://openrouter.ai/api/v1/"
    "https://kilocode.ai/api/openrouter"
    "https://api.openai.com/v1"
    "https://api.example.com"
)

expected_finals=(
    "https://ark.cn-beijing.volces.com/api/v3/chat/completions"
    "https://api.x.ai/v1/chat/completions"
    "https://openrouter.ai/api/v1/chat/completions"
    "https://kilocode.ai/api/openrouter/chat/completions"
    "https://api.openai.com/v1/chat/completions"
    "https://api.example.com/v1/chat/completions"
)

echo ""
for i in "${!test_urls[@]}"; do
    url="${test_urls[$i]}"
    expected="${expected_finals[$i]}"
    
    echo "Testing: $url"
    
    # Call the API endpoint to get the final URL
    result=$(curl -s "$BASE_URL/api/final-endpoint-url?base_url=$(printf '%s' "$url" | jq -sRr @uri)")
    
    if [ $? -eq 0 ]; then
        final_url=$(echo "$result" | jq -r '.final_url' 2>/dev/null)
        
        if [ "$final_url" = "$expected" ]; then
            echo "✅ PASS: $final_url"
        else
            echo "❌ FAIL: Got '$final_url', expected '$expected'"
        fi
    else
        echo "❌ ERROR: API call failed"
    fi
    
    echo ""
done

echo "🎯 Test completed!"