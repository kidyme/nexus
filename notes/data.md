# Control Meta Schema

## Storage

### `etcd`

默认前缀：

```text
/nexus/meta
```

当前只规划一类 key：节点注册与心跳。

#### 节点 Key

Key 约定：

```text
/nexus/meta/nodes/{service_name}/{node_id}
```

Value 示例：

```json
{
  "node_id": "online-01",
  "service_name": "online",
  "endpoint": "127.0.0.1:8082",
  "status": "online",
  "version": "v0.1.0",
  "heartbeat_at": "2026-03-09T12:00:00Z"
}
```

字段说明：

| Field | Type | Description |
| --- | --- | --- |
| `node_id` | `string` | 节点唯一标识 |
| `service_name` | `string` | 服务名，如 `control`、`offline`、`online` |
| `endpoint` | `string` | 节点对外地址 |
| `status` | `string` | 节点状态，如 `online`、`offline` |
| `version` | `string` | 节点运行版本 |
| `heartbeat_at` | `RFC3339 string` | 最近一次心跳时间 |

#### 使用方式

- 节点启动时注册自己的 key
- 节点运行中通过租约续约或定时覆盖写入
- `control` 按前缀 `/nexus/meta/nodes/` 查询当前节点列表
- 节点异常退出后，租约过期即可自动删除对应 key
