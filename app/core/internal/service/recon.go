package service

import (
	"context"
	"errors"

	commV1 "recon-ai/api/comm/v1"
	reconV1 "recon-ai/api/recon/v1"
	"recon-ai/app/core/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// ReconService 只做协议转换：校验、调用 usecase、翻译错误
type ReconService struct {
	reconV1.UnimplementedReconServer

	uc  *biz.ReconUseCase
	log *log.Helper
}

func NewReconService(uc *biz.ReconUseCase, logger log.Logger) *ReconService {
	return &ReconService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "service/recon")),
	}
}

// CreateTask 阶段 2：建任务记录并返回；解析与匹配在阶段 3 补齐
func (s *ReconService) CreateTask(ctx context.Context, req *reconV1.CreateTaskReq) (*reconV1.CreateTaskRsp, error) {
	if err := req.Validate(); err != nil {
		return nil, commV1.ErrorParamErr(err.Error())
	}

	taskID, total, diff, err := s.uc.CreateTask(ctx, req.Name, req.ChannelCsv, req.SystemCsv)
	if err != nil {
		if errors.Is(err, biz.ErrEmptyFile) {
			return nil, commV1.ErrorFileEmpty("账单文件为空")
		}
		s.log.WithContext(ctx).Errorw("msg", "ReconService.CreateTask failed", "err", err)
		return nil, commV1.ErrorDbInternalError("创建对账任务失败")
	}

	return &reconV1.CreateTaskRsp{TaskId: taskID, TotalCount: total, DiffCount: diff}, nil
}
