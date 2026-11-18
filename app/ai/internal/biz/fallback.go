package biz

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
