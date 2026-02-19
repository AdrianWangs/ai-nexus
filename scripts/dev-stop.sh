#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="$ROOT/.run"

stop_pid() {
  local name="$1"
  local pidfile="$RUN_DIR/$name.pid"
  if [ -f "$pidfile" ]; then
    local pid
    pid="$(cat "$pidfile" || true)"
    if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      sleep 1
      kill -9 "$pid" 2>/dev/null || true
    fi
    rm -f "$pidfile"
  fi
}

stop_pid "web-ui"
stop_pid "orchestrator"
stop_pid "product-service"
stop_pid "order-service"
stop_pid "user-service"

if command -v lsof >/dev/null 2>&1; then
  lsof -ti:8080,8081,8082,8083,8084,8085,9090,9091 | xargs kill -9 2>/dev/null || true
fi

if command -v docker >/dev/null 2>&1; then
  docker compose -f "$ROOT/compose.yaml" stop etcd >/dev/null 2>&1 || true
fi

echo "AI Nexus stopped"
