// Command smoke 是 core 服务的端到端冒烟客户端：
// 建任务（上传两份内联 CSV）→ 查差异列表 → 查汇总。
// 用法：go run ./cmd/smoke -addr 127.0.0.1:9000
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

const channelCSV = `trade_no,out_trade_no,amount,fee,trade_time,status,remark
T1,P1,680.00,2.72,2026-09-01 10:00:00,SUCCESS,学费
T2,P2,50.00,0.20,2026-09-01 11:00:00,SUCCESS,餐费
T3,P3,100.00,0.40,2026-09-01 12:00:00,SUCCESS,校服费
`

const systemCSV = `trade_no,pay_order_no,amount,fee,trade_time,status,remark
CH1,P1,68000,272,2026-09-01 10:00:00,SUCCESS,学费
CH3,P3,9900,0,2026-09-01 12:00:00,SUCCESS,校服费
CH4,P4,3000,0,2026-09-01 13:00:00,SUCCESS,活动费
`

func main() {
	addr := flag.String("addr", "127.0.0.1:9000", "recon-core gRPC address")
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

	client := reconV1.NewReconClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	created, err := client.CreateTask(ctx, &reconV1.CreateTaskReq{
		Name:       "smoke-" + time.Now().Format("150405"),
		ChannelCsv: []byte(channelCSV),
		SystemCsv:  []byte(systemCSV),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "CreateTask:", err)
		os.Exit(1)
	}
	fmt.Printf("CreateTask  -> task_id=%d total=%d diff=%d parse_errors=%d\n",
		created.TaskId, created.TotalCount, created.DiffCount, len(created.ParseErrors))

	list, err := client.ListDiffs(ctx, &reconV1.ListDiffsReq{TaskId: created.TaskId, Page: 1, PageSize: 20})
	if err != nil {
		fmt.Fprintln(os.Stderr, "ListDiffs:", err)
		os.Exit(1)
	}
	fmt.Printf("ListDiffs   -> total=%d amount=%d\n", list.Total, list.TotalAmount)
	for _, d := range list.List {
		fmt.Printf("   diff id=%d type=%s amount_diff=%d state=%s\n", d.Id, d.DiffType, d.AmountDiff, d.State)
	}

	summary, err := client.GetTaskSummary(ctx, &reconV1.GetTaskSummaryReq{TaskId: created.TaskId})
	if err != nil {
		fmt.Fprintln(os.Stderr, "GetTaskSummary:", err)
		os.Exit(1)
	}
	fmt.Printf("Summary     -> diff_count=%d diff_amount=%d accept_rate=%.2f\n",
		summary.DiffCount, summary.DiffAmount, summary.AiAcceptRate)
}
