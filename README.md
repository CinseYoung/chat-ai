# Chat AI · RAG Demo

一个**用于组内技术分享**的最小 RAG 演示项目:同一个问题,一键切换"裸 LLM" vs "接入 RAG",直观对比效果差异。

## 它演示了什么

- **Embedding** —— 文本如何变向量
- **Chunking** —— 文档如何切块
- **向量检索** —— 余弦相似度 + Top-K
- **RAG Prompt 构造** —— 检索结果如何注入上下文
- **可视化命中** —— UI 上展开看到模型实际看了哪些片段
- **幻觉对照** —— 同一问题关 RAG 模型胡编,开 RAG 给出有引用的答案

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | **Go** + Gin |
| LLM | Ollama / `qwen3:8b` |
| Embedding | Ollama / `nomic-embed-text` |
| 向量库 | [chromem-go](https://github.com/philippgille/chromem-go)(纯 Go 内嵌,持久化在 `backend/data/index/`) |
| 流式 | SSE |
| 前端 | Vue 3 + TypeScript + Tailwind |

后端代码:**3 个 Go 文件、不到 500 行**(`main.go` / `ollama.go` / `rag.go`),没有 LangChain、没有 Python、没有外部数据库。

## 项目结构

```
chat-ai/
├── backend/
│   ├── main.go            # 入口 + Gin 路由 + 三个 handler
│   ├── ollama.go          # /api/embeddings + /api/chat 流式封装
│   ├── rag.go             # chunk + chromem-go store + prompt + ingest
│   ├── data/docs/         # 知识库源文件(.md / .txt)
│   ├── data/index/        # chromem-go 自动生成的持久化向量
│   └── go.mod
├── frontend/              # Vue3 前端,加了 RAG 开关 + 命中片段展示
└── scripts/start.sh       # 一键启动后端 + 前端
```

## 准备工作

```bash
# 1. 安装并启动 Ollama(https://ollama.com)
ollama serve                      # 另开终端

# 2. 拉模型
ollama pull qwen3:8b              # ~4.9GB,对话模型
ollama pull nomic-embed-text      # ~270MB,嵌入模型(RAG 必需)
```

## 一键启动

```bash
bash scripts/start.sh
```

脚本会:检查 Ollama → 起 Go 后端(:8000) → 自动 ingest 一次知识库 → 起 Vue 前端(:5173)。

打开 http://localhost:5173,在底部输入框上方点击 **「RAG: 关 / 开」** 切换效果。

## 手动启动

```bash
# 后端
cd backend
go run .                          # 默认监听 :8000
curl -XPOST http://localhost:8000/api/ingest    # 首次启动后灌入知识库

# 前端(另一个终端)
cd frontend
npm install                       # 首次
npm run dev
```

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET  | `/api/health` | Ollama 连接状态 + 当前 chunk 数 |
| POST | `/api/ingest` | 重新读取 `backend/data/docs/*.md/.txt` 入库 |
| GET  | `/api/chat/stream?message=...&use_rag=true` | SSE 流式聊天 |

SSE 事件:
- `context` —— RAG 模式下先推命中片段(用于前端可视化)
- `content` —— 增量 token
- `done` —— 结束
- `error` —— 错误

## 演示问题(已内置在欢迎页快捷卡片)

| 问题 | 关 RAG 表现 | 开 RAG 期望 |
|------|------|------|
| 我们组的代号是什么?周会几点? | "我不确定..." | Phoenix-7 + 周三下午 4 点 [1] |
| 单 PR 行数上限? | 通用答 500/1000 | 800 行 [1] |
| 简要解释 RAG 流水线? | 通用解释 | 引用提纲文档 [1] |
| CEO 办公室在哪一层? | 编一个 | "资料中未提及" |

## 自带的 RAG 知识库

`backend/data/docs/` 下两份示例文档:
- `team-handbook.md` —— 虚构团队信息(代号、周会、CR 规则、值班、发布窗口...)
- `rag-talk-outline.md` —— 分享提纲本身(可同时充当 RAG 检索内容)

要换成自己的内容,直接往这个目录扔 `.md`/`.txt`,然后点前端的「重灌知识库」按钮(或 `curl -XPOST /api/ingest`)。

## 配置

可通过环境变量覆盖:`OLLAMA_BASE_URL` / `CHAT_MODEL` / `EMBED_MODEL` / `PORT` / `DATA_DIR` / `INDEX_DIR`。

```bash
EMBED_MODEL=bge-m3 CHAT_MODEL=qwen3:14b go run ./backend
```

## 分享时可拓展讲的话题

- **Hybrid Search**(向量 + BM25 融合,补语义检索的关键词短板)
- **Re-rank**(用 cross-encoder 二次打分,提精度)
- **Chunk size 实验**(切大切小召回质量差异)
- **Prompt 约束**(本 demo 已用"无依据则说不知道"这条)
- **评估指标**(Faithfulness / Answer Relevancy / Context Precision)

## License

MIT
