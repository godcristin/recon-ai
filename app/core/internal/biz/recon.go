package biz

import (
	"bytes"
	"context"
	"fmt"

	"recon-ai/app/core/internal/data"
	"recon-ai/app/core/internal/data/model"
	"recon-ai/app/core/internal/domain"
	"recon-ai/app/core/internal/parser"

	"github.com/go-kratos/kratos/v2/log"
)

// 差异分类枚举（与 AI prompt 中的标签保持一致）
const (
	CategoryCrossDay        = "cross_day"
	CategoryFeeDiff         = "fee_diff"
	CategoryRefundUnsync    = "refund_unsync"
	CategoryMissingInSystem = "missing_in_system"
	CategoryMissingInChannel = "missing_in_channel"
	CategoryDuplicate       = "duplicate"
)

// ValidCategory 判断是否为合法分类
func ValidCategory(c string) bool {
	switch c {
	case CategoryCrossDay, CategoryFeeDiff, CategoryRefundUnsync,
		CategoryMissingInSystem, CategoryMissingInChannel, CategoryDuplicate:
		return true
	}
	return false
}

// ReconUseCase 对账用例：解析、匹配、差异、确认
type ReconUseCase struct {
	repo *data.ReconRepo
	tm   data.Transaction
	log  *log.Helper
}

func NewReconUseCase(repo *data.ReconRepo, tm data.Transaction, logger log.Logger) *ReconUseCase {
	return &ReconUseCase{
		repo: repo,
		tm:   tm,
		log:  log.NewHelper(log.With(logger, "module", "biz/recon")),
	}
}

// CreateTaskResult 建任务的结果
type CreateTaskResult struct {
	TaskID      int64
	Total       int32
	DiffCount   int32
	ParseErrors []parser.Error
}

// CreateTask 解析两份 CSV → 落库 → 三级匹配 → 生成差异 → 更新统计
func (uc *ReconUseCase) CreateTask(ctx context.Context, name string, channelCSV, systemCSV []byte) (*CreateTaskResult, error) {
	if len(channelCSV) == 0 || len(systemCSV) == 0 {
		return nil, ErrEmptyFile
	}

	channelRecs, channelErrs, err := parser.ParseChannel(bytes.NewReader(channelCSV))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
	}
	systemRecs, systemErrs, err := parser.ParseSystem(bytes.NewReader(systemCSV))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
	}

	parseErrors := make([]parser.Error, 0, len(channelErrs)+len(systemErrs))
	parseErrors = append(parseErrors, channelErrs...)
	parseErrors = append(parseErrors, systemErrs...)

	if len(channelRecs) == 0 || len(systemRecs) == 0 {
		return nil, ErrEmptyFile
	}

	var (
		taskID    int64
		diffCount int32
		diffSum   int64
		total     int32
	)

	err = uc.tm.InTx(ctx, func(txCtx context.Context) error {
		id, txErr := uc.repo.CreateTask(txCtx, name)
		if txErr != nil {
			return txErr
		}
		taskID = id

		all := make([]domain.Record, 0, len(channelRecs)+len(systemRecs))
		all = append(all, channelRecs...)
		all = append(all, systemRecs...)

		all, txErr = uc.repo.BatchCreateRecords(txCtx, id, all)
		if txErr != nil {
			return txErr
		}

		channelRecs = all[:len(channelRecs)]
		systemRecs = all[len(channelRecs):]

		diffs := domain.Match(channelRecs, systemRecs)
		if txErr = uc.repo.BatchCreateDiffs(txCtx, id, diffs); txErr != nil {
			return txErr
		}

		for _, d := range diffs {
			diffSum += d.AmountDiff
		}
		diffCount = int32(len(diffs))
		total = int32(len(all))

		return uc.repo.UpdateTaskStats(txCtx, id, total, diffCount, diffSum)
	})
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "ReconUseCase.CreateTask failed", "err", err, "name", name)
		return nil, err
	}

	uc.log.WithContext(ctx).Infow("msg", "ReconUseCase.CreateTask",
		"taskId", taskID, "total", total, "diffCount", diffCount, "parseErrors", len(parseErrors))

	return &CreateTaskResult{
		TaskID:      taskID,
		Total:       total,
		DiffCount:   diffCount,
		ParseErrors: parseErrors,
	}, nil
}

