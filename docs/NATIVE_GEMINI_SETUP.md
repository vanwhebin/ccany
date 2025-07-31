# Native Gemini API Integration

This document explains how to configure ccany to use Google's native Gemini API instead of the OpenAI-compatible endpoint.

## Overview

ccany now supports two ways to connect to Google's Gemini models:

1. **OpenAI-Compatible Endpoint** (Legacy) - Uses Gemini's OpenAI-compatible API
2. **Native Gemini API** (Recommended) - Uses Google's native Gemini API directly

## Why Use Native Gemini API?

The native Gemini API provides:
- ✅ **Better Tool/Function Calling** - Proper support for Claude-style tools
- ✅ **More Reliable** - Direct integration without compatibility layer issues
- ✅ **Full Feature Support** - Access to all Gemini-specific capabilities
- ✅ **Better Performance** - No conversion overhead

## Configuration

### Automatic Detection

ccany automatically detects when you're using Gemini and routes requests through the native API when:
- Base URL contains `generativelanguage.googleapis.com`
- Model names contain `gemini`

### Recommended Configuration

```bash
# Set these in your ccany configuration:
OPENAI_API_KEY=your_gemini_api_key_here
OPENAI_BASE_URL=https://generativelanguage.googleapis.com/v1beta/openai
BIG_MODEL=gemini-2.5-flash
SMALL_MODEL=gemini-1.5-flash
```

**Note:** Even though the base URL contains `/openai`, ccany will automatically convert this to the native endpoint `https://generativelanguage.googleapis.com/v1beta/models` for native API calls.

### Supported Models

- `gemini-1.5-flash` - Fast, efficient model
- `gemini-1.5-flash-latest` - Latest version of flash model
- `gemini-1.5-pro` - More capable model
- `gemini-1.5-pro-latest` - Latest version of pro model
- `gemini-2.5-flash` - Latest generation flash model

## How It Works

1. **Request Detection** - ccany detects Gemini configuration
2. **Native Conversion** - Converts Claude requests to native Gemini format
3. **Direct API Call** - Sends request to Google's native API
4. **Response Conversion** - Converts Gemini response back to Claude format

## Tool/Function Calling

The native integration properly handles:
- ✅ Tool definitions with proper schema conversion
- ✅ Tool choice modes (`auto`, `required`, `none`)
- ✅ Function calls and responses
- ✅ Mixed content (text + tool calls)

## Troubleshooting

### Common Issues

1. **404 Errors** - Check that your API key is valid and has Gemini access
2. **400 Schema Errors** - Ensure tool definitions follow Claude format
3. **Authentication Errors** - Verify your Gemini API key is correct

### Debug Logging

ccany provides detailed logging for native Gemini requests:
```
🔧 Detected Gemini backend - using Gemini converter
🔧 Configured native Gemini API endpoint
🔧 Converted Claude request to native Gemini format
🔧 Sending request to native Gemini API
🔧 Successfully processed request with native Gemini API
```

## Migration from OpenAI-Compatible

If you're currently using the OpenAI-compatible endpoint:

1. **No Configuration Changes Needed** - ccany automatically uses native API
2. **Better Tool Support** - Tool calling will work more reliably
3. **Same API Interface** - Your client code doesn't need to change

## Testing

Test your configuration with a tool calling request:

```bash
curl -X POST http://localhost:8082/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "max_tokens": 150,
    "messages": [{"role": "user", "content": "What is the weather like?"}],
    "tools": [{
      "name": "get_weather",
      "description": "Get weather for a location",
      "input_schema": {
        "type": "object",
        "properties": {"location": {"type": "string"}},
        "required": ["location"]
      }
    }],
    "tool_choice": "auto"
  }'
```

## Benefits Achieved

With native Gemini integration, you get:
- 🚀 **Reliable tool calling** without hanging or errors
- 🎯 **Proper tool choice handling** (`auto`, `required`, `none`)
- 📊 **Better error handling** with detailed Gemini-specific messages
- ⚡ **Improved performance** with direct API communication
- 🔧 **Full compatibility** with Claude API format
