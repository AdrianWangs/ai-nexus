#!/bin/bash

# AI-Nexus 一键停止脚本
# 停止所有服务

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PID_DIR="$PROJECT_ROOT/pids"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 停止服务
stop_service() {
    local name=$1
    local pid_file="$PID_DIR/$name.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            kill $pid 2>/dev/null
            log_success "已停止 $name (PID: $pid)"
        else
            log_info "$name 已经停止"
        fi
        rm -f "$pid_file"
    else
        log_info "$name PID 文件不存在"
    fi
}

# 主流程
main() {
    echo ""
    echo "=========================================="
    echo "       AI-Nexus 服务停止脚本"
    echo "=========================================="
    echo ""
    
    # 按启动的逆序停止
    stop_service "web-ui"
    stop_service "orchestrator"
    stop_service "product-service"
    stop_service "order-service"
    stop_service "user-service"
    
    echo ""
    log_success "所有服务已停止"
    echo "=========================================="
    echo ""
}

main "$@"
