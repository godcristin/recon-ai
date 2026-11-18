package biz

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"recon-ai/app/ai/internal/conf"
	"recon-ai/app/ai/internal/llm"
	"recon-ai/app/ai/internal/prompts"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewCache, NewClassifyUseCase)

const maxAttempts = 2

type ClassifyUseCase struct {
	client *llm.Client
	cache  *Cache
	log    *log.Helper
	enable bool
}

func NewClassifyUseCase(client *llm.Client, cache *Cache, c *conf.Bootstrap, logger log.Logger) *ClassifyUseCase {
	return &ClassifyUseCase{
		client: client,
		cache:  cache,
		log:    log.NewHelper(log.With(logger, "module", "biz/classify")),
		enable: c.Llm.Enable,
	}
}

func (uc *ClassifyUseCase) Classify(ctx context.Context, in Input) Output {
	contextStr := contextJSON(in)
	key := hashKey(contextStr)

	if out, ok := uc.cache.Get(key); ok {
		out.Cached = true
		return out
	}

	if !uc.enable || !uc.client.Enabled() {
		return Fallback(in, "模型未启用，已使用规则兜底")
	}

	system, tmpl := prompts.Classify()
	user := strings.ReplaceAll(tmpl, "{{context}}", contextStr)
	allowed := allowedEvidence(in)

	start := time.Now()
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		res, err := uc.client.ChatJSON(ctx, system, user)
		if err != nil {
			lastErr = err
			continue
		}

		parsed, err := validate(res.Content, allowed)
		if err != nil {
			lastErr = err
			continue
		}

		out := Output{
			Category:      parsed.Category,
			Confidence:    parsed.Confidence,
			Reason:        parsed.Reason,
			Evidence:      parsed.Evidence,
			Suggestion:    parsed.Suggestion,
			Model:         uc.client.Model(),
			PromptVersion: prompts.ClassifyVersion,
			TokensIn:      res.TokensIn,
			TokensOut:     res.TokensOut,
			LatencyMS:     time.Since(start).Milliseconds(),
		}
		uc.cache.Set(key, out)

		uc.log.WithContext(ctx).Infow("msg", "ClassifyUseCase.Classify",
			"category", out.Category, "confidence", out.Confidence,
			"tokensIn", out.TokensIn, "tokensOut", out.TokensOut,
			"latencyMs", out.LatencyMS, "promptVersion", out.PromptVersion)
		return out
	}

	out := Fallback(in, "模型调用失败，已使用规则兜底")
	out.LatencyMS = time.Since(start).Milliseconds()
	uc.log.WithContext(ctx).Warnw("msg", "ClassifyUseCase.Classify fallback",
		"err", lastErr, "diffType", in.DiffType, "latencyMs", out.LatencyMS)
	return out
}

func contextJSON(in Input) string {
	m := map[string]any{
		"diff_type":   in.DiffType,
		"amount_diff": in.AmountDiff,
	}
	if in.TimeDiffHours != 0 {
		m["time_diff_hours"] = in.TimeDiffHours
	}
	if in.Channel != nil {
		m["channel"] = recordJSON(in.Channel)
	}
	if in.System != nil {
		m["system"] = recordJSON(in.System)
	}

	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func recordJSON(r *Record) map[string]any {
	m := map[string]any{
		"amount": r.Amount,
	}
	if r.TradeNo != "" {
		m["trade_no"] = r.TradeNo
	}
	if r.BizNo != "" {
		m["biz_no"] = r.BizNo
	}
	if r.Fee != 0 {
		m["fee"] = r.Fee
	}
	if r.TradeTime != "" {
		m["trade_time"] = r.TradeTime
	}
	if r.Status != "" {
		m["status"] = r.Status
	}
	if r.Remark != "" {
		m["remark"] = r.Remark
	}
	return m
}

func hashKey(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
