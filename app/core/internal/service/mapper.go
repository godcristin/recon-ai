package service

import (
	"encoding/json"
	"time"

	reconV1 "recon-ai/api/recon/v1"
	"recon-ai/app/core/internal/data/model"
)

func toProtoRecord(r *model.ReconRecord) *reconV1.Record {
	if r == nil {
		return nil
	}
	tradeTime := ""
	if r.TradeTime != nil {
		tradeTime = r.TradeTime.Format(time.RFC3339)
	}
	return &reconV1.Record{
		Id:        r.ID,
		Source:    r.Source,
		TradeNo:   r.TradeNo,
		BizNo:     r.BizNo,
		Amount:    r.Amount,
		Fee:       r.Fee,
		TradeTime: tradeTime,
		Status:    r.Status,
		Remark:    r.Remark,
	}
}

func toProtoDiff(d *model.ReconDiff) *reconV1.Diff {
	if d == nil {
		return nil
	}
	evidence := make([]string, 0, 4)
	if d.Evidence != "" {
		if err := json.Unmarshal([]byte(d.Evidence), &evidence); err != nil {
			evidence = evidence[:0]
		}
	}
	return &reconV1.Diff{
		Id:            d.ID,
		TaskId:        d.TaskID,
		DiffType:      d.DiffType,
		AmountDiff:    d.AmountDiff,
		Category:      d.Category,
		Confidence:    d.Confidence,
		Reason:        d.Reason,
		Evidence:      evidence,
		Suggestion:    d.Suggestion,
		State:         d.State,
		Fallback:      d.Fallback,
		AiModel:       d.AIModel,
		PromptVersion: d.PromptVersion,
		CreateTime:    d.CreatedAt.Unix(),
	}
}

func toProtoTask(t *model.ReconTask) *reconV1.Task {
	if t == nil {
		return nil
	}
	var finish int64
	if t.FinishTime != nil {
		finish = t.FinishTime.Unix()
	}
	return &reconV1.Task{
		Id:         t.ID,
		Name:       t.Name,
		Status:     t.Status,
		TotalCount: t.TotalCount,
		DiffCount:  t.DiffCount,
		DiffAmount: t.DiffAmount,
		CreateTime: t.CreatedAt.Unix(),
		FinishTime: finish,
	}
}
