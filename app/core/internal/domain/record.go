package domain

import "time"

const (
	SourceChannel int32 = 1
	SourceSystem  int32 = 2
)

type Record struct {
	ID        int64
	TaskID    int64
	Source    int32
	TradeNo   string
	BizNo     string
	Amount    int64
	Fee       int64
	TradeTime time.Time
	Status    string
	Remark    string
}
type DiffType string

const (
	DiffChannelOnly    DiffType = "channel_only"
	DiffSystemOnly     DiffType = "system_only"
	DiffAmountMismatch DiffType = "amount_mismatch"
)

type MatchRule string

const (
	MatchExact MatchRule = "exact"
	MatchFuzzy MatchRule = "fuzzy"
	MatchNone  MatchRule = "none"
)

type Diff struct {
	ChannelRecordID int64
	SystemRecordID  int64
	Type            DiffType
	AmountDiff      int64
	MatchRule       MatchRule
}
