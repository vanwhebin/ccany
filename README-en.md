# CCany - Go Version

[中文](README.md) | **English**

A Claude Code proxy server rewritten in Go, providing a complete frontend and backend interface that supports converting Claude API requests to OpenAI API calls.

![CCany](demo.png)

## Features

- **Complete Claude API Compatibility**: Supports full `/v1/messages` endpoint
- **Multi-Provider Support**: OpenAI, Azure OpenAI, local models (Ollama), and any OpenAI-compatible API
- **Intelligent Model Mapping**: Configure large and small models via environment variables
- **Function Calling**: Complete tool usage support and conversion
- **Streaming Response**: Real-time SSE streaming support
- **Image Support**: Base64 encoded image input
- **Web Management Interface**: Apple-style design management panel
- **Database Support**: Data persistence using ent and SQLite3
- **User Management**: Complete user authentication and authorization system
- **Request Logging**: Detailed API request logging and analysis
- **System Monitoring**: Real-time system performance monitoring and health checks
- **Cache Optimization**: Intelligent cache system for performance improvement
- **Error Handling**: Comprehensive error handling and logging

## Quick Start

### 1. Install Dependencies

```bash
# Ensure Go 1.21+ is installed
go version

# Download dependencies
go mod tidy
```

### 2. Optional Environment Variable Configuration

```bash
# Copy configuration file (optional)
cp .env.example .env

# Edit .env file to set data storage directory and security keys
# All API and service configurations are managed through the web interface
```

### 3. Start the Server

```bash
# Run directly
go run cmd/server/main.go

# Or build and run
go build -o ccany cmd/server/main.go
./ccany
```

### 4. Initial Setup

```bash
# After starting the server, initial setup is required for first access
# Visit http://localhost:8082/setup to create admin account and configure API keys
# Or use the deployment script for automated deployment:
chmod +x scripts/deploy.sh
./scripts/deploy.sh start
```

### 5. Configure API Keys

```bash
# After logging into the web interface, configure in the management panel:
# - OpenAI API key and base URL
# - Claude API key and base URL
# - Model configuration
# - Performance parameters
```

### 6. Use Claude Code

```bash
ANTHROPIC_BASE_URL=http://localhost:8082 ANTHROPIC_AUTH_TOKEN="some-api-key" claude
```

### 7. Access Web Interface

Open your browser and visit `http://localhost:8082` to view the management panel.

## Configuration

### Environment Variables (Optional)

**System Configuration:**

- `CLAUDE_PROXY_DATA_PATH` - Data storage directory (default: `./data`)
- `CLAUDE_PROXY_MASTER_KEY` - Master key for encrypting sensitive configurations (recommended for production)
- `JWT_SECRET` - JWT key for user authentication (recommended for production)

### Backend Configuration (Managed through Web Interface)

**API Configuration:**

- OpenAI API key and base URL
- Claude API key and base URL
- Azure API version (optional)

**Model Configuration:**

- Large model (default: `gpt-4o`)
- Small model (default: `gpt-4o-mini`)

**Server Settings:**

- Server host (default: `0.0.0.0`)
- Server port (default: `8082`)
- Log level (default: `info`)

**Performance Configuration:**

- Maximum token limit (default: `4096`)
- Minimum token limit (default: `100`)
- Request timeout seconds (default: `90`)
- Maximum retry attempts (default: `2`)
- Temperature parameter (default: `0.7`)
- Streaming response (default: `true`)

> 💡 **Note**: All API and service configurations are now managed through the web management interface, no longer requiring environment variables. Visit the `/setup` page for initial setup on first run.

### Model Mapping

The proxy maps Claude model requests to your configured models:

| Claude Request                     | Mapped To     | Environment Variable |
| ---------------------------------- | ------------- | ------------------- |
| Models containing "haiku"          | `SMALL_MODEL` | Default: `gpt-4o-mini` |
| Models containing "sonnet" or "opus" | `BIG_MODEL`   | Default: `gpt-4o` |

### Provider Configuration Examples

#### OpenAI
Configure through web interface:
- OpenAI API Key: `sk-your-openai-key`
- OpenAI Base URL: `https://api.openai.com/v1`
- Large Model: `gpt-4o`
- Small Model: `gpt-4o-mini`

