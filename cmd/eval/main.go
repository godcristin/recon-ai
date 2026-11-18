// Command eval 对 AI 归因服务跑离线评测：准确率 / 分类别 / 混淆矩阵 / 延迟与 token。
//
// 用法：
//
//	go run ./cmd/eval -file ai/eval/data/diffs.jsonl -addr 127.0.0.1:9001 -out ai/eval/report.md
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	aiV1 "recon-ai/api/ai/v1"
	reconV1 "recon-ai/api/recon/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type sample struct {
	ID            string `json:"id"`
	DiffType      string `json:"diff_type"`
	Channel       *rec   `json:"channel"`
	System        *rec   `json:"system"`
	AmountDiff    int64  `json:"amount_diff"`
	TimeDiffHours int64  `json:"time_diff_hours"`
	Expected      string `json:"expected"`
}

type rec struct {
	TradeNo   string `json:"trade_no"`
	BizNo     string `json:"biz_no"`
	Amount    int64  `json:"amount"`
	Fee       int64  `json:"fee"`
	TradeTime string `json:"trade_time"`
	Status    string `json:"status"`
	Remark    string `json:"remark"`
}

type outcome struct {
	sample    sample
	predicted string
	confidence float64
	fallback  bool
	cached    bool
	latencyMS int64
	tokensIn  int32
	tokensOut int32
	err       error
}

var categoryOrder = []string{
	"cross_day", "fee_diff", "refund_unsync",
	"missing_in_system", "missing_in_channel", "duplicate",
}

func main() {
	file := flag.String("file", "ai/eval/data/diffs.jsonl", "评测数据文件")
	addr := flag.String("addr", "127.0.0.1:9001", "AI 服务 gRPC 地址")
	out := flag.String("out", "ai/eval/report.md", "报告输出路径")
	verbose := flag.Bool("v", false, "打印每条样本的结果")
	flag.Parse()

	samples, err := loadSamples(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load samples:", err)
		os.Exit(1)
	}
	if len(samples) == 0 {
		fmt.Fprintln(os.Stderr, "no samples")
		os.Exit(1)
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "dial:", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			fmt.Fprintln(os.Stderr, "close:", closeErr)
		}
	}()

	client := aiV1.NewAiClient(conn)
	ctx := context.Background()

	results := make([]outcome, 0, len(samples))
	for _, s := range samples {
		req := &aiV1.ClassifyDiffReq{
			DiffType:      s.DiffType,
			Channel:       toRecord(s.Channel),
			System:        toRecord(s.System),
			AmountDiff:    s.AmountDiff,
			TimeDiffHours: s.TimeDiffHours,
		}
		callCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		rsp, err := client.ClassifyDiff(callCtx, req)
		cancel()

		o := outcome{sample: s}
		if err != nil {
			o.err = err
		} else {
			o.predicted = rsp.Category
			o.confidence = rsp.Confidence
			o.fallback = rsp.Fallback
			o.cached = rsp.Cached
			o.latencyMS = rsp.LatencyMs
			o.tokensIn = rsp.TokensIn
			o.tokensOut = rsp.TokensOut
		}
		results = append(results, o)

		if *verbose {
			mark := "OK "
			if o.predicted != s.Expected || o.err != nil {
				mark = "MISS"
			}
			fmt.Printf("%s %-4s expected=%-18s got=%-18s conf=%.2f fb=%v\n",
				mark, s.ID, s.Expected, o.predicted, o.confidence, o.fallback)
		}
	}

	printSummary(results)
	if err := writeReport(*out, results); err != nil {
		fmt.Fprintln(os.Stderr, "write report:", err)
		os.Exit(1)
	}
	fmt.Printf("\nreport written to %s\n", *out)
}

func loadSamples(path string) ([]sample, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			fmt.Fprintln(os.Stderr, "close:", closeErr)
		}
	}()

	var out []sample
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		var s sample
		if err := json.Unmarshal([]byte(text), &s); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		out = append(out, s)
	}
	return out, scanner.Err()
}

func toRecord(r *rec) *reconV1.Record {
	if r == nil {
		return nil
	}
	return &reconV1.Record{
		TradeNo:   r.TradeNo,
		BizNo:     r.BizNo,
		Amount:    r.Amount,
		Fee:       r.Fee,
		TradeTime: r.TradeTime,
		Status:    r.Status,
		Remark:    r.Remark,
	}
}

