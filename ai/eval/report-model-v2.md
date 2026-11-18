# AI 归因评测报告

- 生成时间：2026-11-17 19:41:47
- 样本数：30（调用失败 0）
- 准确率：**30/30 = 100.0%**
- 规则兜底次数：0
- 平均延迟：1009 ms
- token 消耗：in=23942 / out=2666

## 分类别准确率

| 分类 | 正确/总数 | 准确率 |
|---|---|---|
| cross_day | 5/5 | 100% |
| fee_diff | 5/5 | 100% |
| refund_unsync | 5/5 | 100% |
| missing_in_system | 5/5 | 100% |
| missing_in_channel | 5/5 | 100% |
| duplicate | 5/5 | 100% |

## 混淆矩阵（行=期望，列=预测）

| 期望 \ 预测 | cross_day | fee_diff | refund_unsync | missing_in_system | missing_in_channel | duplicate | (空) |
|---|---|---|---|---|---|---|---|
| cross_day | 5 | 0 | 0 | 0 | 0 | 0 | 0 |
| fee_diff | 0 | 5 | 0 | 0 | 0 | 0 | 0 |
| refund_unsync | 0 | 0 | 5 | 0 | 0 | 0 | 0 |
| missing_in_system | 0 | 0 | 0 | 5 | 0 | 0 | 0 |
| missing_in_channel | 0 | 0 | 0 | 0 | 5 | 0 | 0 |
| duplicate | 0 | 0 | 0 | 0 | 0 | 5 | 0 |

## 错例

无
