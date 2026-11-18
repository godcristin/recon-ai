package prompts

import (
	_ "embed"
	"strings"
)

//go:embed classify.v2.md
var classifyV2 string

// 历史版本保留在目录中，便于回溯与对比评测：
//
//go:embed classify.v1.md
var _ string

// ClassifyVersion 当前 prompt 版本（写入每条归因结果，便于回溯与评测）
const ClassifyVersion = "classify.v2"

// Classify 返回 system prompt 与 user 模板（模板中 {{context}} 为占位符）
func Classify() (system, userTemplate string) {
	const sep = "## User 模板"
	parts := strings.SplitN(classifyV2, sep, 2)

	system = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts[0]), "## System"))
	if len(parts) == 2 {
		userTemplate = strings.TrimSpace(parts[1])
	}
	return system, userTemplate
}
