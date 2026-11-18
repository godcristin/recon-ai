package model

import "time"

// ReconTask 对账任务
type ReconTask struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Name       string     `gorm:"column:name"`
	Status     string     `gorm:"column:status"`
	TotalCount int32      `gorm:"column:total_count"`
	DiffCount  int32      `gorm:"column:diff_count"`
	DiffAmount int64      `gorm:"column:diff_amount"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	FinishTime *time.Time `gorm:"column:finish_time"`
}

func (ReconTask) TableName() string {
	return "recon_task"
}
