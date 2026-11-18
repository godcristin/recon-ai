# recon-ai · 对账 AI 助手

> 上传渠道账单与业务系统订单，自动匹配、AI 归因差异、人工确认。
> 把财务每天 2 小时的 Excel 手工对账，压缩到 15 分钟。

---

## 解决什么问题

某民办学校通过微信/支付宝收学费，渠道 T+1 出对账单。财务每天要把**渠道账单**和**系统订单**逐笔比对：

- 2,000 行 Excel 手工 VLOOKUP，每天 2~2.5 小时；
- 找出 30~50 条差异后，还要靠经验猜原因（跨日？手续费？退款？）；
- 处置结论留在个人 Excel 里，新人无法复用，审计也查不到依据。

本系统把这件事变成：**上传两个文件 → 10 秒出差异清单 → AI 给归因 → 点确认/纠正 → 出报告**。

---

## 主要能力

| 能力 | 说明 |
|---|---|
| **账单归一化** | 渠道账单（元）与系统订单（分）统一为分；脏数据跳过并报行号 |
| **三级匹配** | 单号精确 → 金额+时间窗+唯一候选 → 单边差异；**不确定不猜** |
| **差异管理** | 分页、按分类/状态筛选、金额合计、分类分布 |
| **AI 归因** | 6 类分类 + 人话理由 + 依据字段 + 置信度 + 建议动作 |
| **人机协同** | 确认 / 纠正，纠正记录沉淀为评测样本 |
| **优雅降级** | AI 不可用时，对账、列表、确认、汇总**全部照常可用** |
| **成本可控** | 懒归因（点开才调）+ 结果缓存，2,000 行任务只调用几次模型 |

---

## 架构

```
React (Vite) ──HTTP──> recon-bff ──gRPC──> recon-core ──> MySQL
                           │
                           └──gRPC──> recon-ai ──HTTPS──> LLM
```

| 服务 | 职责 | 不做什么 |
|---|---|---|
| **recon-bff** | 对外 HTTP、协议转换、聚合、降级 | 不查库、不写业务规则 |
| **recon-core** | 解析、匹配、差异、状态机、事务 | 不调用模型 |
| **recon-ai** | prompt、模型调用、输出校验、规则兜底 | 不碰业务库 |

**依赖方向固定**：`bff → core`、`bff → ai`；**core 不依赖 ai**。
这是「AI 挂了不影响对账」的物理保证，不是口头承诺。

---

## 技术栈

| 层 | 选型 | 理由 |
|---|---|---|
| 后端 | Go 1.22 + Kratos v2 + Wire | 三服务需要 gRPC、服务发现、统一错误码 |
| 契约 | protobuf + `google.api.http` | 一份 proto 同时产出 gRPC 与 HTTP |
| 存储 | MySQL 8 + GORM | 团队现有技术栈；事务与聚合够用 |
| AI | DeepSeek（OpenAI 兼容协议） | 直接 `net/http` 调用，便于控制超时/重试/计量 |
| 前端 | Vite + React 18 + TypeScript + Tailwind | 生态成熟、组件丰富、构建快 |
| 部署 | Docker Compose | 一键起全套，换机器可复现 |

---

## 快速开始

### 1. 准备数据库

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

### 3. 配置模型 Key

```powershell
Copy-Item .env.example .env
# 编辑 .env，填入 DEEPSEEK_API_KEY
```

### 4. 启动后端（编译 + 起三个服务）

```powershell
pwsh -File .\scripts\dev.ps1
```

健康检查：

```
http://127.0.0.1:8000/healthz   core 存活
http://127.0.0.1:8000/readyz    core 就绪（含数据库）
http://127.0.0.1:8001/healthz   ai 存活
```

### 5. 启动前端

```powershell
cd web
npm install
npm run dev
```

打开 `http://localhost:5173`，上传 `testdata/channel.csv` 与 `testdata/system.csv`。

### 6. 停止

```powershell
pwsh -File .\scripts\stop-dev.ps1
```

---

## 目录结构

```
api/                 proto 契约（唯一真相）
├── comm/v1/         公共错误码
├── recon/v1/        核心服务：任务、差异、确认
├── ai/v1/           AI 归因
└── app/v1/          BFF 对外 HTTP

app/
├── core/            对账核心：解析、匹配、事务
│   └── internal/{conf,server,service,biz,data,domain,parser}
├── ai/              AI 服务：prompt、模型客户端、校验、兜底
│   └── internal/{conf,server,service,biz,llm,prompts}
└── bff/             HTTP 网关：聚合、降级
    └── internal/{conf,server,service,biz,data}

cmd/
├── smoke/           gRPC 冒烟客户端
├── smoke-ai/        AI 冒烟客户端
├── gen-data/        测试数据生成器
└── eval/            AI 归因评测工具

web/                 Vite + React 前端
ai/eval/             评测数据集与报告
db/migrations/       SQL 迁移
docs/                设计文档
scripts/             代码生成、迁移、启停脚本
third_party/         proto 依赖
```

---

## 关键设计

### 1. 三级匹配：宁可判差异，也不猜

| 级别 | 规则 | 失败处理 |
|---|---|---|
| L1 精确 | 商户单号相同 | 进入 L2 |
| L2 模糊 | 金额相同 + 时间差 ≤24h + **窗口内唯一候选** | 进入 L3 |
| L3 剩余 | 单边差异 | — |

模糊匹配只要出现**两个及以上候选**就放弃匹配，全部判为差异交给人工。
理由：财务场景里错配比漏配更危险——错配会把真实差异"吞掉"。

### 2. 懒归因 + 回写

```
用户点开差异 → BFF 查 core → 若未归因 → 调 ai → 回写 core → 返回
```

- 2,000 行任务只产生 60 条差异，只对**被点开的**调用模型；
- 调用结果落库，第二次打开零成本；
- 模型失败返回 `PENDING` 而不是报错，前端显示"待归因"。

### 3. AI 的四道约束

| 约束 | 实现 |
|---|---|
| 格式 | JSON 解析 + 枚举校验，失败重试 ≤2 次 |
| 防编造 | `evidence` 只能引用输入中存在的字段，编造一次置信度 ×0.8 |
| 不算钱 | 金额差由代码计算好传入，prompt 明确禁止模型复述/计算金额 |
| 可降级 | 模型不可用时按差异类型规则兜底，标记 `fallback=true` |

### 4. 事务边界

建任务、写记录、匹配、写差异、更新统计全部在**一个事务**内完成；
任一步失败整体回滚，不会出现"有任务没差异"的脏数据。

---

## 实测数据

| 指标 | 结果 |
|---|---|
| 对账性能 | 1,960 条记录 → **210 ms**（同一事务内完成解析+匹配+落库） |
| AI 归因延迟 | 平均 **1.0 s**（首次）、**1~3 ms**（缓存命中） |
| 评测准确率 | 自建 30 条样本：规则基线 **50%** → prompt v2 **100%**（存在过拟合，详见 `ai/eval/README.md`） |
| 降级验证 | 停掉 AI 服务后，上传/列表/确认/汇总 **全部可用** |

---

## 文档

| 文档 | 内容 |
|---|---|
| [docs/00-项目立项书.md](docs/00-项目立项书.md) | 场景、用户、需求、概要/详细设计、测试策略、问题库 |
| [docs/decisions.md](docs/decisions.md) | 11 条架构决策记录（ADR） |
| [docs/演示脚本.md](docs/演示脚本.md) | 3 分钟演示流程与口播词 |
| [ai/eval/README.md](ai/eval/README.md) | 评测方法论与过拟合说明 |
| [web/README.md](web/README.md) | 前端说明与常见问题 |

---

## 已知限制

- 只支持单一渠道（微信）账单格式，未做多渠道适配；
- 分类阈值（如 23:00–00:59 判跨日）来自样本拟合，需按真实业务口径校准；
- 未做多租户与权限体系，登录态为占位实现；
- 评测集为自建样本，真实场景准确率预期 75%~85%。
