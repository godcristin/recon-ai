package service

import (
	"context"

	aiV1 "recon-ai/api/ai/v1"
	reconV1 "recon-ai/api/recon/v1"
	"recon-ai/app/ai/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type AiService struct {
	aiV1.UnimplementedAiServer

	uc  *biz.ClassifyUseCase
	log *log.Helper
}

func NewAiService(uc *biz.ClassifyUseCase, logger log.Logger) *AiService {
	return &AiService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "service/ai")),
	}
}

func (s *AiService) ClassifyDiff(ctx context.Context, req *aiV1.ClassifyDiffReq) (*aiV1.ClassifyDiffRsp, error) {
	in := biz.Input{
		DiffType:      req.DiffType,
		Channel:       toRecord(req.Channel),
		System:        toRecord(req.System),
		AmountDiff:    req.AmountDiff,
		TimeDiffHours: req.TimeDiffHours,
	}

	out := s.uc.Classify(ctx, in)

	return &aiV1.ClassifyDiffRsp{
		Category:      out.Category,
		Confidence:    out.Confidence,
		Reason:        out.Reason,
		Evidence:      out.Evidence,
		Suggestion:    out.Suggestion,
		Fallback:      out.Fallback,
		Model:         out.Model,
		PromptVersion: out.PromptVersion,
		TokensIn:      out.TokensIn,
		TokensOut:     out.TokensOut,
		LatencyMs:     out.LatencyMS,
		Cached:        out.Cached,
	}, nil
}

func toRecord(r *reconV1.Record) *biz.Record {
	if r == nil {
		return nil
	}
	return &biz.Record{
		TradeNo:   r.TradeNo,
		BizNo:     r.BizNo,
		Amount:    r.Amount,
		Fee:       r.Fee,
		TradeTime: r.TradeTime,
		Status:    r.Status,
		Remark:    r.Remark,
	}
}