#### Azure OpenAI
Configure through web interface:
- OpenAI API Key: `your-azure-key`
- OpenAI Base URL: `https://your-resource.openai.azure.com/openai/deployments/your-deployment`
- Azure API Version: `2024-02-01`
- Large Model: `gpt-4`
- Small Model: `gpt-35-turbo`

#### Local Model (Ollama)
Configure through web interface:
- OpenAI API Key: `dummy-key` (required but can be dummy)
- OpenAI Base URL: `http://localhost:11434/v1`
- Large Model: `llama3.1:70b`
- Small Model: `llama3.1:8b`

#### Claude Official API
Configure through web interface:
- Claude API Key: `sk-ant-your-claude-key`
- Claude Base URL: `https://api.anthropic.com`

## Web Management Interface

Access `http://localhost:8082` to use the web management interface, which includes:

- **Dashboard**: View service status and statistics
- **Request Logs**: View detailed records and analysis of all API requests
- **System Monitoring**: Real-time system performance monitoring and health checks
- **User Management**: User account management and permission control
- **Configuration Management**: View and modify system configuration parameters
- **API Testing**: Test connections and send test messages

The interface features Apple-style design with responsive layout and frosted glass effects. A complete authentication system ensures the security of the management interface.

## Project Structure

```
ccany/
├── cmd/
│   └── server/
│       └── main.go              # Main program entry
├── internal/
│   ├── app/                     # Application configuration management
│   ├── auth/                    # Authentication service
│   ├── cache/                   # Cache service
│   ├── claudecode/              # Claude Code compatibility services
│   ├── client/                  # OpenAI client
│   ├── config/                  # Configuration management
│   ├── converter/               # Request/response converter
│   ├── crypto/                  # Encryption service
│   ├── database/                # Database management
│   ├── handlers/                # HTTP handlers
│   ├── logging/                 # Request logging
│   ├── middleware/              # Middleware
│   ├── models/                  # Data models
│   └── monitoring/              # System monitoring
├── ent/
│   ├── schema/                  # Database schema definitions
│   └── ...                     # Generated ORM code
├── tests/
│   └── basic_test.go           # Basic test file
├── scripts/
│   └── deploy.sh                # Deployment script
├── web/
│   ├── index.html              # Main page
│   ├── setup.html              # Setup page
│   └── static/
│       ├── css/                # Style files
│       └── js/                 # JavaScript files
├── .env.example                # Configuration template
├── go.mod                      # Go module file
├── Makefile                    # Build script
└── README.md                   # This file
```

## API Endpoints

### Claude Compatible Endpoints

- `POST /v1/messages` - Create message
- `POST /v1/messages/count_tokens` - Count tokens

### Management Endpoints

- `GET /` - Web management interface
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /setup` - Initialization setup interface
- `POST /setup` - Submit initialization setup

### Authentication Endpoints

- `POST /admin/auth/login` - Admin login
- `POST /admin/auth/logout` - Admin logout
- `GET /admin/auth/profile` - Get user information

### Management API Endpoints

- `GET /admin/users` - Get user list
- `POST /admin/users` - Create user
- `PUT /admin/users/:id` - Update user
- `DELETE /admin/users/:id` - Delete user
- `GET /admin/config` - Get configuration information
- `PUT /admin/config` - Update configuration
- `GET /admin/request-logs` - Get request logs
- `GET /admin/request-logs/stats` - Get request statistics
- `DELETE /admin/request-logs` - Clear request logs

### Monitoring Endpoints

- `GET /admin/monitoring/health` - System health check
- `GET /admin/monitoring/metrics` - System metrics
- `GET /admin/monitoring/system` - System information

## Development

### Build

```bash
# Build executable
go build -o ccany cmd/server/main.go

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o ccany-linux cmd/server/main.go
GOOS=windows GOARCH=amd64 go build -o ccany.exe cmd/server/main.go
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./internal/converter

# Run integration tests (requires server to be running)
go test -v ./tests/

# Run benchmark tests
go test -bench=. ./tests/

# Generate test coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Code Formatting

```bash
# Format code
go fmt ./...

# Check code
go vet ./...
```

## Performance

