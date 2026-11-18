package prompts

import (
	_ "embed"
	"strings"
)

//go:embed classify.v2.md
var classifyV2 string

//go:embed classify.v1.md
var _ string

const ClassifyVersion = "classify.v2"

func Classify() (system, userTemplate string) {
	const sep = "## User 模板"
	parts := strings.SplitN(classifyV2, sep, 2)

	system = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts[0]), "## System"))
	if len(parts) == 2 {
		userTemplate = strings.TrimSpace(parts[1])
	}
	return system, userTemplate
}
