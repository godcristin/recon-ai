// Command smoke 是 core 服务的端到端冒烟客户端：
// 建任务 → 查差异 → 确认一条 → 纠正一条 → 查汇总。
// 用法：
//
//	go run ./cmd/smoke                       # 用内置小样本
//	go run ./cmd/smoke -channel testdata/channel.csv -system testdata/system.csv
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	reconV1 "recon-ai/api/recon/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultChannelCSV = `trade_no,out_trade_no,amount,fee,trade_time,status,remark
T1,P1,680.00,2.72,2026-09-01 10:00:00,SUCCESS,学费
T2,P2,50.00,0.20,2026-09-01 11:00:00,SUCCESS,餐费
T3,P3,100.00,0.40,2026-09-01 12:00:00,SUCCESS,校服费
`

const defaultSystemCSV = `trade_no,pay_order_no,amount,fee,trade_time,status,remark
CH1,P1,68000,272,2026-09-01 10:00:00,SUCCESS,学费
CH3,P3,9900,0,2026-09-01 12:00:00,SUCCESS,校服费
CH4,P4,3000,0,2026-09-01 13:00:00,SUCCESS,活动费
`

func main() {
	addr := flag.String("addr", "127.0.0.1:9000", "recon-core gRPC address")
	channelFile := flag.String("channel", "", "渠道账单 CSV 路径（为空则用内置样本）")
	systemFile := flag.String("system", "", "系统订单 CSV 路径（为空则用内置样本）")
	noConfirm := flag.Bool("no-confirm", false, "跳过确认/纠正步骤")
	flag.Parse()

	channelCSV := []byte(defaultChannelCSV)
	systemCSV := []byte(defaultSystemCSV)
	var err error
	if *channelFile != "" {
		if channelCSV, err = os.ReadFile(*channelFile); err != nil {
			fmt.Fprintln(os.Stderr, "read channel:", err)
			os.Exit(1)
		}
	}
	if *systemFile != "" {
		if systemCSV, err = os.ReadFile(*systemFile); err != nil {
			fmt.Fprintln(os.Stderr, "read system:", err)
			os.Exit(1)
		}
	}

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

	client := reconV1.NewReconClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()
	created, err := client.CreateTask(ctx, &reconV1.CreateTaskReq{
		Name:       "smoke-" + time.Now().Format("150405"),
		ChannelCsv: channelCSV,
		SystemCsv:  systemCSV,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "CreateTask:", err)
		os.Exit(1)
	}
	createCost := time.Since(start)

	fmt.Printf("CreateTask  -> task_id=%d total=%d diff=%d parse_errors=%d  (%s)\n",
		created.TaskId, created.TotalCount, created.DiffCount, len(created.ParseErrors), createCost.Round(time.Millisecond))

	list, err := client.ListDiffs(ctx, &reconV1.ListDiffsReq{TaskId: created.TaskId, Page: 1, PageSize: 20})
	if err != nil {
		fmt.Fprintln(os.Stderr, "ListDiffs:", err)
		os.Exit(1)
	}
	fmt.Printf("ListDiffs   -> total=%d amount=%d\n", list.Total, list.TotalAmount)
	show := len(list.List)
	if show > 5 {
		show = 5
	}
	for _, d := range list.List[:show] {
		fmt.Printf("   diff id=%d type=%s amount_diff=%d state=%s\n", d.Id, d.DiffType, d.AmountDiff, d.State)
	}

	if !*noConfirm && len(list.List) >= 2 {
		first, second := list.List[0], list.List[1]

		if _, err := client.ConfirmDiff(ctx, &reconV1.ConfirmDiffReq{
			DiffId: first.Id, Action: "confirm", Operator: "smoke",
		}); err != nil {
			fmt.Fprintln(os.Stderr, "ConfirmDiff:", err)
			os.Exit(1)
		}

		humanCategory := "cross_day"
		switch second.DiffType {
		case "channel_only":
			humanCategory = "missing_in_system"
		case "system_only":
			humanCategory = "missing_in_channel"
		}
		if _, err := client.ConfirmDiff(ctx, &reconV1.ConfirmDiffReq{
			DiffId: second.Id, Action: "correct", HumanCategory: humanCategory, Operator: "smoke",
		}); err != nil {
			fmt.Fprintln(os.Stderr, "CorrectDiff:", err)
			os.Exit(1)
		}
		fmt.Printf("ConfirmDiff -> id=%d confirmed, id=%d corrected to %s\n", first.Id, second.Id, humanCategory)
	}

	summary, err := client.GetTaskSummary(ctx, &reconV1.GetTaskSummaryReq{TaskId: created.TaskId})
	if err != nil {
		fmt.Fprintln(os.Stderr, "GetTaskSummary:", err)
		os.Exit(1)
	}
	fmt.Printf("Summary     -> diff_count=%d diff_amount=%d accept_rate=%.2f\n",
		summary.DiffCount, summary.DiffAmount, summary.AiAcceptRate)
}