// GetTask 查询任务
func (uc *ReconUseCase) GetTask(ctx context.Context, id int64) (*model.ReconTask, error) {
	return uc.repo.GetTask(ctx, id)
}

// ListDiffs 分页查询差异
func (uc *ReconUseCase) ListDiffs(ctx context.Context, taskID int64, category, state string, page, pageSize int32) ([]*model.ReconDiff, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	if category != "" && !ValidCategory(category) {
		return nil, 0, ErrInvalidCategory
	}
	return uc.repo.ListDiffs(ctx, taskID, category, state, page, pageSize)
}

// GetDiff 查询差异详情（含两侧原始记录）
func (uc *ReconUseCase) GetDiff(ctx context.Context, id int64) (*model.ReconDiff, *model.ReconRecord, *model.ReconRecord, error) {
	d, err := uc.repo.GetDiff(ctx, id)
	if err != nil || d == nil {
		return nil, nil, nil, err
	}

	ids := make([]int64, 0, 2)
	if d.ChannelRecordID != nil {
		ids = append(ids, *d.ChannelRecordID)
	}
	if d.SystemRecordID != nil {
		ids = append(ids, *d.SystemRecordID)
	}

	recs, err := uc.repo.GetRecordsByIDs(ctx, ids)
	if err != nil {
		return nil, nil, nil, err
	}

	var channel, system *model.ReconRecord
	if d.ChannelRecordID != nil {
		channel = recs[*d.ChannelRecordID]
	}
	if d.SystemRecordID != nil {
		system = recs[*d.SystemRecordID]
	}
	return d, channel, system, nil
}

// SaveClassification 回写 AI 归因
func (uc *ReconUseCase) SaveClassification(ctx context.Context, c *data.Classification) error {
	if !ValidCategory(c.Category) {
		return ErrInvalidCategory
	}
	return uc.repo.SaveClassification(ctx, c)
}

// ConfirmDiff 确认或纠正
func (uc *ReconUseCase) ConfirmDiff(ctx context.Context, diffID int64, action, humanCategory, operator string) (string, error) {
	d, err := uc.repo.GetDiff(ctx, diffID)
	if err != nil {
		return "", err
	}
	if d == nil {
		return "", ErrDiffNotFound
	}
	if d.State == "CONFIRMED" || d.State == "CORRECTED" {
		return "", ErrDiffSettled
	}

	var (
		state string
		human string
	)
	switch action {
	case "confirm":
		state = "CONFIRMED"
	case "correct":
		if !ValidCategory(humanCategory) {
			return "", ErrInvalidCategory
		}
		state = "CORRECTED"
		human = humanCategory
	default:
		return "", ErrInvalidAction
	}

	if err := uc.repo.ConfirmDiff(ctx, diffID, state, d.Category, human, operator); err != nil {
		return "", err
	}
	return state, nil
}

// Summary 任务汇总
func (uc *ReconUseCase) Summary(ctx context.Context, taskID int64) (*model.ReconTask, []data.CategoryAgg, float64, error) {
	task, err := uc.repo.GetTask(ctx, taskID)
	if err != nil || task == nil {
		return nil, nil, 0, err
	}

	aggs, err := uc.repo.SummaryByCategory(ctx, taskID)
	if err != nil {
		return nil, nil, 0, err
	}

	states, err := uc.repo.CountByState(ctx, taskID)
	if err != nil {
		return nil, nil, 0, err
	}
	confirmed := states["CONFIRMED"]
	corrected := states["CORRECTED"]
	acceptRate := 0.0
	if confirmed+corrected > 0 {
		acceptRate = float64(confirmed) / float64(confirmed+corrected)
	}

	return task, aggs, acceptRate, nil
}
