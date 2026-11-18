package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func yuan(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

func main() {
	n := flag.Int("n", 1000, "生成的订单对数（每对约 2 条记录）")
	out := flag.String("out", "testdata", "输出目录")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir:", err)
		os.Exit(1)
	}

	channelPath := filepath.Join(*out, "channel.csv")
	systemPath := filepath.Join(*out, "system.csv")

	chFile, err := os.Create(channelPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create channel:", err)
		os.Exit(1)
	}
	defer chFile.Close()

	sysFile, err := os.Create(systemPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create system:", err)
		os.Exit(1)
	}
	defer sysFile.Close()

	chw := bufio.NewWriter(chFile)
	sysw := bufio.NewWriter(sysFile)

	fmt.Fprintln(chw, "trade_no,out_trade_no,amount,fee,trade_time,status,remark")
	fmt.Fprintln(sysw, "trade_no,pay_order_no,amount,fee,trade_time,status,remark")

	base := time.Date(2026, 9, 1, 9, 0, 0, 0, time.Local)

	var (
		channelOnly int
		systemOnly  int
		mismatch    int
	)

	for i := 1; i <= *n; i++ {
		biz := fmt.Sprintf("P%06d", i)
		trade := fmt.Sprintf("T%06d", i)
		cents := int64(1000 + (i*37)%500000)
		at := base.Add(time.Duration(i) * time.Minute)
		ts := at.Format("2006-01-02 15:04:05")
		fee := cents * 6 / 1000 // 0.6%

		switch i % 50 {
		case 10:
			fmt.Fprintf(chw, "%s,%s,%s,%s,%s,SUCCESS,渠道单边\n", trade, biz, yuan(cents), yuan(fee), ts)
			channelOnly++
		case 20:
			fmt.Fprintf(sysw, "%s,%s,%d,%d,%s,SUCCESS,系统单边\n", trade, biz, cents, fee, ts)
			systemOnly++
		case 30:
			fmt.Fprintf(chw, "%s,%s,%s,%s,%s,SUCCESS,金额差异\n", trade, biz, yuan(cents), yuan(fee), ts)
			fmt.Fprintf(sysw, "%s,%s,%d,%d,%s,SUCCESS,金额差异\n", trade, biz, cents-100, fee, ts)
			mismatch++
		default:
			fmt.Fprintf(chw, "%s,%s,%s,%s,%s,SUCCESS,\n", trade, biz, yuan(cents), yuan(fee), ts)
			fmt.Fprintf(sysw, "%s,%s,%d,%d,%s,SUCCESS,\n", trade, biz, cents, fee, ts)
		}
	}

	if err := chw.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, "flush channel:", err)
		os.Exit(1)
	}
	if err := sysw.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, "flush system:", err)
		os.Exit(1)
	}

	fmt.Printf("generated %d pairs -> %s / %s\n", *n, channelPath, systemPath)
	fmt.Printf("expected diffs: %d (channel_only=%d system_only=%d amount_mismatch=%d)\n",
		channelOnly+systemOnly+mismatch, channelOnly, systemOnly, mismatch)
}
