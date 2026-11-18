package domain

import (
	"testing"
	"time"
)

func mkRec(id int64, source int32, biz, trade string, amount int64, at time.Time) Record {
	return Record{
		ID:        id,
		Source:    source,
		BizNo:     biz,
		TradeNo:   trade,
		Amount:    amount,
		TradeTime: at,
	}
}

func TestMatch(t *testing.T) {
	base := time.Date(2026, 9, 1, 10, 0, 0, 0, time.Local)

	tests := []struct {
		name      string
		channel   []Record
		system    []Record
		wantTypes []DiffType
		wantAmt   []int64
	}{
		{
			name:    "精确匹配且金额一致：无差异",
			channel: []Record{mkRec(1, SourceChannel, "P1", "T1", 68000, base)},
			system:  []Record{mkRec(2, SourceSystem, "P1", "", 68000, base)},
		},
		{
			name:      "精确匹配但金额不一致：amount_mismatch",
			channel:   []Record{mkRec(1, SourceChannel, "P1", "T1", 67728, base)},
			system:    []Record{mkRec(2, SourceSystem, "P1", "", 68000, base)},
			wantTypes: []DiffType{DiffAmountMismatch},
			wantAmt:   []int64{-272},
		},
		{
			name:      "渠道单边：channel_only",
			channel:   []Record{mkRec(1, SourceChannel, "P9", "T9", 5000, base)},
			system:    nil,
			wantTypes: []DiffType{DiffChannelOnly},
			wantAmt:   []int64{5000},
		},
		{
			name:      "系统单边：system_only",
			channel:   nil,
			system:    []Record{mkRec(2, SourceSystem, "P8", "", 7000, base)},
			wantTypes: []DiffType{DiffSystemOnly},
			wantAmt:   []int64{-7000},
		},
		{
			name:    "单号不同但金额相同且时间接近：模糊匹配成功",
			channel: []Record{mkRec(1, SourceChannel, "", "T100", 68000, base)},
			system:  []Record{mkRec(2, SourceSystem, "P100", "", 68000, base.Add(2*time.Hour))},
		},
		{
			name:      "模糊匹配出现两个候选：放弃匹配，全部判差异",
			channel:   []Record{mkRec(1, SourceChannel, "", "T100", 68000, base)},
			system:    []Record{mkRec(2, SourceSystem, "P201", "", 68000, base.Add(time.Hour)), mkRec(3, SourceSystem, "P202", "", 68000, base.Add(2*time.Hour))},
			wantTypes: []DiffType{DiffChannelOnly, DiffSystemOnly, DiffSystemOnly},
			wantAmt:   []int64{68000, -68000, -68000},
		},
		{
			name:      "超出 24 小时窗口：不做模糊匹配",
			channel:   []Record{mkRec(1, SourceChannel, "", "T100", 68000, base)},
			system:    []Record{mkRec(2, SourceSystem, "P100", "", 68000, base.Add(48*time.Hour))},
			wantTypes: []DiffType{DiffChannelOnly, DiffSystemOnly},
			wantAmt:   []int64{68000, -68000},
		},
		{
			name: "重复单号按一配一匹配",
			channel: []Record{
				mkRec(1, SourceChannel, "P1", "T1", 1000, base),
				mkRec(3, SourceChannel, "P1", "T1", 1000, base),
			},
			system: []Record{
				mkRec(2, SourceSystem, "P1", "", 1000, base),
				mkRec(4, SourceSystem, "P1", "", 1000, base),
			},
		},
		{
			name:      "双方单号都为空且金额时间都不同：不误配",
			channel:   []Record{mkRec(1, SourceChannel, "", "", 1000, base)},
			system:    []Record{mkRec(2, SourceSystem, "", "", 2000, base.Add(time.Hour))},
			wantTypes: []DiffType{DiffChannelOnly, DiffSystemOnly},
			wantAmt:   []int64{1000, -2000},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Match(tc.channel, tc.system)

			if len(got) != len(tc.wantTypes) {
				t.Fatalf("diff count = %d, want %d, got = %+v", len(got), len(tc.wantTypes), got)
			}
			for i := range got {
				if got[i].Type != tc.wantTypes[i] {
					t.Errorf("diff[%d].Type = %s, want %s", i, got[i].Type, tc.wantTypes[i])
				}
				if got[i].AmountDiff != tc.wantAmt[i] {
					t.Errorf("diff[%d].AmountDiff = %d, want %d", i, got[i].AmountDiff, tc.wantAmt[i])
				}
			}
		})
	}
}

func TestMatchOne(t *testing.T) {
	base := time.Date(2026, 9, 1, 10, 0, 0, 0, time.Local)

	system := []Record{
		mkRec(10, SourceSystem, "P1", "", 1000, base),
		mkRec(11, SourceSystem, "P2", "", 2000, base),
	}

	idx, rule := MatchOne(mkRec(1, SourceChannel, "P2", "T2", 2000, base), system)
	if idx != 1 || rule != MatchExact {
		t.Errorf("MatchOne exact = (%d,%s), want (1,exact)", idx, rule)
	}

	idx, rule = MatchOne(mkRec(2, SourceChannel, "", "T3", 2000, base.Add(time.Hour)), system)
	if idx != 1 || rule != MatchFuzzy {
		t.Errorf("MatchOne fuzzy = (%d,%s), want (1,fuzzy)", idx, rule)
	}

	idx, rule = MatchOne(mkRec(3, SourceChannel, "P9", "T9", 9999, base), system)
	if idx != -1 || rule != MatchNone {
		t.Errorf("MatchOne none = (%d,%s), want (-1,none)", idx, rule)
	}
}
