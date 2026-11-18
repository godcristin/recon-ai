package data

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"recon-ai/app/core/internal/data/model"
	"recon-ai/app/core/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// ReconRepo 对账任务、记录与差异的持久化
type ReconRepo struct {
	data *Data
	log  *log.Helper
}

func NewReconRepo(d *Data, logger log.Logger) *ReconRepo {
	return &ReconRepo{
		data: d,
		log:  log.NewHelper(log.With(logger, "module", "data/recon")),
	}
}

// CreateTask 插入一条对账任务，返回自增 ID
func (r *ReconRepo) CreateTask(ctx context.Context, name string) (int64, error) {
	db, err := r.data.DB(ctx)
	if err != nil {
		return 0, err
	}

	task := &model.ReconTask{
		Name:      name,
		Status:    "CREATED",
		CreatedAt: time.Now(),
	}
	if err := db.Create(task).Error; err != nil {
		r.log.WithContext(ctx).Errorw("msg", "CreateTask failed", "err", err, "name", name)
		return 0, err
	}
	return task.ID, nil
}

// BatchCreateRecords 批量写入账单记录，并回填自增 ID
func (r *ReconRepo) BatchCreateRecords(ctx context.Context, taskID int64, recs []domain.Record) ([]domain.Record, error) {
	if len(recs) == 0 {
		return recs, nil
	}
	db, err := r.data.DB(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	rows := make([]*model.ReconRecord, 0, len(recs))
	for i := range recs {
		var tt *time.Time
		if !recs[i].TradeTime.IsZero() {
			t := recs[i].TradeTime
			tt = &t
		}
		rows = append(rows, &model.ReconRecord{
			TaskID:    taskID,
			Source:    recs[i].Source,
			TradeNo:   recs[i].TradeNo,
			BizNo:     recs[i].BizNo,
			Amount:    recs[i].Amount,
			Fee:       recs[i].Fee,
			TradeTime: tt,
			Status:    recs[i].Status,
			Remark:    recs[i].Remark,
			CreatedAt: now,
		})
	}

	if err := db.CreateInBatches(rows, 500).Error; err != nil {
		r.log.WithContext(ctx).Errorw("msg", "BatchCreateRecords failed", "err", err, "count", len(rows))
		return nil, err
	}

	for i := range rows {
		recs[i].ID = rows[i].ID
		recs[i].TaskID = taskID
	}
	return recs, nil
}

// BatchCreateDiffs 批量写入差异
func (r *ReconRepo) BatchCreateDiffs(ctx context.Context, taskID int64, diffs []domain.Diff) error {
	if len(diffs) == 0 {
		return nil
	}
	db, err := r.data.DB(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	rows := make([]*model.ReconDiff, 0, len(diffs))
	for _, d := range diffs {
		var chID, sysID *int64
		if d.ChannelRecordID > 0 {
			v := d.ChannelRecordID
			chID = &v
		}
		if d.SystemRecordID > 0 {
			v := d.SystemRecordID
			sysID = &v
		}
		rows = append(rows, &model.ReconDiff{
			TaskID:          taskID,
			ChannelRecordID: chID,
			SystemRecordID:  sysID,
			DiffType:        string(d.Type),
			AmountDiff:      d.AmountDiff,
			Evidence:        "[]",
			State:           "PENDING",
			MatchRule:       string(d.MatchRule),
			CreatedAt:       now,
			UpdatedAt:       now,
		})
	}

	if err := db.CreateInBatches(rows, 500).Error; err != nil {
		r.log.WithContext(ctx).Errorw("msg", "BatchCreateDiffs failed", "err", err, "count", len(rows))
		return err
	}
	return nil
}

// UpdateTaskStats 更新任务统计并置为已匹配
func (r *ReconRepo) UpdateTaskStats(ctx context.Context, taskID int64, total, diffCount int32, diffAmount int64) error {
	db, err := r.data.DB(ctx)
	if err != nil {
		return err
	}
	return db.Model(&model.ReconTask{}).Where("id = ?", taskID).Updates(map[string]any{
		"total_count": total,
		"diff_count":  diffCount,
		"diff_amount": diffAmount,
		"status":      "MATCHED",
		"finish_time": time.Now(),
	}).Error
}

// GetTask 查询任务
func (r *ReconRepo) GetTask(ctx context.Context, id int64) (*model.ReconTask, error) {
	db, err := r.data.DB(ctx)
	if err != nil {
		return nil, err
	}
	var task model.ReconTask
	err = db.Where("id = ?", id).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// ListDiffs 分页查询差异
func (r *ReconRepo) ListDiffs(ctx context.Context, taskID int64, category, state string, page, pageSize int32) ([]*model.ReconDiff, int64, error) {
	db, err := r.data.DB(ctx)
	if err != nil {
		return nil, 0, err
	}

	base := db.Model(&model.ReconDiff{}).Where("task_id = ?", taskID)
	if category != "" {
		base = base.Where("category = ?", category)
	}
	if state != "" {
		base = base.Where("state = ?", state)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.ReconDiff
	offset := (page - 1) * pageSize
	if err := base.Order("id asc").Offset(int(offset)).Limit(int(pageSize)).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetDiff 查询单条差异
func (r *ReconRepo) GetDiff(ctx context.Context, id int64) (*model.ReconDiff, error) {
	db, err := r.data.DB(ctx)
	if err != nil {
		return nil, err
	}
	var d model.ReconDiff
	err = db.Where("id = ?", id).First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// GetRecordsByIDs 按 ID 批量查询记录
func (r *ReconRepo) GetRecordsByIDs(ctx context.Context, ids []int64) (map[int64]*model.ReconRecord, error) {
	out := make(map[int64]*model.ReconRecord, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	db, err := r.data.DB(ctx)
	if err != nil {
		return nil, err
	}
	var list []*model.ReconRecord
	if err := db.Where("id in ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, rec := range list {
		out[rec.ID] = rec
	}
	return out, nil
}

// SaveClassification 回写 AI 归因结果
func (r *ReconRepo) SaveClassification(ctx context.Context, d *Classification) error {
	db, err := r.data.DB(ctx)
	if err != nil {
		return err
	}

	evidence, err := json.Marshal(d.Evidence)
	if err != nil {
		return err
	}

	return db.Model(&model.ReconDiff{}).Where("id = ?", d.DiffID).Updates(map[string]any{
		"category":       d.Category,
		"confidence":     d.Confidence,
		"reason":         d.Reason,
		"evidence":       string(evidence),
		"suggestion":     d.Suggestion,
		"fallback":       d.Fallback,
		"ai_model":       d.AIModel,
		"prompt_version": d.PromptVersion,
		"state":          "CLASSIFIED",
		"updated_at":     time.Now(),
	}).Error
}

// Classification 归因结果
type Classification struct {
	DiffID        int64
	Category      string
	Confidence    float64
	Reason        string
	Evidence      []string
	Suggestion    string
	Fallback      bool
	AIModel       string
	PromptVersion string
}

// ConfirmDiff 确认或纠正差异
func (r *ReconRepo) ConfirmDiff(ctx context.Context, diffID int64, state, aiCategory, humanCategory, operator string) error {
	db, err := r.data.DB(ctx)
	if err != nil {
		return err
	}

	if err := db.Model(&model.ReconDiff{}).Where("id = ?", diffID).Updates(map[string]any{
		"state":      state,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return err
	}

	if humanCategory == "" {
		return nil
	}

	fb := &model.DiffFeedback{
		DiffID:        diffID,
		AICategory:    aiCategory,
		HumanCategory: humanCategory,
		Operator:      operator,
		CreatedAt:     time.Now(),
	}
	return db.Create(fb).Error
}

// CategoryAgg 分类聚合
type CategoryAgg struct {
	Category string
	Count    int64
	Amount   int64
}

// SummaryByCategory 按分类聚合差异
func (r *ReconRepo) SummaryByCategory(ctx context.Context, taskID int64) ([]CategoryAgg, error) {
	db, err := r.data.DB(ctx)
	if err != nil {
		return nil, err
	}
	var aggs []CategoryAgg
	err = db.Model(&model.ReconDiff{}).
		Select("category, COUNT(*) AS count, COALESCE(SUM(amount_diff), 0) AS amount").
		Where("task_id = ?", taskID).
		Group("category").
		Scan(&aggs).Error
	return aggs, err
}

// CountByState 按状态统计差异数量
func (r *ReconRepo) CountByState(ctx context.Context, taskID int64) (map[string]int64, error) {
	db, err := r.data.DB(ctx)
	if err != nil {
		return nil, err
	}
	var rows []struct {
		State string
		Count int64
	}
	err = db.Model(&model.ReconDiff{}).
		Select("state, COUNT(*) AS count").
		Where("task_id = ?", taskID).
		Group("state").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.State] = row.Count
	}
	return out, nil
}
