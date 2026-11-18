package biz

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// modelOutput 模型返回的原始结构
type modelOutput struct {
	Category   string   `json:"category"`
	Confidence float64  `json:"confidence"`
	Reason     string   `json:"reason"`
	Evidence   []string `json:"evidence"`
	Suggestion string   `json:"suggestion"`
}

// validate 解析模型输出并做四项校验：
// 1) JSON 合法 2) category 在枚举内 3) confidence 在 [0,1] 4) evidence 必须在白名单内
func validate(raw string, allowedEvidence map[string]bool) (*modelOutput, error) {
	var out modelOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	if !validCategory(out.Category) {
		return nil, fmt.Errorf("invalid category: %q", out.Category)
	}

	switch {
	case out.Confidence < 0:
		out.Confidence = 0
	case out.Confidence > 1:
		out.Confidence = 1
	}

	if !validSuggestion(out.Suggestion) {
		out.Suggestion = SuggestionInvestigate
	}

	// evidence 白名单：只保留输入中真实存在的字段；一旦出现编造，降低置信度
	cleaned := make([]string, 0, len(out.Evidence))
	hadFabrication := false
	for _, e := range out.Evidence {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if allowedEvidence[e] {
			cleaned = append(cleaned, e)
		} else {
			hadFabrication = true
		}
	}
	if hadFabrication {
		out.Confidence *= 0.8
	}
	out.Evidence = cleaned

	out.Reason = truncateRunes(strings.TrimSpace(out.Reason), 200)
	if out.Reason == "" {
		out.Reason = "模型未给出理由"
	}

	return &out, nil
}

// allowedEvidence 根据输入实际出现的字段生成白名单
func allowedEvidence(in Input) map[string]bool {
	m := map[string]bool{}
	if in.Channel != nil {
		m["trade_no"] = in.Channel.TradeNo != ""
		m["biz_no"] = in.Channel.BizNo != ""
		m["amount"] = true
		m["fee"] = in.Channel.Fee != 0
		m["trade_time"] = in.Channel.TradeTime != ""
		m["status"] = in.Channel.Status != ""
		m["remark"] = in.Channel.Remark != ""
	}
	if in.System != nil {
		m["trade_no"] = m["trade_no"] || in.System.TradeNo != ""
		m["biz_no"] = m["biz_no"] || in.System.BizNo != ""
		m["amount"] = true
		m["fee"] = m["fee"] || in.System.Fee != 0
		m["trade_time"] = m["trade_time"] || in.System.TradeTime != ""
		m["status"] = m["status"] || in.System.Status != ""
		m["remark"] = m["remark"] || in.System.Remark != ""
	}
	m["amount_diff"] = true
	if in.TimeDiffHours != 0 {
		m["time_diff_hours"] = true
	}

	allowed := make(map[string]bool, len(m))
	for k, ok := range m {
		if ok {
			allowed[k] = true
		}
	}
	return allowed
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return string(r[:n]) + "..."
}
