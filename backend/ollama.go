// Ollama 客户端:调用 /api/embeddings 生成向量,/api/chat 流式生成回答。
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Embed 把一段文本转成向量。
func Embed(ctx context.Context, baseURL, model, text string) ([]float32, error) {
	body, _ := json.Marshal(map[string]string{"model": model, "prompt": text})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/embeddings", bytes.NewReader(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embed status=%d body=%s", resp.StatusCode, raw)
	}
	var r struct{ Embedding []float32 }
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r.Embedding, nil
}

// ChatStream 调用 /api/chat,逐 token 回调 onToken。
func ChatStream(ctx context.Context, baseURL, model string, messages []Message, onToken func(string) error) error {
	body, _ := json.Marshal(map[string]any{
		"model":    model,
		"messages": messages,
		"stream":   true,
		"options":  map[string]any{"temperature": 0.3, "num_predict": 2048},
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/chat", bytes.NewReader(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama chat status=%d body=%s", resp.StatusCode, raw)
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			var chunk struct {
				Message Message `json:"message"`
				Done    bool    `json:"done"`
			}
			if json.Unmarshal(bytes.TrimSpace(line), &chunk) == nil {
				if chunk.Message.Content != "" {
					if cberr := onToken(chunk.Message.Content); cberr != nil {
						return cberr
					}
				}
				if chunk.Done {
					return nil
				}
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
