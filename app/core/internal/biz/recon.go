package biz

import (
	"context"

	"recon-ai/app/core/internal/data"

	"github.com/go-kratos/kratos/v2/log"
)

// ReconUseCase 对账用例：编排任务、匹配与差异
type ReconUseCase struct {
	repo *data.ReconRepo
	log  *log.Helper
}

func NewReconUseCase(repo *data.ReconRepo, logger log.Logger) *ReconUseCase {
	return &ReconUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "biz/recon")),
	}
}

// CreateTask 阶段 2：只落任务记录；CSV 解析与匹配在阶段 3 实现
func (uc *ReconUseCase) CreateTask(ctx context.Context, name string, channelCSV, systemCSV []byte) (int64, int32, int32, error) {
	if len(channelCSV) == 0 || len(systemCSV) == 0 {
		return 0, 0, 0, ErrEmptyFile
	}

	id, err := uc.repo.CreateTask(ctx, name)
	if err != nil {
		return 0, 0, 0, err
	}

	uc.log.WithContext(ctx).Infow("msg", "ReconUseCase.CreateTask", "taskId", id, "name", name)
	return id, 0, 0, nil
}
