# Nexus

Nexus 是一个通用分布式推荐系统框架。

运行时采用三类节点：

- `control`：控制面 + 对外接口 + 推荐编排
- `offline`：离线推荐计算
- `online`：实时推荐计算

代码组织采用四个独立 Go module：

- `github.com/kidyme/nexus/common`：共享基础能力模块
- `github.com/kidyme/nexus/control`：控制面服务模块
- `github.com/kidyme/nexus/offline`：离线计算服务模块
- `github.com/kidyme/nexus/online`：实时计算服务模块

---

## 1. 项目定位

Nexus 是推荐基础框架，不是单一业务推荐应用。

重点能力：

- 通用数据抽象：`User`、`Item`、`Feedback`
- 可插拔 Retrieve & Rank 架构
- 分布式执行，支持可观测与可回滚
- 以 MVP 为先，逐步增强效果与扩展性

---

## 2. 节点职责（当前版本）

### 2.1 `control` 节点

`control` 负责除纯计算节点以外的大部分职责：

- 对外 API（基础 CRUD + 推荐接口）
- 推荐模块编排：
  - 聚合离线召回与实时召回
  - 统一排序/重排流程
- 配置与元数据管理
- 任务调度与集群协调
- 模型版本与注册信息管理

### 2.2 `offline` 节点

`offline` 负责批量离线计算：

- 周期性扫描用户/物品/反馈数据
- 离线召回生成（latest/popular/item2item/collaborative 等）
- 可选离线排序
- 按用户维度写入推荐缓存
- 更新缓存元信息（`digest`、`update_time`、`version`）

### 2.3 `online` 节点

`online` 负责实时计算：

- 实时召回（会话、上下文、事件驱动）
- 在线过滤与轻量重排
- 使用实时特征（短期兴趣、近期行为）
- 输出结果供 `control` 完成最终响应编排

---

## 3. 模块结构

仓库根目录仅用于组织代码，不再作为 Go module。

- `common/`
  - 共享模块，模块名为 `github.com/kidyme/nexus/common`
  - 提供 `log`、`config`、`client`、`server` 等公共能力
- `control/`
  - 独立 Go module，模块名为 `github.com/kidyme/nexus/control`
  - 对应 `cmd/control` 启动入口与服务私有实现
- `offline/`
  - 独立 Go module，模块名为 `github.com/kidyme/nexus/offline`
  - 对应 `cmd/offline` 启动入口与服务私有实现
- `online/`
  - 独立 Go module，模块名为 `github.com/kidyme/nexus/online`
  - 对应 `cmd/online` 启动入口与服务私有实现

---

## 4. 核心数据抽象

Nexus 采用通用推荐系统的数据模型：

- `User`：推荐目标对象
- `Item`：可推荐实体
- `Feedback`：用户与物品的交互事件（`type/value/timestamp`）

并采用分层存储：

- `data store`：用户/物品/反馈主数据
- `cache store`：高频推荐缓存与中间结果
- `meta store`：配置、任务状态、模型元数据
- `blob store`：模型文件、向量索引等大对象

---

## 5. 推荐流程

### 4.1 离线路径

1. 多路召回生成候选
2. 可选排序/重排
3. 按用户写入缓存
4. 按策略清理过期或陈旧缓存

### 4.2 在线路径

1. 实时召回候选
2. 读取离线缓存候选
3. 合并 + 去重 + 过滤（`excludeSet`）
4. 可选排序/重排
5. 候选不足时执行 fallback
6. 返回结果，并可选 write-back 反馈

---

## 6. 技术栈

### 5.1 运行时

- 语言：`Go 1.24`
- 通信：`REST`（必选），`gRPC`（内部与可选外部）
- 配置：`YAML/TOML + 环境变量覆盖`
- 日志：`slog` 结构化日志，开发环境使用 `tint` 彩色输出，生产环境使用 JSON Handler

### 5.2 存储（MVP 默认）

- Data Store：`MySQL`
- Cache Store：`Redis`
- Meta Store：`MySQL`
- Blob Store：`S3/MinIO`（本地开发推荐 MinIO）

### 5.3 可观测

- 指标：`Prometheus`
- 看板：`Grafana`
- 链路追踪：`OpenTelemetry`
- 日志门面：统一由 `github.com/kidyme/nexus/common/log` 提供薄封装，业务侧使用 `log.Info(...)`、`log.Error(...)`

### 5.4 部署

- 本地：`Docker Compose`
- 生产：`Docker Compose`

---

## 7. 本地开发

- 根目录不是 Go module，不执行 `go mod init`
- `common`、`control`、`offline`、`online` 各自维护自己的 `go.mod`
- 根目录使用 `go.work` 管理四个本地模块联调
- 服务模块通过完整模块路径依赖 `github.com/kidyme/nexus/common`
- 为了在模块目录内单独执行命令，服务模块保留 `replace github.com/kidyme/nexus/common => ../common` 作为本地兜底
- 开发与测试时，请进入各自模块目录执行命令

常用命令示例：

- `cd common && go test ./...`
- `cd control && go run ./cmd/control`
- `cd offline && go run ./cmd/offline`
- `cd online && go run ./cmd/online`
- 仓库根目录可使用 `go work use ./common ./control ./offline ./online` 维护工作区

日志环境切换：

- 默认按开发环境启动，使用 `tint` 彩色输出
- 设置 `NEX_ENV=production` 后切换为 JSON 日志输出

---

## 8. MVP 范围（v1）

v1 目标优先保证闭环可用与稳定，而不是一次性堆满复杂算法。

- 三服务可运行：`control`、`offline`、`online`
- 一个共享基础模块：`common`
- `User`/`Item`/`Feedback` 基础 CRUD
- 离线推荐构建与缓存写入
- 在线推荐过滤与 fallback
- 基础调度（cron + 事件触发）
- 指标与健康检查

v1 暂不包含：

- 复杂可视化 Dashboard
- 完整 FM 训练系统
- 大规模在线学习闭环

---

## 9. 开发原则

- 先做 MVP，保证可运行、可测、可发布
- 每个迭代批次都要可执行、可观测、可回滚
- 避免一次性过度设计，稳定接口边界
- 从一开始预留插件化扩展点（Retrieve & Rank、FM/LLM）
- 共享代码进入 `common` 模块，服务私有实现保留在各自模块内部
- 使用 `go.work` 管理多模块本地开发，并保留服务模块对 `common` 的本地 `replace` 兜底

---

## 10. 当前状态

Nexus 处于架构定稿与 MVP 启动阶段。

建议贡献入口：

- 模块边界契约（common/control/offline/online）
- 存储接口与适配器
- 端到端推荐闭环的 smoke test
