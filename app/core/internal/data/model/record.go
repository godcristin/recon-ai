package model

import "time"

// ReconRecord 账单原始记录（渠道与系统合并，source 区分）
type ReconRecord struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement"`
	TaskID    int64      `gorm:"column:task_id"`
	Source    int32      `gorm:"column:source"`
	TradeNo   string     `gorm:"column:trade_no"`
	BizNo     string     `gorm:"column:biz_no"`
	Amount    int64      `gorm:"column:amount"`
	Fee       int64      `gorm:"column:fee"`
	TradeTime *time.Time `gorm:"column:trade_time"`
	Status    string     `gorm:"column:status"`
	Remark    string     `gorm:"column:remark"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (ReconRecord) TableName() string {
	return "recon_record"
}
