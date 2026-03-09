# SmartStream Recommender: 智流推荐引擎
基于大模型与实时流的智能推荐系统

## 项目目标
构建一个端到端的实时个性化推荐系统，能够：
- 捕获用户行为（点击、搜索、浏览）；
- 利用大模型理解用户意图与内容语义；
- 在 **<500ms** 内返回个性化推荐结果；
- 支持自然语言交互式推荐（如“帮我找适合露营的装备”）；
- 兼容电商、内容平台、短视频等场景。

## 核心功能

| 模块 | 功能描述 |
|------|--------|
| **1. 实时数据采集** | 前端埋点上报用户行为至 Kafka |
| **2. 实时特征计算** | Flink 实时生成用户短期兴趣向量、行为序列 |
| **3. 大模型 Embedding** | 使用开源大模型为商品/内容生成高质量语义向量 |
| **4. 向量召回** | Milvus 实现语义相似度召回（支持多路融合） |
| **5. 排序与重排** | 轻量级排序模型 + LLM 重排（可选） |
| **6. 生成式推荐** | 支持 RAG 生成个性化推荐理由或方案 |
| **7. 反馈闭环** | 用户反馈回流，持续优化实时画像 |

## 架构设计（当前版本）

### 1) 角色划分

- **Main**：模型训练、配置管理、任务调度、集群管理、Dashboard/管理 API
- **Offline**：离线生成推荐（多路召回 + 合并 + 排序 + 缓存）
- **Online**：在线推荐 API（读取缓存 + 兜底策略）
- **In-one**：单进程模式（Master + 离线推荐 + 在线 API）

### 2) 数据与存储

- **数据源**：用户、物品、反馈（行为）
- **存储层**：
  - **DataStore**：原始数据（用户/物品/反馈）
  - **CacheStore**：推荐结果、相似度、热门榜等缓存
  - **MetaStore**：配置/元信息
  - **Blob**：模型文件

### 3) 推荐管道（核心流程）

- **召回（多路推荐器）**：
  - non-personalized / item-to-item / user-to-user / collaborative / external / latest
- **合并候选**
- **排序（可选）**：FM/LLM 等 ranker
- **放回（replacement）**：允许已读物品再次出现
- **兜底（fallback）**：候选不足时补充

### 4) 训练流程

- 由 Main 定时任务触发
- 加载数据集 → 训练 CF / CTR → 生成模型 → 下发/保存
- Worker 用模型离线产出推荐缓存

### 5) 在线请求流程

- Online 接收请求 → 从缓存读取推荐 →（必要时兜底）→ 返回结果
- 在线请求不训练，训练在 Main 周期任务中完成

### 6) 适合你借鉴的结构

- **入口层**：CLI/配置
- **调度层**：训练与任务周期
- **模型层**：CF / CTR / 召回策略
- **数据层**：Data/Cache/Meta/Blob
- **服务层**：API + Dashboard

## 技术栈（2026 年主流开源方案）

| 层级 | 技术选型 | 说明 |
|------|--------|------|
| **数据采集** | Web SDK (自研 / Sentry-like) + REST API | 上报点击、曝光、搜索词 |
| **消息队列** | **Apache Kafka** | 高吞吐、持久化事件总线 |
| **流处理引擎** | **Apache Flink** | 实时特征工程、会话切分、窗口聚合 |
| **大模型（Embedding）** | **BGE-M3** / **Jina Embeddings v2** / **Qwen-Audio**（多模态） | 开源 SOTA embedding 模型，支持多语言、多粒度 |
| **大模型（生成）** | **Qwen-1.8B-Chat** / **Llama3-8B-Instruct**（可选） | 用于 RAG 生成推荐文案 |
| **向量数据库** | **Milvus**（或 **FAISS** for MVP） | 支持 HNSW、IVF_PQ 等高效索引 |
| **特征存储** | **Redis**（实时） + **Feast**（可选） | 存储用户实时向量、标签 |
| **Web 服务** | **FastAPI**（Python） or **Spring Boot**（Java） | 提供推荐 API |
| **部署与编排** | **Docker + Docker Compose**（MVP） / **Kubernetes**（生产） | 容器化部署 |
| **监控** | **Prometheus + Grafana** | 监控延迟、吞吐、错误率 |
| **日志** | **ELK**（Elasticsearch + Logstash + Kibana） or **Loki** | 行为日志分析 |

> ✅ 全部基于开源技术，无商业依赖，可本地/云上部署。

## 数据流（简化）

[User]
↓ (click/search/view)
[Frontend SDK] → HTTP → [Ingestion API] → Kafka (topic: user_events)
↓
[Flink Job]
↓
Real-time User Vector + Short-term Profile → Redis
↓
[Recommendation API] ← Milvus (recall by vector similarity)
↓
[Optional: LLM Rerank / Generate]
↓
[Top-K Items + Reason]
↓
[User App]

## 典型使用场景（Demo 示例）

1. **用户搜索**：“适合夏天的连衣裙”  
   → LLM 将 query 转为 embedding  
   → Milvus 召回语义相似商品  
   → 返回 Top 10 + 生成文案：“清爽棉麻材质，透气不粘身”

2. **用户刚看了露营视频**  
   → Flink 实时更新兴趣向量  
   → 下次打开首页，推荐“帐篷、睡袋、便携炉”  
   → 延迟 < 300ms

3. **运营活动**：实时监控“加购未支付”用户  
   → 触发 Flink 规则  
   → 5 分钟后推送优惠券（通过推荐位）

## 性能指标（MVP 目标）

| 指标 | 目标值 |
|------|-------|
| 端到端延迟 | < 500 ms（P95） |
| 吞吐能力 | 1,000 events/sec（单机） |
| 召回相关性 | Recall@10 > 70%（人工评估） |
| 资源占用 | ≤ 16 GB RAM, 无 GPU（CPU 推理） |

## 扩展方向（进阶）

- ✅ **多模态推荐**：结合图像（CLIP）+ 文本（BGE）embedding  
- ✅ **在线学习**：Flink + TensorFlow 实现分钟级模型更新  
- ✅ **A/B 测试框架**：集成推荐策略实验平台  
- ✅ **隐私保护**：联邦学习 or 端侧 embedding 生成  

## 开源参考项目

- [**realtime-recsys-demo**](https://github.com/feast-dev/feast/tree/master/examples)（Feast + Flink）
- [**milvus-bootcamp**](https://github.com/milvus-io/bootcamp)（含推荐场景）
- [**BAAI/bge**](https://github.com/FlagOpen/FlagEmbedding)（SOTA embedding 模型）
- [**apache/flink-ml**](https://github.com/apache/flink-ml)（Flink 机器学习）
