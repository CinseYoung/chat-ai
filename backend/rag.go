// RAG 核心:文本分块 → 向量化(chromem-go 持久化) → Top-K 检索 → 拼增强 prompt。
package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	chromem "github.com/philippgille/chromem-go"
)

const collectionName = "knowledge"

// Hit 是一条命中片段,前端展示与 prompt 拼接共用。
type Hit struct {
	Source     string  `json:"source"`
	Content    string  `json:"content"`
	Similarity float32 `json:"similarity"`
}

type Store struct {
	cfg *Config
	db  *chromem.DB
	col *chromem.Collection
}

func NewStore(cfg *Config) (*Store, error) {
	db, err := chromem.NewPersistentDB(cfg.IndexDir, false)
	if err != nil {
		return nil, fmt.Errorf("init chromem: %w", err)
	}
	col, err := db.GetOrCreateCollection(collectionName, nil, embedFn(cfg))
	if err != nil {
		return nil, err
	}
	return &Store{cfg: cfg, db: db, col: col}, nil
}

// embedFn 包装 Embed 并归一化向量。
// chromem-go v0.7 在使用自定义 embed 函数时不会自动 L2-normalize 入库向量,
// 这里手动归一化,使返回的 Similarity 落在 [-1, 1] 区间(标准余弦相似度)。
func embedFn(cfg *Config) chromem.EmbeddingFunc {
	return func(ctx context.Context, text string) ([]float32, error) {
		v, err := Embed(ctx, cfg.OllamaBaseURL, cfg.EmbedModel, text)
		if err != nil {
			return nil, err
		}
		var sum float64
		for _, x := range v {
			sum += float64(x) * float64(x)
		}
		if sum == 0 {
			return v, nil
		}
		norm := float32(math.Sqrt(sum))
		out := make([]float32, len(v))
		for i, x := range v {
			out[i] = x / norm
		}
		return out, nil
	}
}

func (s *Store) Count() int { return s.col.Count() }

// Reset 清空集合(每次重灌前调用)。
func (s *Store) Reset() error {
	if err := s.db.DeleteCollection(collectionName); err != nil {
		return err
	}
	col, err := s.db.GetOrCreateCollection(collectionName, nil, embedFn(s.cfg))
	if err != nil {
		return err
	}
	s.col = col
	return nil
}

// Query 返回 Top-K 最相似的片段。
func (s *Store) Query(ctx context.Context, q string, k int) ([]Hit, error) {
	if s.col.Count() == 0 {
		return nil, nil
	}
	if k > s.col.Count() {
		k = s.col.Count()
	}
	results, err := s.col.Query(ctx, q, k, nil, nil)
	if err != nil {
		return nil, err
	}
	hits := make([]Hit, 0, len(results))
	for _, r := range results {
		hits = append(hits, Hit{
			Source:     r.Metadata["source"],
			Content:    r.Content,
			Similarity: r.Similarity,
		})
	}
	return hits, nil
}

// Chunk 按字符长度滑动切分,带 overlap。
func Chunk(text string, size, overlap int) []string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) == 0 {
		return nil
	}
	if len(runes) <= size {
		return []string{string(runes)}
	}
	step := size - overlap
	var out []string
	for i := 0; i < len(runes); i += step {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		piece := strings.TrimSpace(string(runes[i:end]))
		if piece != "" {
			out = append(out, piece)
		}
		if end == len(runes) {
			break
		}
	}
	return out
}

// IngestReport 是 /api/ingest 的返回结构。
type IngestReport struct {
	Files  []string `json:"files"`
	Chunks int      `json:"chunks"`
}

// IngestDir 读取 dir 下所有 .md/.txt,清空集合后入库。
func IngestDir(ctx context.Context, store *Store, dir string, chunkSize, overlap int) (*IngestReport, error) {
	if err := store.Reset(); err != nil {
		return nil, err
	}
	report := &IngestReport{}

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		chunks := Chunk(string(raw), chunkSize, overlap)
		if len(chunks) == 0 {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		if rel == "" {
			rel = filepath.Base(path)
		}
		docs := make([]chromem.Document, 0, len(chunks))
		for i, c := range chunks {
			docs = append(docs, chromem.Document{
				ID:       fmt.Sprintf("%s#%d", rel, i),
				Content:  c,
				Metadata: map[string]string{"source": rel},
			})
		}
		if err := store.col.AddDocuments(ctx, docs, 4); err != nil {
			return fmt.Errorf("add %s: %w", rel, err)
		}
		report.Files = append(report.Files, rel)
		report.Chunks += len(chunks)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}

// ----- Prompt 模板 -----

const systemNoRAG = `你是一个本地聊天助手,基于通用知识回答用户问题。
若不确定就直说"我不确定",不要编造。`

const systemRAG = `你是一个严格遵守"参考资料"的问答助手。

回答规则:
1. 只能基于下方【参考资料】回答,不能引入资料外的信息。
2. 若资料中无相关内容,必须明确回答"资料中未提及"。
3. 关键事实后用 [n] 标注引用了第几条资料,例如:周会时间是周三下午 4 点 [1]。
4. 回答简洁,不要重复整段资料。`

func buildRAGUser(question string, hits []Hit) string {
	var sb strings.Builder
	sb.WriteString("【参考资料】\n")
	if len(hits) == 0 {
		sb.WriteString("(无)\n")
	}
	for i, h := range hits {
		fmt.Fprintf(&sb, "[%d] (来源: %s, 相似度: %.3f)\n%s\n\n",
			i+1, h.Source, h.Similarity, h.Content)
	}
	sb.WriteString("\n【问题】\n")
	sb.WriteString(question)
	return sb.String()
}
