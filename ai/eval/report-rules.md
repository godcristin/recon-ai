# AI 归因评测报告

- 生成时间：2026-11-17 19:42:08
- 样本数：30（调用失败 0）
- 准确率：**15/30 = 50.0%**
- 规则兜底次数：30
- 平均延迟：0 ms
- token 消耗：in=0 / out=0

## 分类别准确率

| 分类 | 正确/总数 | 准确率 |
|---|---|---|
| cross_day | 0/5 | 0% |
| fee_diff | 5/5 | 100% |
| refund_unsync | 0/5 | 0% |
| missing_in_system | 5/5 | 100% |
| missing_in_channel | 5/5 | 100% |
| duplicate | 0/5 | 0% |

## 混淆矩阵（行=期望，列=预测）

| 期望 \ 预测 | cross_day | fee_diff | refund_unsync | missing_in_system | missing_in_channel | duplicate | (空) |
|---|---|---|---|---|---|---|---|
| cross_day | 0 | 0 | 0 | 3 | 2 | 0 | 0 |
| fee_diff | 0 | 5 | 0 | 0 | 0 | 0 | 0 |
| refund_unsync | 0 | 0 | 0 | 0 | 5 | 0 | 0 |
| missing_in_system | 0 | 0 | 0 | 5 | 0 | 0 | 0 |
| missing_in_channel | 0 | 0 | 0 | 0 | 5 | 0 | 0 |
| duplicate | 0 | 0 | 0 | 5 | 0 | 0 | 0 |

## 错例

| ID | 期望 | 预测 | 置信度 | 兜底 |
|---|---|---|---|---|
| cd-1 | cross_day | missing_in_system | 0.60 | true |
| cd-2 | cross_day | missing_in_channel | 0.60 | true |
| cd-3 | cross_day | missing_in_system | 0.60 | true |
| cd-4 | cross_day | missing_in_channel | 0.60 | true |
| cd-5 | cross_day | missing_in_system | 0.60 | true |
| dp-1 | duplicate | missing_in_system | 0.60 | true |
| dp-2 | duplicate | missing_in_system | 0.60 | true |
| dp-3 | duplicate | missing_in_system | 0.60 | true |
| dp-4 | duplicate | missing_in_system | 0.60 | true |
| dp-5 | duplicate | missing_in_system | 0.60 | true |
| ru-1 | refund_unsync | missing_in_channel | 0.60 | true |
| ru-2 | refund_unsync | missing_in_channel | 0.60 | true |
| ru-3 | refund_unsync | missing_in_channel | 0.60 | true |
| ru-4 | refund_unsync | missing_in_channel | 0.60 | true |
| ru-5 | refund_unsync | missing_in_channel | 0.60 | true |
