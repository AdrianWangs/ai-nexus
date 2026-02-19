#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="$ROOT/.run"
LOG_DIR="$ROOT/.logs"

mkdir -p "$RUN_DIR" "$LOG_DIR"

SHOW_LOGS=${SHOW_LOGS:-false}
if [[ "${1:-}" == "-l" || "${1:-}" == "--logs" ]]; then
  SHOW_LOGS=true
fi

cleanup() {
  if [[ "$SHOW_LOGS" == "true" ]]; then
    echo ""
    echo "Stopping log tail..."
  fi
}
trap cleanup EXIT

if command -v docker >/dev/null 2>&1; then
  echo "Starting ETCD..."
  docker compose -f "$ROOT/compose.yaml" up -d etcd >/dev/null 2>&1 || true
  sleep 1
fi

start_service() {
  local name="$1"
  local dir="$2"
  local port="$3"
  local pidfile="$RUN_DIR/$name.pid"
  
  if lsof -ti:"$port" >/dev/null 2>&1; then
    echo "✓ $name already running on port $port"
    return
  fi
  if [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile")" 2>/dev/null; then
    echo "✓ $name already running (PID: $(cat "$pidfile"))"
    return
  fi
  
  > "$LOG_DIR/$name.log"
  
  (cd "$dir" && CGO_ENABLED=0 go run . >> "$LOG_DIR/$name.log" 2>&1 & echo $! > "$pidfile")
  echo "→ Starting $name on port $port..."
}

echo ""
echo "=== Starting AI Nexus Services ==="
echo ""

start_service "user-service" "$ROOT/examples/user-service" 8080
start_service "order-service" "$ROOT/examples/order-service" 8082
start_service "product-service" "$ROOT/examples/product-service" 8084
start_service "orchestrator" "$ROOT/examples/orchestrator" 9091
start_service "web-ui" "$ROOT/examples/web-ui" 9090

sleep 2

echo ""
echo "=== Service Status ==="
echo ""

check_service() {
  local name="$1"
  local port="$2"
  if lsof -ti:"$port" >/dev/null 2>&1; then
    echo "✓ $name: running on port $port"
  else
    echo "✗ $name: NOT running"
  fi
}

check_service "user-service" 8080
check_service "order-service" 8082
check_service "product-service" 8084
check_service "orchestrator" 9091
check_service "web-ui" 9090

echo ""
echo "=== Access URLs ==="
echo ""
echo "  Web UI:     http://localhost:9090"
echo "  Orchestrator A2A: http://localhost:9091"
echo "  User API:   http://localhost:8080/api/users"
echo "  Order API:  http://localhost:8082/api/orders"
echo "  Product API: http://localhost:8084/api/products"
echo ""
echo "=== Logs ==="
echo ""
echo "  View all:   tail -f $LOG_DIR/*.log"
echo "  Web UI:     tail -f $LOG_DIR/web-ui.log"
echo "  Orchestrator: tail -f $LOG_DIR/orchestrator.log"
echo ""

if [[ "$SHOW_LOGS" == "true" ]]; then
  echo "=== Live Logs (Ctrl+C to exit) ==="
  echo ""
  tail -f "$LOG_DIR"/*.log
fi
