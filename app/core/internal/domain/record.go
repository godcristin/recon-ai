package domain

import "time"

// 记录来源
const (
	SourceChannel int32 = 1 // 渠道账单
	SourceSystem  int32 = 2 // 系统订单
)

// Record 一条账单记录（渠道或系统），金额单位统一为分
type Record struct {
	ID        int64
	TaskID    int64
	Source    int32
	TradeNo   string // 渠道交易号
	BizNo     string // 商户单号（渠道 out_trade_no / 系统 pay_order_no）
	Amount    int64  // 分
	Fee       int64  // 分
	TradeTime time.Time
	Status    string
	Remark    string
}

// DiffType 差异类型
type DiffType string

const (
	DiffChannelOnly    DiffType = "channel_only"    // 渠道有、系统无
	DiffSystemOnly     DiffType = "system_only"     // 系统有、渠道无
	DiffAmountMismatch DiffType = "amount_mismatch" // 两侧都有但金额不一致
)

// MatchRule 命中规则
type MatchRule string

const (
	MatchExact MatchRule = "exact" // 单号精确匹配
	MatchFuzzy MatchRule = "fuzzy" // 金额 + 时间窗 + 唯一候选
	MatchNone  MatchRule = "none"  // 未匹配
)

// Diff 一条差异（纯领域对象，不含持久化字段）
type Diff struct {
	ChannelRecordID int64
	SystemRecordID  int64
	Type            DiffType
	AmountDiff      int64 // 分：渠道 - 系统；单边时为该侧金额
	MatchRule       MatchRule
}
