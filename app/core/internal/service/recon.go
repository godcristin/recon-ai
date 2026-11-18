package service

import (
	"context"
	"errors"

	commV1 "recon-ai/api/comm/v1"
	reconV1 "recon-ai/api/recon/v1"
	"recon-ai/app/core/internal/biz"
	"recon-ai/app/core/internal/data"
	"recon-ai/app/core/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

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

func (s *ReconService) CreateTask(ctx context.Context, req *reconV1.CreateTaskReq) (*reconV1.CreateTaskRsp, error) {
	if err := req.Validate(); err != nil {
		return nil, commV1.ErrorParamErr("%s", err.Error())
	}

	res, err := s.uc.CreateTask(ctx, req.Name, req.ChannelCsv, req.SystemCsv)
	if err != nil {
		return nil, s.translate(ctx, "CreateTask", err)
	}

	parseErrs := make([]*reconV1.ParseError, 0, len(res.ParseErrors))
	for _, e := range res.ParseErrors {
		parseErrs = append(parseErrs, &reconV1.ParseError{Line: int32(e.Line), Message: e.Message})
	}

	return &reconV1.CreateTaskRsp{
		TaskId:      res.TaskID,
		TotalCount:  res.Total,
		DiffCount:   res.DiffCount,
		ParseErrors: parseErrs,
	}, nil
}

func (s *ReconService) GetTask(ctx context.Context, req *reconV1.GetTaskReq) (*reconV1.Task, error) {
	if err := req.Validate(); err != nil {
		return nil, commV1.ErrorParamErr("%s", err.Error())
	}
	task, err := s.uc.GetTask(ctx, req.Id)
	if err != nil {
		return nil, s.translate(ctx, "GetTask", err)
	}
	if task == nil {
		return nil, commV1.ErrorTaskNotFound("任务不存在")
	}
	return toProtoTask(task), nil
}

func (s *ReconService) ListDiffs(ctx context.Context, req *reconV1.ListDiffsReq) (*reconV1.ListDiffsRsp, error) {
	if err := req.Validate(); err != nil {
		return nil, commV1.ErrorParamErr("%s", err.Error())
	}

	list, total, err := s.uc.ListDiffs(ctx, req.TaskId, req.Category, req.State, req.Page, req.PageSize)
	if err != nil {
		return nil, s.translate(ctx, "ListDiffs", err)
	}

	rsp := &reconV1.ListDiffsRsp{Total: int32(total), List: make([]*reconV1.Diff, 0, len(list))}
	for _, d := range list {
		rsp.List = append(rsp.List, toProtoDiff(d))
	}

	if task, err := s.uc.GetTask(ctx, req.TaskId); err == nil && task != nil {
		rsp.TotalAmount = task.DiffAmount
	}

	return rsp, nil
}

func (s *ReconService) GetDiff(ctx context.Context, req *reconV1.GetDiffReq) (*reconV1.DiffDetail, error) {
	if err := req.Validate(); err != nil {
		return nil, commV1.ErrorParamErr("%s", err.Error())
	}

	d, channel, system, err := s.uc.GetDiff(ctx, req.Id)
	if err != nil {
		return nil, s.translate(ctx, "GetDiff", err)
	}
	if d == nil {
		return nil, commV1.ErrorDiffNotFound("差异不存在")
	}

	matchRule := "none"
	if d.ChannelRecordID != nil && d.SystemRecordID != nil {
		matchRule = d.MatchRule
	}

	return &reconV1.DiffDetail{
		Diff:          toProtoDiff(d),
		Channel:       toProtoRecord(channel),
		System:        toProtoRecord(system),
		MatchRule:     matchRule,
		TimeDiffHours: timeDiffHours(channel, system),
	}, nil
}

func timeDiffHours(channel, system *model.ReconRecord) int64 {
	if channel == nil || system == nil || channel.TradeTime == nil || system.TradeTime == nil {
		return 0
	}
	d := system.TradeTime.Sub(*channel.TradeTime)
	if d < 0 {
		d = -d
	}
	return int64(d.Hours())
}

func (s *ReconService) SaveClassification(ctx context.Context, req *reconV1.SaveClassificationReq) (*reconV1.SaveClassificationRsp, error) {
	if err := req.Validate(); err != nil {
		return nil, commV1.ErrorParamErr("%s", err.Error())
	}

	err := s.uc.SaveClassification(ctx, &data.Classification{
		DiffID:        req.DiffId,
		Category:      req.Category,
		Confidence:    req.Confidence,
		Reason:        req.Reason,
		Evidence:      req.Evidence,
		Suggestion:    req.Suggestion,
		Fallback:      req.Fallback,
		AIModel:       req.AiModel,
		PromptVersion: req.PromptVersion,
	})
	if err != nil {
		return nil, s.translate(ctx, "SaveClassification", err)
	}
	return &reconV1.SaveClassificationRsp{Ok: true}, nil
}

func (s *ReconService) ConfirmDiff(ctx context.Context, req *reconV1.ConfirmDiffReq) (*reconV1.ConfirmDiffRsp, error) {
	if err := req.Validate(); err != nil {
		return nil, commV1.ErrorParamErr("%s", err.Error())
	}

	state, err := s.uc.ConfirmDiff(ctx, req.DiffId, req.Action, req.HumanCategory, req.Operator)
	if err != nil {
		return nil, s.translate(ctx, "ConfirmDiff", err)
	}
	return &reconV1.ConfirmDiffRsp{Ok: true, State: state}, nil
}

func (s *ReconService) GetTaskSummary(ctx context.Context, req *reconV1.GetTaskSummaryReq) (*reconV1.TaskSummaryRsp, error) {
	if err := req.Validate(); err != nil {
		return nil, commV1.ErrorParamErr("%s", err.Error())
	}

	task, aggs, acceptRate, err := s.uc.Summary(ctx, req.TaskId)
	if err != nil {
		return nil, s.translate(ctx, "GetTaskSummary", err)
	}
	if task == nil {
		return nil, commV1.ErrorTaskNotFound("任务不存在")
	}

	stats := make([]*reconV1.CategoryStat, 0, len(aggs))
	for _, a := range aggs {
		stats = append(stats, &reconV1.CategoryStat{
			Category: a.Category,
			Count:    int32(a.Count),
			Amount:   a.Amount,
		})
	}

	return &reconV1.TaskSummaryRsp{
		Total:        task.TotalCount,
		DiffCount:    task.DiffCount,
		DiffAmount:   task.DiffAmount,
		Stats:        stats,
		AiAcceptRate: acceptRate,
	}, nil
}

func (s *ReconService) translate(ctx context.Context, op string, err error) error {
	switch {
	case errors.Is(err, biz.ErrEmptyFile):
		return commV1.ErrorFileEmpty("账单文件为空")
	case errors.Is(err, biz.ErrParseFailed):
		return commV1.ErrorFileFormatError("账单文件格式错误")
	case errors.Is(err, biz.ErrInvalidCategory):
		return commV1.ErrorParamErr("非法的差异分类")
	case errors.Is(err, biz.ErrInvalidAction):
		return commV1.ErrorParamErr("非法的操作类型，仅支持 confirm / correct")
	case errors.Is(err, biz.ErrDiffNotFound):
		return commV1.ErrorDiffNotFound("差异不存在")
	case errors.Is(err, biz.ErrDiffSettled):
		return commV1.ErrorDiffAlreadySettled("该差异已处理")
	default:
		s.log.WithContext(ctx).Errorw("msg", "recon service failed", "op", op, "err", err)
		return commV1.ErrorDbInternalError("处理失败，请稍后重试")
	}
}
