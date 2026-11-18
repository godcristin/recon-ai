package biz

// 差异分类枚举（与 prompt 中的标签一致）
const (
	CategoryCrossDay         = "cross_day"
	CategoryFeeDiff          = "fee_diff"
	CategoryRefundUnsync     = "refund_unsync"
	CategoryMissingInSystem  = "missing_in_system"
	CategoryMissingInChannel = "missing_in_channel"
	CategoryDuplicate        = "duplicate"
)

// 建议动作枚举
const (
	SuggestionConfirm     = "confirm"
	SuggestionInvestigate = "investigate"
	SuggestionIgnore      = "ignore"
)

func validCategory(c string) bool {
	switch c {
	case CategoryCrossDay, CategoryFeeDiff, CategoryRefundUnsync,
		CategoryMissingInSystem, CategoryMissingInChannel, CategoryDuplicate:
		return true
	}
	return false
}

func validSuggestion(s string) bool {
	switch s {
	case SuggestionConfirm, SuggestionInvestigate, SuggestionIgnore:
		return true
	}
	return false
}

// Record 归因所需的记录视图（只保留与判断相关的字段）
type Record struct {
	TradeNo   string
	BizNo     string
	Amount    int64
	Fee       int64
	TradeTime string
	Status    string
	Remark    string
}

// Input 一次归因的输入：金额差与时间差由调用方计算好传入
type Input struct {
	DiffType      string
	Channel       *Record
	System        *Record
	AmountDiff    int64
	TimeDiffHours int64
}

// Output 归因结果
type Output struct {
	Category      string
	Confidence    float64
	Reason        string
	Evidence      []string
	Suggestion    string
	Fallback      bool
	Model         string
	PromptVersion string
	TokensIn      int32
	TokensOut     int32
	LatencyMS     int64
	Cached        bool
}
