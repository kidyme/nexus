# Nexus 项目交接 TODO

## 已完成

- 已明确三服务职责边界：
  - `control`：负责数据 CRUD、对外接口、变更记录
  - `offline`：负责离线召回、中间产物构建、推荐缓存刷新
  - `online`：预留给在线排序与实时决策
- 已完成 `control` 侧 `user`、`item`、`feedback` 的 DDD 风格实现
- 已补齐 `control` 的批量操作、分页能力与 OpenAPI 示例
- 已修复 `control` 中幂等更新误报 `404` 的问题
- 已为 `control` 接入 Redis 变更元数据记录：
  - `last_modify_user_time/<userID>`
  - `last_modify_item_time/<itemID>`
- 已统一 `control` 与 `offline` 的配置加载方式，解决工作目录敏感问题
- 已补充本地 Redis 环境配置

## 离线推荐当前状态

- 已实现串行 worker 调度语义：
  - 跑完一轮
  - 等待一段时间
  - 再跑下一轮
- 已实现推荐结果写入 Redis：
  - `recommend_cache/<userID>`
  - `recommend_update_time/<userID>`
  - `recommend_digest/<userID>`
- 已实现用户是否需要刷新的判断逻辑，当前依据包括：
  - 用户是否活跃
  - 缓存是否不存在
  - 缓存是否过期
  - 配置摘要是否变化
  - 用户修改时间是否晚于缓存更新时间
- 已实现 recaller 注册表与启动时校验
  - 配置了系统不支持的 recaller 时，`offline` 会启动失败而不是静默吞掉
- 已实现召回配额分配与剩余额度动态再分配

## 已实现的召回器

- `non_personal/popular`
- `non_personal/latest`
- `cf/item_to_item/users`
- `cf/item_to_item/tags`
- `cf/item_to_item/embedding`
- `cf/item_to_item/auto`

## Item-to-Item 当前实现

- 已实现统一的 `item_to_item` recaller，内部按 strategy 分不同 `type`
- 已支持离线预计算 item 邻居
- 已支持基于 digest 的 Redis 邻居缓存复用
- 当前各类型含义如下：
  - `users`：基于共同正反馈用户计算 item 相似度
  - `tags`：基于类别/标签重叠计算 item 相似度
  - `embedding`：基于 `labels` 中向量做余弦相似度
  - `auto`：按权重融合 `users` 和 `tags`

## 配置模型重构情况

- 已将 recaller 配置重构为：
  - `CommonRecallerConfig`
  - 各 recaller 对应的强类型配置 struct
- YAML 当前先解码到 `offline/config` 内部私有 raw struct
- 再由 `config` 层将 raw 配置转换为业务可用的强类型配置
- 应用层 / 业务层已不再依赖“扁平大一统”的 recaller config
- 命名已统一从 `Base` 调整为 `Common`

## 代码结构优化

- 已将 `item_to_item` 相关实现迁移到：
  - `offline/internal/application/recall/cf/item/`
- 已按策略拆分文件：
  - `users.go`
  - `tags.go`
  - `embedding.go`
  - `auto.go`
- 离线推荐整体分层风格已尽量与 `control` 保持一致

## 测试与验证

- `go test ./offline/...` 当前通过
- 本次涉及的 `offline` 文件 lint 已清理干净

## 待完成事项

- 实现 `cf/user_to_user`
- 实现 `cf/mf`
- 完善训练抽象与真实训练流程
- 完成 `online` 的排序 / 服务链路
- 评估 `offline` 配置是否继续演进为更清晰的 `spec:` 风格
- 复查并串联架构文档，给接手人一个明确阅读顺序

## 重要设计结论

- `control` 负责记录数据变化，但不负责理解推荐算法影响范围
- `offline` 负责召回计算、离线中间产物以及缓存刷新策略
- recaller 的扩展方式是：注册表 + 强类型配置 + 按类型拆分策略
- 配置转换逻辑应留在 `config` 层，不应散落到业务层

## 建议接手阅读顺序

1. `notes/arch.md`
2. `notes/offline-refresh-flow.md`
3. `offline/config/config.go`
4. `offline/internal/offline.go`
5. `offline/internal/application/recommendation/service.go`
6. `offline/internal/application/recall/cf/item/`
