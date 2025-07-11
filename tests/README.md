# CCANY Test Suite

This directory contains comprehensive integration tests for the CCANY (Claude-to-OpenAI) proxy server.

## Test Structure

- `configuration_api_test.go` - Tests for configuration management API
- `openai_sdk_integration_test.go` - Tests for OpenAI SDK integration and API compatibility
- `run_tests.sh` - Automated test runner script
- `.env.test.example` - Template for test environment variables

## Running Tests

### Prerequisites

1. **Server Running**: The CCANY server must be running on `localhost:8082`
   ```bash
   # Start the server in one terminal
   go run ./cmd/server
   ```

2. **Test Configuration** (optional): Copy and configure test environment variables
   ```bash
   # Copy the example file
   cp tests/.env.test.example tests/.env.test
   
   # Edit with your test settings
   nano tests/.env.test
   ```

### Quick Start

Run all tests with the test runner script:

```bash
# Make script executable (first time only)
chmod +x tests/run_tests.sh

# Run all tests
./tests/run_tests.sh
```

### Manual Test Execution

#### Configuration Tests (No External Dependencies)

These tests verify the configuration API functionality and don't require external API keys:

```bash
go test -v ./tests -run TestConfigurationAPI
```

#### Integration Tests (Requires Test API Configuration)

These tests require valid API credentials in your `.env.test` file:

```bash
# Set environment variables
export TEST_API_KEY="your-test-api-key"
export TEST_BASE_URL="https://api.example.com/v1"
export TEST_MODEL="your-test-model"

# Run integration tests
go test -v ./tests -run TestOpenAISDKIntegration
```

#### All Tests

```bash
go test -v ./tests
```

## Test Configuration

### Environment Variables

Create `tests/.env.test` with your test configuration:

```bash
# Example for SiliconFlow API
TEST_API_KEY=sk-your-test-api-key-here
TEST_BASE_URL=https://api.siliconflow.cn/v1
TEST_MODEL=deepseek-ai/DeepSeek-V3

# Example for OpenAI API
# TEST_API_KEY=sk-your-openai-test-key
# TEST_BASE_URL=https://api.openai.com/v1
# TEST_MODEL=gpt-3.5-turbo
```

### Security Notes

- **Never commit real API keys** to version control
- Use test API keys with limited quotas when possible
- The `.env.test` file is ignored by git
- Tests use placeholder values when real keys aren't available

## Test Coverage

### Configuration API Tests

- ✅ Admin user setup and authentication
- ✅ Configuration retrieval (GET /admin/config)
- ✅ Configuration updates (PUT /admin/config)
- ✅ Sensitive field masking
- ✅ Partial configuration updates
- ✅ Configuration persistence
- ✅ Input validation

### OpenAI SDK Integration Tests

- ✅ OpenAI compatible API calls (POST /v1/chat/completions)
- ✅ Streaming responses
- ✅ Claude-to-OpenAI format conversion (POST /v1/messages)
- ✅ Model routing (Claude model names → configured models)
- ✅ Authentication with different header formats
- ✅ Response format validation

### What Gets Tested

1. **API Compatibility**: Ensures the server correctly implements OpenAI-compatible endpoints
2. **Format Conversion**: Verifies Claude API requests are properly converted to OpenAI format
3. **Authentication**: Tests both OpenAI-style (`Authorization: Bearer`) and Claude-style (`x-api-key`) auth
4. **Configuration Management**: Validates configuration persistence and security
5. **Model Routing**: Confirms Claude model names are correctly mapped to configured models
6. **Error Handling**: Tests graceful handling of invalid requests and configurations

## Continuous Integration

These tests are designed to be run in CI/CD environments:

```yaml
# Example GitHub Actions workflow
- name: Run Integration Tests
  env:
    TEST_API_KEY: ${{ secrets.TEST_API_KEY }}
    TEST_BASE_URL: ${{ secrets.TEST_BASE_URL }}
    TEST_MODEL: ${{ secrets.TEST_MODEL }}
  run: |
    go run ./cmd/server &
    sleep 5
    ./tests/run_tests.sh
```

## Troubleshooting

### Common Issues

1. **Server Not Running**: Ensure the server is started on port 8082
2. **Test Failures**: Check that environment variables are set correctly
3. **API Key Issues**: Verify test API keys are valid and have sufficient quota
4. **Port Conflicts**: Make sure port 8082 is available

### Test Debugging

Run tests with verbose output to see detailed information:

```bash
go test -v ./tests -run TestName
```

Add custom logging in tests for debugging:

```go
t.Logf("Debug info: %v", someValue)
```

## Adding New Tests

When adding new test cases:

1. **Follow the naming pattern**: `Test*` functions for test cases
2. **Use subtests**: Organize related tests with `t.Run()`
3. **Clean up**: Ensure tests don't leave system in inconsistent state
4. **Document**: Add clear descriptions of what each test verifies
5. **Security**: Never hardcode sensitive information

Example test structure:

```go
func TestNewFeature(t *testing.T) {
    t.Run("Setup", func(t *testing.T) {
        // Setup code
    })
    
    t.Run("Test Case 1", func(t *testing.T) {
        // Test implementation
    })
    
    t.Run("Cleanup", func(t *testing.T) {
        // Cleanup code
    })
}
```