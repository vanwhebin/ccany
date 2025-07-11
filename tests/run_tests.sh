#!/bin/bash

# Test runner script for ccany integration tests
# This script helps run tests with proper environment setup

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🧪 CCANY Integration Test Runner${NC}"
echo "=================================="

# Check if server is running
echo -e "\n${YELLOW}Checking if server is running...${NC}"
if curl -s http://localhost:8082/health > /dev/null; then
    echo -e "${GREEN}✅ Server is running${NC}"
else
    echo -e "${RED}❌ Server is not running on localhost:8082${NC}"
    echo -e "${YELLOW}Please start the server first with: go run ./cmd/server${NC}"
    exit 1
fi

# Load test environment variables if available
if [ -f "tests/.env.test" ]; then
    echo -e "\n${YELLOW}Loading test environment variables...${NC}"
    export $(grep -v '^#' tests/.env.test | xargs)
    echo -e "${GREEN}✅ Environment variables loaded${NC}"
else
    echo -e "\n${YELLOW}⚠️  No tests/.env.test file found${NC}"
    echo -e "${YELLOW}   Copy tests/.env.test.example to tests/.env.test and configure your test settings${NC}"
    echo -e "${YELLOW}   Some tests will be skipped without proper configuration${NC}"
fi

# Run tests
echo -e "\n${YELLOW}Running configuration tests...${NC}"
go test -v ./tests -run TestConfigurationAPI

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Configuration tests passed${NC}"
else
    echo -e "${RED}❌ Configuration tests failed${NC}"
    exit 1
fi

# Run integration tests if environment is configured
if [ -n "$TEST_API_KEY" ] && [ -n "$TEST_BASE_URL" ] && [ -n "$TEST_MODEL" ]; then
    echo -e "\n${YELLOW}Running OpenAI SDK integration tests...${NC}"
    go test -v ./tests -run TestOpenAISDKIntegration
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Integration tests passed${NC}"
    else
        echo -e "${RED}❌ Integration tests failed${NC}"
        exit 1
    fi
else
    echo -e "\n${YELLOW}⚠️  Skipping integration tests - missing TEST_API_KEY, TEST_BASE_URL, or TEST_MODEL${NC}"
fi

# Run all tests
echo -e "\n${YELLOW}Running all tests...${NC}"
go test -v ./tests

if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests completed successfully!${NC}"
else
    echo -e "\n${RED}❌ Some tests failed${NC}"
    exit 1
fi

echo -e "\n${GREEN}📊 Test Summary:${NC}"
echo -e "${GREEN}   ✅ Configuration API tests${NC}"
echo -e "${GREEN}   ✅ Authentication flow tests${NC}"
echo -e "${GREEN}   ✅ Data persistence tests${NC}"

if [ -n "$TEST_API_KEY" ]; then
    echo -e "${GREEN}   ✅ OpenAI SDK integration tests${NC}"
    echo -e "${GREEN}   ✅ Claude-to-OpenAI conversion tests${NC}"
    echo -e "${GREEN}   ✅ Streaming functionality tests${NC}"
fi

echo -e "\n${GREEN}🚀 Test suite completed successfully!${NC}"