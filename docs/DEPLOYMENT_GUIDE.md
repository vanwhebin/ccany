# Enhanced CCany Deployment Guide

This guide covers the deployment of the enhanced CCany application with full Claude Code compatibility.

## Quick Start

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd ccany
   ```

2. **Set up environment**:
   ```bash
   cp .env.example .env
   # Edit .env file with your API keys and configuration
   ```

3. **Deploy using the deployment script**:
   ```bash
   ./scripts/deploy.sh start
   ```

## Deployment Options

### Basic Deployment
```bash
./scripts/deploy.sh start
```

### With Monitoring Stack
```bash
./scripts/deploy.sh monitoring
```

### With Nginx Reverse Proxy
```bash
./scripts/deploy.sh nginx
```

### Test Claude Code Compatibility
```bash
./scripts/deploy.sh test
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI API key | Required |
| `CLAUDE_API_KEY` | Claude API key | Optional |
| `BIG_MODEL` | Large model for complex tasks | `gpt-4o` |
| `SMALL_MODEL` | Small model for simple tasks | `gpt-4o-mini` |
| `REASONING_MODEL` | Model for thinking mode | `gpt-4o` |
| `LONG_CONTEXT_MODEL` | Model for long context | `gpt-4o` |
| `MAX_TOKENS_LIMIT` | Maximum tokens per request | `8192` |
| `REQUEST_TIMEOUT` | Request timeout in seconds | `120` |
| `CLAUDE_CODE_COMPATIBLE` | Enable Claude Code compatibility | `true` |
| `CLAUDE_PROXY_DATA_PATH` | Data storage directory | `./data` |
| `CLAUDE_PROXY_MASTER_KEY` | Master key for encrypting sensitive configs | Required for production |
| `JWT_SECRET` | JWT key for user authentication | Required for production |

### Claude Code Features

The enhanced version includes:

- ✅ **Complete SSE Event Sequence**: Proper `message_start`, `content_block_start`, `ping`, `content_block_delta`, `content_block_stop`, `message_delta`, and `message_stop` events
- ✅ **Request Cancellation**: Client disconnect detection and graceful request cancellation
- ✅ **Claude Configuration**: Automatic `~/.claude.json` configuration file creation
- ✅ **Thinking Mode**: Support for `thinking` field with intelligent model routing
- ✅ **Enhanced Tool Calls**: Proper tool call streaming with incremental JSON parsing
- ✅ **Cache Tokens**: Support for `cache_read_input_tokens` in usage reporting
- ✅ **Smart Routing**: Intelligent model selection based on complexity and token count

## Docker Compose Services

### Core Services

- **ccany**: Main application with Claude Code enhancements
- **redis**: Caching and session management
- **nginx**: Reverse proxy (optional)

### Monitoring Services (optional)

- **prometheus**: Metrics collection
- **grafana**: Visualization dashboard

### Test Services

- **test-claude-code**: Claude Code compatibility tests

## Health Checks

The application includes comprehensive health checks:

```bash
# Check application health
curl http://localhost:8082/health

# Check detailed health status
curl http://localhost:8082/health/detailed

# Check system metrics
curl http://localhost:8082/health/metrics
```

## Monitoring

When deployed with monitoring:

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090

## Troubleshooting

### Common Issues

1. **Port already in use**:
   ```bash
   docker-compose down
   ./scripts/deploy.sh cleanup
   ```

2. **Permission denied**:
   ```bash
   chmod +x scripts/deploy.sh
   ```

3. **Database issues**:
   ```bash
   docker volume rm ccany_ccany_data
   ./scripts/deploy.sh restart
   ```

### Logs

```bash
# Show all service logs
./scripts/deploy.sh logs

# Show specific service logs
docker-compose logs ccany

# Follow logs in real-time
docker-compose logs -f ccany
```

## Development

### Building Locally

```bash
# Build the application
./scripts/deploy.sh build

# Run tests
./scripts/deploy.sh test

# Development with hot reload
docker-compose up --build
```

### Adding New Features

1. Make code changes
2. Test locally: `./scripts/deploy.sh test`
3. Build: `./scripts/deploy.sh build`
4. Deploy: `./scripts/deploy.sh restart`

## Production Deployment

### Prerequisites

- Docker and Docker Compose installed
- Sufficient system resources (2GB RAM minimum)
- Network access to OpenAI/Claude APIs

### Security Considerations

1. **Use strong passwords**:
   ```bash
   export JWT_SECRET=$(openssl rand -base64 32)
   export CLAUDE_PROXY_MASTER_KEY=$(openssl rand -base64 32)
   ```

2. **Enable HTTPS**:
   - Configure SSL certificates in `ssl/` directory
   - Use the nginx profile for reverse proxy

3. **Firewall configuration**:
   ```bash
   # Allow only necessary ports
   ufw allow 22/tcp
   ufw allow 80/tcp
   ufw allow 443/tcp
   ufw enable
   ```

### Scaling

For high-traffic deployments:

1. **Use multiple replicas**:
   ```bash
   docker-compose up --scale ccany=3
   ```

2. **External database**:
   - Configure PostgreSQL instead of SQLite
   - Update `DATABASE_URL` in `.env`

3. **Load balancing**:
   - Use nginx upstream configuration
   - Configure health checks

## Backup and Recovery

### Backup

```bash
# Backup data volumes
docker run --rm -v ccany_ccany_data:/data -v $(pwd):/backup alpine tar czf /backup/ccany-backup.tar.gz /data

# Backup configuration
cp .env .env.backup
```

### Recovery

```bash
# Restore data volumes
docker run --rm -v ccany_ccany_data:/data -v $(pwd):/backup alpine tar xzf /backup/ccany-backup.tar.gz

# Restore configuration
cp .env.backup .env
```

## API Documentation

### Enhanced Endpoints

- `POST /v1/messages` - Claude Code compatible messages
- `GET /v1/models/capabilities` - Model capabilities
- `POST /v1/messages/count_tokens` - Enhanced token counting
- `POST /v1/chat/completions` - OpenAI compatible endpoint

### Claude Code Specific Features

- **Thinking Mode**: Include `"thinking": true` in requests
- **Model Commands**: Use `/model provider,model` in messages
- **Enhanced Streaming**: Complete SSE event sequence
- **Tool Calls**: Proper incremental JSON parsing

## Support

For issues and questions:

1. Check the [troubleshooting guide](docs/CLAUDE_CODE_BUGFIX.md)
2. Review application logs
3. Check health endpoints
4. Verify configuration