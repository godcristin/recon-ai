package biz

import (
	"context"

	aiV1 "recon-ai/api/ai/v1"
	reconV1 "recon-ai/api/recon/v1"
	"recon-ai/app/bff/internal/data"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewReconUseCase)

type ReconUseCase struct {
	core reconV1.ReconClient
	ai   aiV1.AiClient
	log  *log.Helper
}

func NewReconUseCase(clients *data.Clients, logger log.Logger) *ReconUseCase {
	return &ReconUseCase{
		core: clients.Core,
		ai:   clients.AI,
		log:  log.NewHelper(log.With(logger, "module", "biz/recon")),
	}
}

func (uc *ReconUseCase) CreateTask(ctx context.Context, name string, channelCSV, systemCSV []byte) (*reconV1.CreateTaskRsp, error) {
	return uc.core.CreateTask(ctx, &reconV1.CreateTaskReq{
		Name:       name,
		ChannelCsv: channelCSV,
		SystemCsv:  systemCSV,
	})
}

func (uc *ReconUseCase) GetTask(ctx context.Context, id int64) (*reconV1.Task, error) {
	return uc.core.GetTask(ctx, &reconV1.GetTaskReq{Id: id})
}

func (uc *ReconUseCase) ListDiffs(ctx context.Context, req *reconV1.ListDiffsReq) (*reconV1.ListDiffsRsp, error) {
	return uc.core.ListDiffs(ctx, req)
}

func (uc *ReconUseCase) GetDiff(ctx context.Context, id int64) (*reconV1.DiffDetail, error) {
	detail, err := uc.core.GetDiff(ctx, &reconV1.GetDiffReq{Id: id})
	if err != nil {
		return nil, err
	}
	if detail.Diff == nil || detail.Diff.Category != "" {
		return detail, nil
	}

	rsp, err := uc.ai.ClassifyDiff(ctx, &aiV1.ClassifyDiffReq{
		DiffType:      detail.Diff.DiffType,
		Channel:       detail.Channel,
		System:        detail.System,
		AmountDiff:    detail.Diff.AmountDiff,
		TimeDiffHours: detail.TimeDiffHours,
	})
	if err != nil {
		uc.log.WithContext(ctx).Warnw("msg", "ai classify failed, keep pending", "diffId", id, "err", err)
		return detail, nil
	}

	if _, err := uc.core.SaveClassification(ctx, &reconV1.SaveClassificationReq{
		DiffId:        id,
		Category:      rsp.Category,
		Confidence:    rsp.Confidence,
		Reason:        rsp.Reason,
		Evidence:      rsp.Evidence,
		Suggestion:    rsp.Suggestion,
		Fallback:      rsp.Fallback,
		AiModel:       rsp.Model,
		PromptVersion: rsp.PromptVersion,
	}); err != nil {
		uc.log.WithContext(ctx).Warnw("msg", "save classification failed", "diffId", id, "err", err)
		return detail, nil
	}

	detail.Diff.Category = rsp.Category
	detail.Diff.Confidence = rsp.Confidence
	detail.Diff.Reason = rsp.Reason
	detail.Diff.Evidence = rsp.Evidence
	detail.Diff.Suggestion = rsp.Suggestion
	detail.Diff.Fallback = rsp.Fallback
	detail.Diff.AiModel = rsp.Model
	detail.Diff.PromptVersion = rsp.PromptVersion
	detail.Diff.State = "CLASSIFIED"

	return detail, nil
}

func (uc *ReconUseCase) ConfirmDiff(ctx context.Context, req *reconV1.ConfirmDiffReq) (*reconV1.ConfirmDiffRsp, error) {
	return uc.core.ConfirmDiff(ctx, req)
}

func (uc *ReconUseCase) Summary(ctx context.Context, taskID int64) (*reconV1.TaskSummaryRsp, error) {
	return uc.core.GetTaskSummary(ctx, &reconV1.GetTaskSummaryReq{TaskId: taskID})
}
