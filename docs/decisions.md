# 决策记录（ADR）

> 只记「为什么」，不记「怎么做」。每做一次影响面较大的决定就追加一条。

## ADR-001 拆三个服务而不是一个

- 日期：2026-11-17
- 决策：`recon-bff` / `recon-core` / `recon-ai` 三个独立进程。
- 理由：BFF 负责对外协议与聚合，core 负责业务与事务，ai 负责模型调用；三者变更频率、扩缩容需求、失败模式都不同。
- 放弃：单体。单体在简历上表达不出服务治理，也无法演示「AI 挂了主流程仍可用」。
- 代价：本地要起三个进程 + etcd（开发期用直连规避）。

## ADR-002 后端选 Kratos

- 日期：2026-11-17
- 决策：Kratos v2 + Wire + proto-first。
- 理由：三服务需要 gRPC、服务发现、统一错误码；且与既有工作经验一致。
- 放弃：Gin。单服务够用，但表达不出服务间契约治理。
- 代价：每次改接口要重新生成，脚手架成本高。

## ADR-003 proto 是唯一真相

- 日期：2026-11-17
- 决策：所有接口先改 `.proto`，再 `scripts/gen.ps1` 生成；禁止手改 `*_pb.go`。
- 理由：三个服务 + 前端共四份代码，契约漂移是最大的隐性成本。
- 代价：改字段多一步生成。

## ADR-004 core 不依赖 ai

- 日期：2026-11-17
- 决策：AI 归因由 BFF 编排调用，core 只提供数据与写回接口。
- 理由：模型调用是秒级、会失败、成本敏感；不能让它出现在核心交易链路上。
- 效果：AI 服务宕机时，对账、列表、导出仍 100% 可用（见立项书 5.6 降级矩阵）。

## ADR-005 单仓库单 go.mod

- 日期：2026-11-17
- 决策：`recon-ai` 一个 module，三个服务放在 `app/<svc>/` 下。
- 理由：共享 proto 与 model，避免跨仓库拷贝契约（公司做法是每服务一仓库 + 手动同步 proto）。
- 代价：服务不能独立发版，但本项目不需要。

## ADR-006 well-known types 随仓库入库

- 日期：2026-11-17
- 决策：把 `google/protobuf/*.proto` 复制进 `third_party/`，而不是依赖本机 protoc 的 include 目录。
- 理由：① 本机路径含中文，作为 protoc 参数时被按 ANSI 编码传给子进程，路径损坏导致 `File not found`；② 仓库自包含，换机器不用配 `-I`。
- 代价：仓库多约 10 个小文件；protoc 升级时 WKT 不会自动跟随（影响可忽略）。

## ADR-007 数据库惰性连接 + 存活/就绪探针分离

- 日期：2026-11-17
- 背景：阶段 2 首次启动时本机还没有 PostgreSQL，`gorm.Open` 会立即建连并失败，导致进程直接 panic。
- 决策：`Data.DB(ctx)` 首次使用时才 `gorm.Open`；`/healthz` 只反映进程存活（200），`/readyz` 真正 ping 数据库（不可用返回 503）。
- 理由：① 服务启动不应被外部依赖绑死；② 探针语义正确，K8s 的 liveness/readiness 正好对应这两者；③ 演示时可以先起服务、后起数据库。
- 代价：数据库故障要到首个业务请求或 `/readyz` 才暴露，因此部署时必须配 readiness 探针。

## ADR-008 数据库从 PostgreSQL 改为 MySQL 8

- 日期：2026-11-17
- 背景：开发机已有 MySQL 8.0（含 Workbench），没有 PostgreSQL；为一个个人项目再装一套 PG 不划算。
- 决策：改用 MySQL 8.0 + `gorm.io/driver/mysql`；迁移脚本改为 MySQL 方言（`AUTO_INCREMENT`、`DATETIME(3)`、`JSON`、`TINYINT(1)`）。
- 影响：
  - 插入改为 GORM `Create` 取 `LastInsertId`（MySQL 无 `RETURNING`）；
  - `evidence` 用 `JSON` 而非 `JSONB`，复杂 JSON 查询能力弱一些（本项目只做整存整取，无影响）；
  - 与公司技术栈（MySQL）一致，面试讲述时更贴近真实经验。
- 代价：失去 PG 的部分高级能力（部分索引、GIN 索引），但本项目用不到。
