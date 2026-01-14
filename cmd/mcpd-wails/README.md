# Wails 应用入口说明

> 当前 Wails 应用入口已迁移至仓库根目录的 `app.go`，本目录仅保留历史说明。

## 文件结构

```
app.go                   # Wails 应用入口
internal/ui/
├── service.go           # ServiceRegistry（统一注册入口）
├── service_deps.go      # 共享依赖容器
├── core_service.go      # CoreService
├── discovery_service.go
├── config_service.go
├── profile_service.go
├── runtime_service.go
├── log_service.go
├── subagent_service.go
├── system_service.go
└── debug_service.go
```

## 设计理念

### 1. 入口文件职责

- **入口只做启动与依赖注入**：创建 logger、core app、Manager，并注册 services。
- **不在入口层堆积业务逻辑**：业务逻辑仍在 `internal/app` 与 `internal/infra`。

### 2. 服务分层

```
app.go
  ↓ (启动、注册、转发)
internal/ui/*_service.go
  ↓ (桥接、事件流)
internal/app/app.go
  ↓ (核心编排)
internal/domain + internal/infra
  (领域逻辑)
```

### 3. 符合架构约定

- ✅ Wails 入口**只做启动与依赖注入**
- ✅ **不在入口层堆积业务逻辑**
- ✅ URL Scheme **在入口层注册**，解析转发给 `internal/ui`
- ✅ 核心逻辑仍在 `internal/app`，与 CLI 共享

## 使用方式

### 开发模式

```bash
# 使用 Wails CLI 运行开发服务器
wails3 dev

# 或使用 Makefile
make wails-dev
```

### 构建

```bash
# 构建所有平台
wails3 build

# 构建特定平台
wails3 build -platform darwin/arm64
```

## URL Scheme 处理流程

1. **注册**：在 `app.go` 的 `application.Options.Protocols` 中注册
2. **接收**：通过 `ApplicationOpenedWithURL` 事件接收
3. **转发**：入口层调用 `SystemService.HandleURLScheme(url)`
4. **处理**：`internal/ui/system_service.go` 解析并发送前端事件

示例 URL：
```
mcpd://open/server?id=123
mcpd://settings/profiles
```

## 扩展点

### 添加新的导出方法

在 `internal/ui/*_service.go` 中新增导出方法（首字母大写），并在 `ServiceRegistry` 注册对应 service：

```go
// GetServerList returns server names.
func (s *ProfileService) GetServerList(ctx context.Context) ([]string, error) {
    return []string{"server1", "server2"}, nil
}
```

Wails 会自动生成 JS 绑定到前端。

### 发送事件到前端

```go
emitRuntimeStatusUpdated(s.deps.wailsApp(), snapshot)
```

## 参考

- `docs/WAILS_STRUCTURE.md` - Wails 工程结构建议
- `docs/STRUCTURE.md` - 整体项目结构
