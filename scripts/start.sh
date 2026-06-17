#!/usr/bin/env bash
# 一键启动 Chat AI · RAG Demo:Go 后端 + Vue 前端 + 自动灌库
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# 1) 检查 Ollama
if ! curl -fsS http://localhost:11434/api/tags > /dev/null 2>&1; then
  echo "[start] Ollama 未运行,先在另一个终端执行: ollama serve" >&2
  exit 1
fi

# 2) 检查必需模型
for m in qwen3:8b nomic-embed-text; do
  if ! ollama list 2>/dev/null | awk 'NR>1{print $1}' | grep -qx "$m"; then
    echo "[start] 缺少模型 $m,执行: ollama pull $m" >&2
    exit 1
  fi
done

# 3) 起后端
cd "$ROOT/backend"
echo "[start] go run . (port 8000)"
go run . &
BACKEND_PID=$!
trap 'kill $BACKEND_PID 2>/dev/null || true' EXIT

# 等就绪
for i in {1..30}; do
  if curl -fsS http://localhost:8000/api/health > /dev/null; then break; fi
  sleep 0.3
done

# 4) 灌库(每次启动都重新读 data/docs)
echo "[start] ingesting backend/data/docs ..."
curl -fsS -XPOST http://localhost:8000/api/ingest && echo

# 5) 起前端
cd "$ROOT/frontend"
[ -d node_modules ] || npm install
echo "[start] vite dev (port 5173)"
npm run dev
