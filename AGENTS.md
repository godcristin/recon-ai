# AGENTS.md

## 0. 项目定位

**recon-ai**：对账 AI 助手。上传渠道账单与系统订单，自动匹配、AI 归因差异、人工确认。

Monorepo，三个独立部署的 Kratos 服务 + 一个前端：

- `recon-bff`：HTTP 对入口，鉴权、上传、聚合、SSE
- `recon-core`：gRPC，业务与事务，独占 MySQL
- `recon-ai`：gRPC，prompt 与模型调用，无 DB

技术栈：Go 1.22 / Kratos v2 / Wire / proto / MySQL 8 / etcd / Vite + React + TS。

当前阶段见 `docs/00-项目立项书.md` 第 10 节。**不要**自行升级架构或引入新组件。

## 1. 目录结构

```
api/{app,recon,ai,comm}/v1/    proto 契约（唯一真相）
app/bff/    cmd/server + internal/{conf,server,service,biz,data}
app/core/   cmd/server + internal/{conf,server,service,biz,data,repo}
app/ai/     cmd/server + internal/{conf,server,service,biz,llm,prompts}
web/        Vite + React；web/src/api/ 是所有请求的唯一出口
db/migrations/                 只追加
docs/                          立项书、ADR、决策日志
```

## 2. 分层契约（最高优先级）

- 依赖方向唯一：`service → biz → data → model`，禁止反向
- service 只做协议转换与校验，不写业务规则
- biz 不直接访问 `*gorm.DB`，只能通过 data 层
- data 不返回 proto 类型，只返回 model
- ai 服务不碰业务库；core 不调 ai
- 单个文件 ≤ 300 行

## 3. 契约与数据

- 改接口先改 `api/**/*.proto`，再 `make api`，禁止手改 `*_pb.go`
- `google.api.http` 注解决定 HTTP 路由，新增接口必须带
- 金额一律 `int64` 分，禁止 float
- 时间字段统一 `created_at` / `updated_at`，数据库用 `DATETIME(3)`
- 状态用 int 常量 + 状态机，禁止裸数字

## 4. AI 链路约定

- 模型调用只允许走 `app/ai/internal/llm/client.go`
- prompt 存 `app/ai/internal/prompts/*.md`，文件名带版本（`classify.v1.md`）
- 输出必须 JSON Schema 校验；失败重试 ≤ 2 次，再失败走规则兜底并标记 `fallback`
- **金额/条数/比例由代码计算，禁止模型生成或复述**
- `evidence` 字段只能引用输入中存在的字段名（白名单校验）
- 每次调用记录：model / prompt_version / tokens / latency / cached
- `temperature=0`；相同输入按 hash 缓存 7 天

## 5. 命令

```bash
make api            # proto → pb.go / http / validate
make config         # conf.proto → conf.pb.go
make wire           # 重新生成 wire_gen.go
make test           # go test ./...
make lint           # go vet + golangci-lint
cd web && pnpm lint && pnpm build
```

## 6. 验收标准（每个任务）

- [ ] `make api && make wire` 后 `go build ./...` 通过
- [ ] 相关接口 curl 返回预期 JSON（贴实际输出）
- [ ] 页面能走完主流程（贴路径或截图）
- [ ] AI 相关改动，评测准确率不低于上一版

## 7. 禁止事项

- 不改生成文件（`*_pb.go`、`*_grpc.pb.go`、`wire_gen.go`、`conf.pb.go`）
- 不新增未经确认的第三方依赖
- 不提交密钥；配置走 `.env`
- 不在 service 层写 SQL，不在 core 里调模型
- 不用 `_` 忽略 error，不用 panic 做业务控制
- 不改 `db/migrations/` 历史文件

## 8. 工作方式

- 一次一个垂直切片（见立项书第 6 节），做完立即验收
- 动手前先输出「改哪些文件、每个改什么、怎么验证」，等我确认
- 完成后执行第 5 节命令并贴结果
- 关键决策追加到 `docs/decisions.md`
