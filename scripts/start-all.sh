#!/bin/bash

# AI-Nexus 一键启动脚本
# 启动所有服务（假设 Redis 和 ETCD 已运行）

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_DIR="$PROJECT_ROOT/logs"
PID_DIR="$PROJECT_ROOT/pids"

# 创建目录
mkdir -p "$LOG_DIR" "$PID_DIR"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查端口是否被占用
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0  # 端口被占用
    else
        return 1  # 端口空闲
    fi
}

# 启动服务
start_service() {
    local name=$1
    local dir=$2
    local port=$3
    
    log_info "启动 $name..."
    
    # 检查端口，如果被占用先杀掉
    if check_port $port; then
        log_warn "$name 端口 $port 已被占用，正在释放..."
        lsof -ti :$port | xargs kill -9 2>/dev/null
        sleep 1
    fi
    
    # 进入目录
    cd "$PROJECT_ROOT/$dir"
    
    # 编译（使用 CGO_ENABLED=0 避免 dyld 问题）
    log_info "编译 $name..."
    CGO_ENABLED=0 go build -o "$name" . 2>"$LOG_DIR/$name-build.log"
    if [ $? -ne 0 ]; then
        log_error "$name 编译失败，请检查日志: $LOG_DIR/$name-build.log"
        return
    fi
    
    # 后台启动，日志输出到文件
    ./"$name" > "$LOG_DIR/$name.log" 2>&1 &
    local pid=$!
    echo $pid > "$PID_DIR/$name.pid"
    
    # 等待服务启动
    sleep 2
    
    # 检查是否启动成功
    if check_port $port; then
        log_success "$name 已启动 (PID: $pid, Port: $port)"
    else
        log_error "$name 启动失败，请检查日志: $LOG_DIR/$name.log"
    fi
}

# 主流程
main() {
    echo ""
    echo "=========================================="
    echo "       AI-Nexus 服务启动脚本"
    echo "=========================================="
    echo ""
    
    # 检查基础设施
    log_info "检查基础设施..."
    
    if check_port 2379; then
        log_success "ETCD 运行中 (端口 2379)"
    else
        log_error "ETCD 未运行，请先启动 ETCD"
        exit 1
    fi
    
    if check_port 6379; then
        log_success "Redis 运行中 (端口 6379)"
    else
        log_warn "Redis 未运行，将使用内存存储"
    fi
    
    echo ""
    log_info "启动服务..."
    echo ""
    
    # 启动业务 Agents
    start_service "user-service" "cmd/user-service" 8081
    start_service "order-service" "cmd/order-service" 8082
    start_service "product-service" "cmd/product-service" 8083
    
    # 等待 Agents 注册
    sleep 2
    
    # 启动 Orchestrator
    start_service "orchestrator" "cmd/orchestrator" 9091
    
    # 等待 Orchestrator 注册
    sleep 2
    
    # 启动 WebUI
    start_service "web-ui" "cmd/web-ui" 8080
    
    echo ""
    echo "=========================================="
    log_success "所有服务已启动！"
    echo ""
    echo "  WebUI:           http://localhost:8080"
    echo "  Orchestrator:    http://localhost:9091"
    echo "  User Service:    http://localhost:8081"
    echo "  Order Service:   http://localhost:8082"
    echo "  Product Service: http://localhost:8083"
    echo ""
    echo "  日志目录: $LOG_DIR"
    echo "  停止服务: ./scripts/stop-all.sh"
    echo "=========================================="
    echo ""
}

main "$@"
