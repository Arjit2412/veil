#!/bin/bash

# Veil Observability Stack Startup Script
# This script starts Loki, Grafana, and Promtail for centralized logging

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if a port is available
port_available() {
    ! nc -z localhost "$1" 2>/dev/null
}

# Function to wait for service to be ready
wait_for_service() {
    local service_name=$1
    local url=$2
    local max_attempts=30
    local attempt=1

    print_status "Waiting for $service_name to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url" >/dev/null 2>&1; then
            print_success "$service_name is ready!"
            return 0
        fi
        
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_error "$service_name failed to start within expected time"
    return 1
}

# Main function
main() {
    print_status "Starting Veil Observability Stack..."
    
    # Check prerequisites
    if ! command_exists docker; then
        print_error "Docker is not installed. Please install Docker and try again."
        exit 1
    fi
    
    if ! command_exists docker-compose; then
        print_error "Docker Compose is not installed. Please install Docker Compose and try again."
        exit 1
    fi
    
    # Check if ports are available
    ports=(3000 3100 9080)
    for port in "${ports[@]}"; do
        if ! port_available "$port"; then
            print_warning "Port $port is already in use. This might cause conflicts."
        fi
    done
    
    # Navigate to the logging directory
    cd "$(dirname "$0")/.."
    
    # Create necessary directories
    print_status "Creating necessary directories..."
    mkdir -p config/grafana/provisioning/datasources
    mkdir -p config/grafana/provisioning/dashboards
    mkdir -p config/grafana/dashboards
    
    # Stop any existing containers
    print_status "Stopping any existing containers..."
    docker-compose down 2>/dev/null || true
    
    # Pull latest images
    print_status "Pulling latest Docker images..."
    docker-compose pull
    
    # Start the services
    print_status "Starting observability services..."
    docker-compose up -d
    
    # Wait for services to be ready
    wait_for_service "Loki" "http://localhost:3100/ready"
    wait_for_service "Grafana" "http://localhost:3000/api/health"
    
    # Display service information
    echo ""
    print_success "Veil Observability Stack is now running!"
    echo ""
    echo "Services:"
    echo "  📊 Grafana Dashboard: http://localhost:3000"
    echo "     Username: admin"
    echo "     Password: admin"
    echo ""
    echo "  📝 Loki (Logs):      http://localhost:3100"
    echo "  📤 Promtail:         http://localhost:9080"
    echo ""
    echo "Pre-configured dashboards:"
    echo "  • Veil API Gateway Dashboard"
    echo ""
    echo "To stop the stack, run:"
    echo "  docker-compose down"
    echo ""
    echo "To view logs, run:"
    echo "  docker-compose logs -f [service_name]"
    echo ""
    
    # Check if Veil is running and suggest next steps
    if curl -s "http://localhost:2020" >/dev/null 2>&1; then
        print_success "Veil API Gateway detected on port 2020. Logs will be automatically collected."
    else
        print_warning "Veil API Gateway not detected. Start Veil to see logs in Grafana."
        echo "  To start Veil: make run"
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Start the Veil Observability Stack (Loki + Grafana + Promtail)"
        echo ""
        echo "Options:"
        echo "  --help, -h    Show this help message"
        echo "  --stop        Stop the observability stack"
        echo "  --restart     Restart the observability stack"
        echo "  --status      Show status of running services"
        echo ""
        exit 0
        ;;
    --stop)
        print_status "Stopping Veil Observability Stack..."
        cd "$(dirname "$0")/.."
        docker-compose down
        print_success "Observability stack stopped."
        exit 0
        ;;
    --restart)
        print_status "Restarting Veil Observability Stack..."
        cd "$(dirname "$0")/.."
        docker-compose down
        sleep 2
        main
        exit 0
        ;;
    --status)
        print_status "Checking service status..."
        cd "$(dirname "$0")/.."
        docker-compose ps
        echo ""
        echo "Health checks:"
        echo -n "Loki: "
        if curl -s "http://localhost:3100/ready" >/dev/null 2>&1; then
            print_success "Ready"
        else
            print_error "Not ready"
        fi
        echo -n "Grafana: "
        if curl -s "http://localhost:3000/api/health" >/dev/null 2>&1; then
            print_success "Ready"
        else
            print_error "Not ready"
        fi
        exit 0
        ;;
    "")
        main
        ;;
    *)
        print_error "Unknown option: $1"
        echo "Use --help for usage information."
        exit 1
        ;;
esac