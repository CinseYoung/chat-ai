// chat-ai 的 RAG 演示后端:
// - GET  /api/health           — Ollama 连接状态 + 当前 chunk 数
// - POST /api/ingest           — 重新读取 data/docs/*.md/.txt 灌入向量库
// - GET  /api/chat/stream      — SSE 流式聊天(use_rag=true 走 RAG 流程)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Config struct {
	OllamaBaseURL string
	ChatModel     string
	EmbedModel    string
	Port          string
	DataDir       string
	IndexDir      string
	ChunkSize     int
	ChunkOverlap  int
	TopK          int
}

func loadConfig() *Config {
	return &Config{
		OllamaBaseURL: env("OLLAMA_BASE_URL", "http://localhost:11434"),
		ChatModel:     env("CHAT_MODEL", "qwen3:8b"),
		EmbedModel:    env("EMBED_MODEL", "nomic-embed-text"),
		Port:          env("PORT", "8000"),
		DataDir:       env("DATA_DIR", "data/docs"),
		IndexDir:      env("INDEX_DIR", "data/index"),
		ChunkSize:     400,
		ChunkOverlap:  60,
		TopK:          3,
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	cfg := loadConfig()
	store, err := NewStore(cfg)
	if err != nil {
		log.Fatalf("init store: %v", err)
	}
	log.Printf("[chat-ai] vector store ready, chunks=%d", store.Count())

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:    []string{"*"},
	}))

	r.GET("/api/health", health(cfg, store))
	r.POST("/api/ingest", ingest(cfg, store))
	r.GET("/api/chat/stream", chatStream(cfg, store))

	addr := ":" + cfg.Port
	log.Printf("[chat-ai] listening on %s | chat=%s embed=%s", addr, cfg.ChatModel, cfg.EmbedModel)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func health(cfg *Config, store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.OllamaBaseURL+"/api/tags", nil)
		status := "disconnected"
		if resp, err := http.DefaultClient.Do(req); err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				status = "connected"
			}
		}
		c.JSON(200, gin.H{
			"ollama":      status,
			"chat_model":  cfg.ChatModel,
			"embed_model": cfg.EmbedModel,
			"docs_count":  store.Count(),
			"top_k":       cfg.TopK,
		})
	}
}

func ingest(cfg *Config, store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		report, err := IngestDir(c.Request.Context(), store, cfg.DataDir, cfg.ChunkSize, cfg.ChunkOverlap)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"ok":         true,
			"files":      report.Files,
			"chunks":     report.Chunks,
			"chunk_size": cfg.ChunkSize,
			"top_k":      cfg.TopK,
		})
	}
}

func chatStream(cfg *Config, store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		message := c.Query("message")
		if message == "" {
			c.JSON(400, gin.H{"error": "message required"})
			return
		}
		useRAG := c.Query("use_rag") == "true"

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("X-Accel-Buffering", "no")
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.JSON(500, gin.H{"error": "streaming unsupported"})
			return
		}
		send := func(event string, data any) {
			payload, _ := json.Marshal(data)
			fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, payload)
			flusher.Flush()
		}

		ctx := c.Request.Context()
		var messages []Message
		var userContent string

		if useRAG {
			hits, err := store.Query(ctx, message, cfg.TopK)
			if err != nil {
				send("error", gin.H{"content": "RAG 检索失败: " + err.Error()})
				return
			}
			send("context", gin.H{"hits": hits, "top_k": cfg.TopK})
			messages = append(messages, Message{Role: "system", Content: systemRAG})
			userContent = buildRAGUser(message, hits)
		} else {
			messages = append(messages, Message{Role: "system", Content: systemNoRAG})
			userContent = message
		}
		messages = append(messages, Message{Role: "user", Content: userContent})

		err := ChatStream(ctx, cfg.OllamaBaseURL, cfg.ChatModel, messages, func(t string) error {
			send("content", gin.H{"content": t})
			return nil
		})
		if err != nil {
			send("error", gin.H{"content": "模型调用失败: " + err.Error()})
			return
		}
		send("done", gin.H{})
	}
}
