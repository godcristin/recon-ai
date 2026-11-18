# recon-ai

对账 AI 助手：上传渠道账单与系统订单，自动匹配、AI 归因差异、人工确认。

- 三个 Kratos 服务：`recon-bff`（HTTP 网关）、`recon-core`（gRPC + MySQL）、`recon-ai`（gRPC + LLM）
- 前端：Vite + React + TypeScript
- 设计文档见 [`docs/00-项目立项书.md`](docs/00-项目立项书.md)

## 架构

```
React(Vite) ──HTTP──> recon-bff ──gRPC──> recon-core ──> MySQL
                          │
                          └────gRPC──> recon-ai ──HTTPS──> LLM
```

依赖方向固定：`bff → core`、`bff → ai`；**core 不依赖 ai**（模型不可用时对账主流程仍可用）。

## 目录

```
api/             proto 契约（唯一真相），由 scripts/gen.ps1 生成 Go 代码
app/bff/         HTTP 网关：鉴权、上传、聚合、SSE
app/core/        对账核心：任务、匹配、差异、确认
app/ai/          AI 服务：prompt、模型调用、校验、降级
cmd/smoke/       gRPC 冒烟客户端
configs/         服务配置
db/migrations/   SQL 迁移（只追加）
docs/            立项书、决策记录
scripts/         代码生成、迁移脚本
web/             前端
third_party/     proto 依赖（errors / google.api / validate / well-known types）
```

## 本地运行

### 1. 准备数据库（MySQL 8）

```sql
CREATE DATABASE IF NOT EXISTS recon CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE USER IF NOT EXISTS 'recon'@'localhost' IDENTIFIED BY 'recon123';
GRANT ALL PRIVILEGES ON recon.* TO 'recon'@'localhost';
FLUSH PRIVILEGES;
```

### 2. 建表

```powershell
pwsh -File .\scripts\migrate.ps1
```

### 3. 启动 core

```powershell
go run ./app/core/cmd/server -conf configs/core.yaml
```

健康检查：

- `http://127.0.0.1:8000/healthz` —— 存活（进程活着即 200）
- `http://127.0.0.1:8000/readyz` —— 就绪（数据库不可用返回 503）

### 4. 冒烟

```powershell
go run ./cmd/smoke
# CreateTask ok: task_id=1 total=0 diff=0
```

### 5. 前端

```powershell
cd web
pnpm install
pnpm dev          # http://localhost:5173
```

前端通过 Vite 代理把 `/api` 转发到 BFF（`127.0.0.1:8002`），所以要先起三个后端服务。
详见 [`web/README.md`](web/README.md)。

## 生成代码

```powershell
pwsh -File .\scripts\gen.ps1
```

依赖本机 `protoc` 及插件：`protoc-gen-go`、`protoc-gen-go-grpc`、`protoc-gen-go-http`、`protoc-gen-go-errors`、`protoc-gen-validate`。

> `third_party/google/protobuf/` 下已随仓库带上 well-known types，无需额外指定 protoc 安装目录。

## 状态

当前处于阶段 2（core 骨架 + 建表 + 冒烟跑通）。进度见立项书第 10 节。
