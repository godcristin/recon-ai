// Command smoke 是 core 服务的冒烟客户端：调用 CreateTask 并打印返回。
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rsp, err := reconV1.NewReconClient(conn).CreateTask(ctx, &reconV1.CreateTaskReq{
		Name:       "smoke-" + time.Now().Format("150405"),
		ChannelCsv: []byte("trade_no,amount\n"),
		SystemCsv:  []byte("pay_order_no,amount\n"),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "CreateTask:", err)
		os.Exit(1)
	}

	fmt.Printf("CreateTask ok: task_id=%d total=%d diff=%d\n", rsp.TaskId, rsp.TotalCount, rsp.DiffCount)
}
