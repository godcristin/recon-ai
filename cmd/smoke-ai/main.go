package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	aiV1 "recon-ai/api/ai/v1"
	reconV1 "recon-ai/api/recon/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9001", "recon-ai gRPC address")
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "dial:", err)
		os.Exit(1)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "close:", err)
		}
	}()

	client := aiV1.NewAiClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cases := []struct {
		name string
		req  *aiV1.ClassifyDiffReq
	}{
		{
			name: "手续费差异",
			req: &aiV1.ClassifyDiffReq{
				DiffType:   "amount_mismatch",
				AmountDiff: -272,
				Channel: &reconV1.Record{
					TradeNo: "4200002841", BizNo: "P20260901001",
					Amount: 67728, Fee: 272,
					TradeTime: "2026-09-01T10:23:11+08:00", Status: "SUCCESS", Remark: "学费",
				},
				System: &reconV1.Record{
					BizNo: "P20260901001", Amount: 68000,
					TradeTime: "2026-09-01T10:23:11+08:00", Status: "SUCCESS", Remark: "学费",
				},
			},
		},
		{
			name: "渠道单边（系统查无此单）",
			req: &aiV1.ClassifyDiffReq{
				DiffType:   "channel_only",
				AmountDiff: 5000,
				Channel: &reconV1.Record{
					TradeNo: "4200002842", BizNo: "P20260901002",
					Amount: 5000, Fee: 20,
					TradeTime: "2026-09-01T11:00:00+08:00", Status: "SUCCESS", Remark: "餐费",
				},
			},
		},
		{
			name: "跨日入账",
			req: &aiV1.ClassifyDiffReq{
				DiffType:      "system_only",
				AmountDiff:    -10000,
				TimeDiffHours: 6,
				System: &reconV1.Record{
					TradeNo: "CH3", BizNo: "P20260831099",
					Amount:    10000,
					TradeTime: "2026-09-01T00:05:00+08:00", Status: "SUCCESS", Remark: "校服费",
				},
			},
		},
	}

	for _, tc := range cases {
		start := time.Now()
		rsp, err := client.ClassifyDiff(ctx, tc.req)
		cost := time.Since(start).Round(time.Millisecond)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] error: %v\n", tc.name, err)
			continue
		}
		fmt.Printf("[%s] %s\n", tc.name, cost)
		fmt.Printf("  category=%s confidence=%.2f suggestion=%s fallback=%v cached=%v\n",
			rsp.Category, rsp.Confidence, rsp.Suggestion, rsp.Fallback, rsp.Cached)
		fmt.Printf("  reason=%s\n", rsp.Reason)
		fmt.Printf("  evidence=%v model=%s tokens=%d/%d latency=%dms\n\n",
			rsp.Evidence, rsp.Model, rsp.TokensIn, rsp.TokensOut, rsp.LatencyMs)
	}
}
