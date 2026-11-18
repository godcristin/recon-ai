package biz

// Fallback 规则兜底：模型不可用或输出非法时使用。
// 这里只做保守判断，置信度普遍不高，提示人工核查。
func Fallback(in Input, reason string) Output {
	category := CategoryMissingInSystem
	confidence := 0.5
	suggestion := SuggestionInvestigate

	switch in.DiffType {
	case "channel_only":
		category = CategoryMissingInSystem
		confidence = 0.6
	case "system_only":
		category = CategoryMissingInChannel
		confidence = 0.6
	case "amount_mismatch":
		category = CategoryFeeDiff
		confidence = 0.55
	default:
		category = CategoryMissingInSystem
		confidence = 0.4
	}

	return Output{
		Category:      category,
		Confidence:    confidence,
		Reason:        reason,
		Evidence:      []string{"amount_diff"},
		Suggestion:    suggestion,
		Fallback:      true,
		PromptVersion: "fallback.v1",
	}
}
