package biz

const (
	CategoryCrossDay         = "cross_day"
	CategoryFeeDiff          = "fee_diff"
	CategoryRefundUnsync     = "refund_unsync"
	CategoryMissingInSystem  = "missing_in_system"
	CategoryMissingInChannel = "missing_in_channel"
	CategoryDuplicate        = "duplicate"
)

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

type Record struct {
	TradeNo   string
	BizNo     string
	Amount    int64
	Fee       int64
	TradeTime string
	Status    string
	Remark    string
}

type Input struct {
	DiffType      string
	Channel       *Record
	System        *Record
	AmountDiff    int64
	TimeDiffHours int64
}

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
