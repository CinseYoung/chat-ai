# 组内 AI 技术分享提纲(本文档同时作为 RAG 知识源)

## 1. 为什么需要 RAG
LLM 训练数据有截止时间,无法回答企业内部、私有、最新的信息。
直接用长 prompt 塞资料受上下文窗口限制,且每次调用都很贵。
RAG(Retrieval Augmented Generation)= 检索增强生成:在生成前先从外部知识库检索最相关片段,把它们作为上下文给 LLM,从而扩展模型知识边界且不必重新训练。

## 2. RAG 标准流水线
索引侧(离线):
- Loader → Chunker → Embedder → Vector Store
查询侧(在线):
- Query → Embedder → Retriever (Top-K) → Re-rank(可选) → Augmented Prompt → LLM → Answer

## 3. 关键概念
- Embedding:把文本映射为高维向量,语义相近的向量距离近。常用余弦相似度。
- 向量数据库:存向量并支持近似最近邻(ANN)搜索,如 chromem-go、Qdrant、Milvus、pgvector。
- Chunking:文本切块。粒度太大召回不精,粒度太小丢上下文,通常 200-800 字符并配 50-100 字符 overlap。
- Hybrid Search:向量召回(语义) + BM25 召回(关键词)再融合,弥补各自短板。
- Re-rank:用更强的模型(如 cross-encoder)对召回结果再打分,提升精度。
- HyDE:让 LLM 先生成"假想答案",用假想答案的向量去检索,提升召回率。

## 4. RAG vs Fine-tuning vs Long Context
- Fine-tuning:改变模型参数,适合让模型"学风格、学技能",不适合频繁更新的事实知识。
- Long Context:把所有资料塞进上下文,简单但贵且慢,且远端信息容易"被遗忘"(lost in the middle)。
- RAG:可控、可解释、可热更新,是企业内知识问答的主流方案。

## 5. 评估指标
- Faithfulness:回答是否忠实于检索内容,有没有幻觉。
- Answer Relevancy:回答是否切题。
- Context Precision/Recall:检索的上下文是否准确、完整。
- 工程上常用 Ragas、TruLens 框架做自动评估。

## 6. 常见坑
- Chunk size 拍脑袋定:建议针对自己的文档实验。
- 只做语义检索:容易漏关键词,加 BM25 hybrid 通常涨 5-10 分。
- 不做 prompt 约束:LLM 仍然会幻觉,需要明确要求"无依据则说不知道"。
- 忽视 metadata 过滤:加 source/date 过滤可极大提升召回质量。
