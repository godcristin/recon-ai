package parser

import (
	"strings"
	"testing"
)

func TestParseYuanToFen(t *testing.T) {
	tests := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"680", 68000, false},
		{"680.00", 68000, false},
		{"680.5", 68050, false},
		{"680.05", 68005, false},
		{"0.01", 1, false},
		{"-1.50", -150, false},
		{"", 0, true},
		{"abc", 0, true},
		{"1.234", 0, true},
	}

	for _, tc := range tests {
		got, err := ParseYuanToFen(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseYuanToFen(%q) want error, got %d", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseYuanToFen(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseYuanToFen(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestParseChannel(t *testing.T) {
	const data = `trade_no,out_trade_no,amount,fee,trade_time,status,remark
T1,P1,680.00,2.72,2026-09-01 10:23:11,SUCCESS,学费
T2,P2,100.50,0.40,2026/09/01 11:00:00,SUCCESS,
`

	records, errs, err := ParseChannel(strings.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected row errors: %+v", errs)
	}
	if len(records) != 2 {
		t.Fatalf("record count = %d, want 2", len(records))
	}

	if records[0].Amount != 68000 {
		t.Errorf("amount = %d, want 68000", records[0].Amount)
	}
	if records[0].Fee != 272 {
		t.Errorf("fee = %d, want 272", records[0].Fee)
	}
	if records[0].BizNo != "P1" || records[0].TradeNo != "T1" {
		t.Errorf("biz/trade = %q/%q", records[0].BizNo, records[0].TradeNo)
	}
	if records[1].Amount != 10050 {
		t.Errorf("amount = %d, want 10050", records[1].Amount)
	}
}

func TestParseChannelRowError(t *testing.T) {
	const data = `trade_no,out_trade_no,amount,trade_time,status
T1,P1,680.00,2026-09-01 10:00:00,SUCCESS
T2,P2,not-a-number,2026-09-01 10:00:00,SUCCESS
T3,P3,10.00,2026-09-01 10:00:00,SUCCESS
`

	records, errs, err := ParseChannel(strings.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("record count = %d, want 2 (bad row skipped)", len(records))
	}
	if len(errs) != 1 {
		t.Fatalf("row errors = %d, want 1", len(errs))
	}
	if errs[0].Line != 3 {
		t.Errorf("error line = %d, want 3", errs[0].Line)
	}
}

func TestParseSystem(t *testing.T) {
	const data = `trade_no,pay_order_no,amount,trade_time,status
CH1,PAY001,68000,2026-09-01 10:23:11,SUCCESS
`

	records, errs, err := ParseSystem(strings.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected row errors: %+v", errs)
	}
	if records[0].Amount != 68000 {
		t.Errorf("system amount should stay in fen, got %d", records[0].Amount)
	}
	if records[0].BizNo != "PAY001" {
		t.Errorf("biz no = %q, want PAY001", records[0].BizNo)
	}
}

func TestParseMissingColumn(t *testing.T) {
	const data = `trade_no,amount,trade_time,status
T1,680.00,2026-09-01 10:00:00,SUCCESS
`
	if _, _, err := ParseChannel(strings.NewReader(data)); err == nil {
		t.Fatal("want error for missing out_trade_no column, got nil")
	}
}

func TestParseTime(t *testing.T) {
	if _, err := ParseTime("2026-09-01 10:00:00"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if _, err := ParseTime("2026-09-01T10:00:00+08:00"); err != nil {
		t.Errorf("unexpected error for RFC3339: %v", err)
	}
	if _, err := ParseTime("09/01/2026"); err == nil {
		t.Error("want error for unsupported format")
	}
}