- **Concurrent Processing**: High concurrency using Goroutines
- **Connection Pool**: Efficient HTTP connection management
- **Streaming Support**: Real-time response streaming
- **Intelligent Caching**: Multi-strategy cache system (LRU, LFU, TTL)
- **Request Logging**: Asynchronous logging without performance impact
- **System Monitoring**: Low-overhead real-time performance monitoring
- **Database Optimization**: Connection pooling and query optimization
- **Memory Management**: Graceful memory usage and garbage collection
- **Configurable Timeouts**: Configurable timeouts and retries
- **Intelligent Error Handling**: Detailed logging

## Integration with Claude Code

This proxy is designed to work seamlessly with Claude Code CLI. **The enhanced version includes complete Claude Code compatibility support**:

```bash
# Start the enhanced proxy using deployment script
./scripts/deploy.sh start

# Use Claude Code with the proxy
ANTHROPIC_BASE_URL=http://localhost:8082 claude

# Or set permanently
export ANTHROPIC_BASE_URL=http://localhost:8082
claude
```

### Enhanced Claude Code Features

- ✅ **Complete SSE Event Sequence**: Support for `message_start`, `content_block_start`, `ping`, `content_block_delta`, `content_block_stop`, `message_delta`, `message_stop` events
- ✅ **Request Cancellation Support**: Client disconnect detection and graceful request cancellation
- ✅ **Claude Configuration Automation**: Automatic creation of `~/.claude.json` configuration file
- ✅ **Thinking Mode**: Support for `thinking` field and intelligent model routing
- ✅ **Enhanced Tool Calls**: Tool call streaming with incremental JSON parsing
- ✅ **Cache Tokens**: Support for `cache_read_input_tokens` usage reporting
- ✅ **Smart Routing**: Intelligent model selection based on complexity and token count

### Deployment Options

```bash
# Basic deployment
./scripts/deploy.sh start

# Deployment with monitoring
./scripts/deploy.sh monitoring

# Deployment with Nginx
./scripts/deploy.sh nginx

# Test Claude Code compatibility
./scripts/deploy.sh test

# Show help
./scripts/deploy.sh help
```

## Docker Deployment

### Using Docker Compose

```bash
# Copy environment configuration
cp .env.example .env
# Edit .env file to configure API keys

# Basic deployment
docker-compose up -d

# Deployment with monitoring
docker-compose --profile monitoring up -d

# Deployment with Nginx
docker-compose --profile nginx up -d

# Test Claude Code compatibility
docker-compose --profile test up --build test-claude-code
```

### Using Deployment Script

```bash
# Automated deployment (recommended)
./scripts/deploy.sh start

# Deployment with monitoring stack
./scripts/deploy.sh monitoring

# Check service status
./scripts/deploy.sh status

# View logs
./scripts/deploy.sh logs
```

For detailed deployment guide, please refer to: [Deployment Documentation](docs/DEPLOYMENT_GUIDE.md)

## License

MIT License

## Contributing

Issues and Pull Requests are welcome!

## Changelog

### v1.3.0 (Enhanced - Claude Code Compatibility)
- ✅ Complete Claude Code compatibility support
- ✅ Enhanced SSE event sequence (message_start, content_block_start, ping, content_block_delta, content_block_stop, message_delta, message_stop)
- ✅ Request cancellation and client disconnect detection
- ✅ Claude configuration auto-initialization (~/.claude.json)
- ✅ Thinking mode support and intelligent model routing
- ✅ Enhanced tool call streaming
- ✅ Cache token usage reporting
- ✅ Enhanced Docker and Docker Compose configuration
- ✅ GitHub Actions CI/CD pipeline
- ✅ Complete deployment scripts and monitoring support
- ✅ Enhanced Nginx configuration and performance optimization

### v1.2.0
- Complete backend management system
- User authentication and authorization
- Request logging and analysis
- System monitoring and health checks
- Intelligent cache system
- Complete test suite
- Enhanced documentation and deployment guide

### v1.1.0
- Database integration and ORM support
- Configuration management system
- Secure configuration storage
- Database migration support

### v1.0.0
- Initial Go version release
- Complete Claude API compatibility
- Web management interface
- Database support
- Apple-style design UI