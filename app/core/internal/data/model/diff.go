package model

import "time"

// ReconDiff 对账差异
type ReconDiff struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TaskID          int64     `gorm:"column:task_id"`
	ChannelRecordID *int64    `gorm:"column:channel_record_id"`
	SystemRecordID  *int64    `gorm:"column:system_record_id"`
	DiffType        string    `gorm:"column:diff_type"`
	AmountDiff      int64     `gorm:"column:amount_diff"`
	Category        string    `gorm:"column:category"`
	Confidence      float64   `gorm:"column:confidence"`
	Reason          string    `gorm:"column:reason"`
	Evidence        string    `gorm:"column:evidence"`
	Suggestion      string    `gorm:"column:suggestion"`
	State           string    `gorm:"column:state"`
	Fallback        bool      `gorm:"column:fallback"`
	AIModel         string    `gorm:"column:ai_model"`
	PromptVersion   string    `gorm:"column:prompt_version"`
	MatchRule       string    `gorm:"column:match_rule"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (ReconDiff) TableName() string {
	return "recon_diff"
}

// DiffFeedback 人工纠正记录
type DiffFeedback struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	DiffID        int64     `gorm:"column:diff_id"`
	AICategory    string    `gorm:"column:ai_category"`
	HumanCategory string    `gorm:"column:human_category"`
	Operator      string    `gorm:"column:operator"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (DiffFeedback) TableName() string {
	return "diff_feedback"
}
