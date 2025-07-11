#!/bin/bash

# Enhanced CCany Deployment Script with Claude Code Support
# This script handles deployment of the enhanced CCany application

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="$PROJECT_DIR/docker-compose.yml"
ENV_FILE="$PROJECT_DIR/.env"
ENV_EXAMPLE="$PROJECT_DIR/.env.example"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is installed and running
check_docker() {
    log_info "Checking Docker installation..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker is not running. Please start Docker first."
        exit 1
    fi
    
    log_success "Docker is installed and running"
}

# Check if Docker Compose is installed
check_docker_compose() {
    log_info "Checking Docker Compose installation..."
    
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    log_success "Docker Compose is available"
}

# Create environment file if it doesn't exist
setup_environment() {
    log_info "Setting up environment configuration..."
    
    if [ ! -f "$ENV_FILE" ]; then
        if [ -f "$ENV_EXAMPLE" ]; then
            log_info "Creating .env file from template..."
            cp "$ENV_EXAMPLE" "$ENV_FILE"
            log_warning "Please edit .env file and configure your API keys and settings"
        else
            log_error ".env.example file not found. Cannot create .env file."
            exit 1
        fi
    else
        log_info ".env file already exists"
    fi
}

# Build the application
build_application() {
    log_info "Building enhanced CCany application..."
    
    cd "$PROJECT_DIR"
    
    # Set build arguments
    export VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
    export BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    log_info "Building with VERSION=$VERSION, BUILD_TIME=$BUILD_TIME"
    
    if command -v docker-compose &> /dev/null; then
        docker-compose build --no-cache
    else
        docker compose build --no-cache
    fi
    
    log_success "Application built successfully"
}

# Start the services
start_services() {
    log_info "Starting enhanced CCany services..."
    
    cd "$PROJECT_DIR"
    
    if command -v docker-compose &> /dev/null; then
        docker-compose up -d
    else
        docker compose up -d
    fi
    
    log_success "Services started successfully"
}

# Stop the services
stop_services() {
    log_info "Stopping CCany services..."
    
    cd "$PROJECT_DIR"
    
    if command -v docker-compose &> /dev/null; then
        docker-compose down
    else
        docker compose down
    fi
    
    log_success "Services stopped successfully"
}

# Show service status
show_status() {
    log_info "Checking service status..."
    
    cd "$PROJECT_DIR"
    
    if command -v docker-compose &> /dev/null; then
        docker-compose ps
    else
        docker compose ps
    fi
}

# Show service logs
show_logs() {
    log_info "Showing service logs..."
    
    cd "$PROJECT_DIR"
    
    if command -v docker-compose &> /dev/null; then
        docker-compose logs -f
    else
        docker compose logs -f
    fi
}

# Test Claude Code compatibility
test_claude_code() {
    log_info "Testing Claude Code compatibility..."
    
    cd "$PROJECT_DIR"
    
    if command -v docker-compose &> /dev/null; then
        docker-compose --profile test up --build test-claude-code
    else
        docker compose --profile test up --build test-claude-code
    fi
    
    log_success "Claude Code compatibility test completed"
}

# Health check
health_check() {
    log_info "Performing health check..."
    
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -f -s http://localhost:8082/health >/dev/null; then
            log_success "Health check passed - CCany is running"
            return 0
        fi
        
        log_info "Waiting for CCany to start... ($((attempt+1))/$max_attempts)"
        sleep 2
        attempt=$((attempt+1))
    done
    
    log_error "Health check failed - CCany may not be running properly"
    return 1
}

# Start with monitoring
start_with_monitoring() {
    log_info "Starting CCany with monitoring stack..."
    
    cd "$PROJECT_DIR"
    
    if command -v docker-compose &> /dev/null; then
        docker-compose --profile monitoring up -d
    else
        docker compose --profile monitoring up -d
    fi
    
    log_success "CCany with monitoring started successfully"
    log_info "Grafana dashboard available at: http://localhost:3000"
    log_info "Prometheus metrics available at: http://localhost:9090"
}

# Start with nginx
start_with_nginx() {
    log_info "Starting CCany with Nginx reverse proxy..."
    
    cd "$PROJECT_DIR"
    
    if command -v docker-compose &> /dev/null; then
        docker-compose --profile nginx up -d
    else
        docker compose --profile nginx up -d
    fi
    
    log_success "CCany with Nginx started successfully"
    log_info "Application available at: http://localhost"
}

# Update the application
update_application() {
    log_info "Updating CCany application..."
    
    # Pull latest code (if in git repo)
    if [ -d "$PROJECT_DIR/.git" ]; then
        log_info "Pulling latest code..."
        cd "$PROJECT_DIR"
        git pull
    fi
    
    # Rebuild and restart
    build_application
    stop_services
    start_services
    
    log_success "Application updated successfully"
}

# Cleanup old images and containers
cleanup() {
    log_info "Cleaning up old Docker images and containers..."
    
    cd "$PROJECT_DIR"
    
    if command -v docker-compose &> /dev/null; then
        docker-compose down --remove-orphans
    else
        docker compose down --remove-orphans
    fi
    
    # Remove unused images
    docker image prune -f
    
    log_success "Cleanup completed"
}

# Show help
show_help() {
    echo "Enhanced CCany Deployment Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  build          Build the application"
    echo "  start          Start the services"
    echo "  stop           Stop the services"
    echo "  restart        Restart the services"
    echo "  status         Show service status"
    echo "  logs           Show service logs"
    echo "  health         Perform health check"
    echo "  test           Test Claude Code compatibility"
    echo "  monitoring     Start with monitoring stack"
    echo "  nginx          Start with Nginx reverse proxy"
    echo "  update         Update the application"
    echo "  cleanup        Clean up old images and containers"
    echo "  help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build               # Build the application"
    echo "  $0 start               # Start basic services"
    echo "  $0 monitoring          # Start with monitoring"
    echo "  $0 nginx               # Start with Nginx"
    echo "  $0 test                # Test Claude Code compatibility"
    echo ""
}

# Main execution
main() {
    case "${1:-start}" in
        "build")
            check_docker
            check_docker_compose
            setup_environment
            build_application
            ;;
        "start")
            check_docker
            check_docker_compose
            setup_environment
            start_services
            sleep 5
            health_check
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            stop_services
            sleep 2
            start_services
            sleep 5
            health_check
            ;;
        "status")
            show_status
            ;;
        "logs")
            show_logs
            ;;
        "health")
            health_check
            ;;
        "test")
            check_docker
            check_docker_compose
            test_claude_code
            ;;
        "monitoring")
            check_docker
            check_docker_compose
            setup_environment
            start_with_monitoring
            sleep 5
            health_check
            ;;
        "nginx")
            check_docker
            check_docker_compose
            setup_environment
            start_with_nginx
            sleep 5
            health_check
            ;;
        "update")
            check_docker
            check_docker_compose
            update_application
            sleep 5
            health_check
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            log_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Check if script is run directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi