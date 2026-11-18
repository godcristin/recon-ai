# recon-web

对账助手前端：Vite + React 18 + TypeScript + Tailwind CSS。

## 环境要求

- Node.js **18+**（`node -v` 确认）
- pnpm（可选，用 npm 也行）

## 启动

```powershell
cd D:\dsh\recon-ai\web
pnpm install        # 或 npm install
pnpm dev            # 或 npm run dev
```

打开 http://localhost:5173

> 开发服务器把 `/api` 代理到 `http://127.0.0.1:8002`（BFF），见 `vite.config.ts`。
> 所以启动前端之前，先把 core / ai / bff 三个服务起起来。

## 依赖的三个后端服务

```powershell
# 分别在三个终端，或在仓库根目录执行
go run ./app/core/cmd/server -conf configs/core.yaml   # gRPC 9000
go run ./app/ai/cmd/server   -conf configs/ai.yaml     # gRPC 9001（需要 .env 里的 Key）
go run ./app/bff/cmd/server  -conf configs/bff.yaml    # HTTP 8002
```

## 构建

```powershell
pnpm build          # 产物在 dist/，可直接交给 nginx
pnpm preview        # 本地预览构建结果
```

## 目录

```
src/
├── api/            client.ts 是唯一的请求出口；types.ts 与 proto 对齐
├── components/     DiffTable（列表）、DiffDrawer（详情抽屉）
├── pages/          UploadPage（上传）、TaskPage（任务详情）
├── lib/format.ts   金额换算、枚举中文名、置信度配色
└── App.tsx         两个视图之间切换（未引入路由）
```

## 三个容易踩的坑

1. **int64 是字符串**：protojson 把 int64 输出成 `"8"` 而不是 `8`，因为 JS 数字只有 53 位精度。`types.ts` 里的 ID、金额字段都声明为 `string`。
2. **零值字段会被省略**：protojson 默认不输出零值，所以类型里大量用 `?`，取值时要给默认值。
3. **点开差异会慢 1~2 秒**：这时后端在调模型做归因（第一次）。同一个差异第二次点开是毫秒级，因为结果已回写数据库。

## 常见报错

| 现象 | 原因 |
|---|---|
| 页面空白、控制台 404 `/api/...` | BFF（8002）没启动 |
| 上传报 `CREATE_TASK_FAILED` | core 没启动，或 MySQL 没起 |
| 差异详情一直「AI 暂未给出结论」 | ai 服务没起，或 `.env` 里没有 `DEEPSEEK_API_KEY` |
| `pnpm install` 卡住 | 换 npm 试试，或配国内镜像 `npm config set registry https://registry.npmmirror.com` |