func printSummary(results []outcome) {
	correct, fallback, cached, failed := 0, 0, 0, 0
	var latencySum int64
	var tokIn, tokOut int32

	for _, r := range results {
		if r.err != nil {
			failed++
			continue
		}
		if r.predicted == r.sample.Expected {
			correct++
		}
		if r.fallback {
			fallback++
		}
		if r.cached {
			cached++
		}
		latencySum += r.latencyMS
		tokIn += r.tokensIn
		tokOut += r.tokensOut
	}

	total := len(results)
	success := total - failed
	fmt.Printf("\n===== 汇总 =====\n")
	fmt.Printf("样本数      : %d\n", total)
	fmt.Printf("调用失败    : %d\n", failed)
	fmt.Printf("准确率      : %d/%d = %.2f%%\n", correct, success, ratio(correct, success)*100)
	fmt.Printf("兜底次数    : %d\n", fallback)
	fmt.Printf("命中缓存    : %d\n", cached)
	if success > 0 {
		fmt.Printf("平均延迟    : %d ms\n", latencySum/int64(success))
	}
	fmt.Printf("token 合计  : in=%d out=%d\n", tokIn, tokOut)

	fmt.Printf("\n===== 分类别 =====\n")
	byClass := map[string][2]int{}
	for _, r := range results {
		if r.err != nil {
			continue
		}
		v := byClass[r.sample.Expected]
		v[1]++
		if r.predicted == r.sample.Expected {
			v[0]++
		}
		byClass[r.sample.Expected] = v
	}
	for _, c := range categoryOrder {
		v, ok := byClass[c]
		if !ok {
			continue
		}
		fmt.Printf("%-18s %d/%d  %.0f%%\n", c, v[0], v[1], ratio(v[0], v[1])*100)
	}
}

func writeReport(path string, results []outcome) error {
	correct, fallback, failed := 0, 0, 0
	var latencySum int64
	var tokIn, tokOut int32
	byClass := map[string][2]int{}
	confusion := map[string]map[string]int{}

	for _, r := range results {
		if r.err != nil {
			failed++
			continue
		}
		if r.predicted == r.sample.Expected {
			correct++
		}
		if r.fallback {
			fallback++
		}
		latencySum += r.latencyMS
		tokIn += r.tokensIn
		tokOut += r.tokensOut

		v := byClass[r.sample.Expected]
		v[1]++
		if r.predicted == r.sample.Expected {
			v[0]++
		}
		byClass[r.sample.Expected] = v

		if confusion[r.sample.Expected] == nil {
			confusion[r.sample.Expected] = map[string]int{}
		}
		confusion[r.sample.Expected][r.predicted]++
	}

	success := len(results) - failed

	var b strings.Builder
	b.WriteString("# AI 归因评测报告\n\n")
	b.WriteString(fmt.Sprintf("- 生成时间：%s\n", time.Now().Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("- 样本数：%d（调用失败 %d）\n", len(results), failed))
	b.WriteString(fmt.Sprintf("- 准确率：**%d/%d = %.1f%%**\n", correct, success, ratio(correct, success)*100))
	b.WriteString(fmt.Sprintf("- 规则兜底次数：%d\n", fallback))
	if success > 0 {
		b.WriteString(fmt.Sprintf("- 平均延迟：%d ms\n", latencySum/int64(success)))
	}
	b.WriteString(fmt.Sprintf("- token 消耗：in=%d / out=%d\n\n", tokIn, tokOut))

	b.WriteString("## 分类别准确率\n\n| 分类 | 正确/总数 | 准确率 |\n|---|---|---|\n")
	for _, c := range categoryOrder {
		v, ok := byClass[c]
		if !ok {
			continue
		}
		b.WriteString(fmt.Sprintf("| %s | %d/%d | %.0f%% |\n", c, v[0], v[1], ratio(v[0], v[1])*100))
	}

	b.WriteString("\n## 混淆矩阵（行=期望，列=预测）\n\n")
	labels := append([]string{}, categoryOrder...)
	labels = append(labels, "(空)")
	b.WriteString("| 期望 \\ 预测 | " + strings.Join(labels, " | ") + " |\n")
	b.WriteString("|---|" + strings.Repeat("---|", len(labels)) + "\n")
	for _, exp := range categoryOrder {
		row := confusion[exp]
		if row == nil {
			continue
		}
		cells := make([]string, 0, len(labels))
		for _, pred := range labels {
			p := pred
			if p == "(空)" {
				p = ""
			}
			cells = append(cells, fmt.Sprintf("%d", row[p]))
		}
		b.WriteString("| " + exp + " | " + strings.Join(cells, " | ") + " |\n")
	}

	b.WriteString("\n## 错例\n\n")
	var misses []outcome
	for _, r := range results {
		if r.err != nil || r.predicted != r.sample.Expected {
			misses = append(misses, r)
		}
	}
	if len(misses) == 0 {
		b.WriteString("无\n")
	} else {
		sort.Slice(misses, func(i, j int) bool { return misses[i].sample.ID < misses[j].sample.ID })
		b.WriteString("| ID | 期望 | 预测 | 置信度 | 兜底 |\n|---|---|---|---|---|\n")
		for _, m := range misses {
			pred := m.predicted
			if m.err != nil {
				pred = "ERROR"
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %.2f | %v |\n",
				m.sample.ID, m.sample.Expected, pred, m.confidence, m.fallback))
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func ratio(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}
