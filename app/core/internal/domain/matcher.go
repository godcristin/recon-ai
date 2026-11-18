package domain

import "time"

const fuzzyWindow = 24 * time.Hour

func Match(channel, system []Record) []Diff {
	c := newRecordIndex(channel)
	s := newRecordIndex(system)

	diffs := make([]Diff, 0, len(channel)+len(system))

	for i := range c.recs {
		j := s.findExact(c.recs[i])
		if j < 0 {
			continue
		}
		c.used[i] = true
		s.used[j] = true
		if d := compare(c.recs[i], s.recs[j], MatchExact); d != nil {
			diffs = append(diffs, *d)
		}
	}

	for i := range c.recs {
		if c.used[i] {
			continue
		}
		j := s.findFuzzy(c.recs[i])
		if j < 0 {
			continue
		}
		c.used[i] = true
		s.used[j] = true
		if d := compare(c.recs[i], s.recs[j], MatchFuzzy); d != nil {
			diffs = append(diffs, *d)
		}
	}

	for i := range c.recs {
		if c.used[i] {
			continue
		}
		diffs = append(diffs, Diff{
			ChannelRecordID: c.recs[i].ID,
			Type:            DiffChannelOnly,
			AmountDiff:      c.recs[i].Amount,
			MatchRule:       MatchNone,
		})
	}

	for j := range s.recs {
		if s.used[j] {
			continue
		}
		diffs = append(diffs, Diff{
			SystemRecordID: s.recs[j].ID,
			Type:           DiffSystemOnly,
			AmountDiff:     -s.recs[j].Amount,
			MatchRule:      MatchNone,
		})
	}

	return diffs
}

func MatchOne(channel Record, system []Record) (int, MatchRule) {
	s := newRecordIndex(system)
	if j := s.findExact(channel); j >= 0 {
		return j, MatchExact
	}
	if j := s.findFuzzy(channel); j >= 0 {
		return j, MatchFuzzy
	}
	return -1, MatchNone
}

func compare(channel, system Record, rule MatchRule) *Diff {
	if channel.Amount == system.Amount {
		return nil
	}
	return &Diff{
		ChannelRecordID: channel.ID,
		SystemRecordID:  system.ID,
		Type:            DiffAmountMismatch,
		AmountDiff:      channel.Amount - system.Amount,
		MatchRule:       rule,
	}
}

type recordIndex struct {
	recs    []Record
	used    []bool
	byBiz   map[string][]int
	byTrade map[string][]int
}

func newRecordIndex(recs []Record) *recordIndex {
	ix := &recordIndex{
		recs:    recs,
		used:    make([]bool, len(recs)),
		byBiz:   make(map[string][]int, len(recs)),
		byTrade: make(map[string][]int, len(recs)),
	}
	for i := range recs {
		if recs[i].BizNo != "" {
			ix.byBiz[recs[i].BizNo] = append(ix.byBiz[recs[i].BizNo], i)
		}
		if recs[i].TradeNo != "" {
			ix.byTrade[recs[i].TradeNo] = append(ix.byTrade[recs[i].TradeNo], i)
		}
	}
	return ix
}

func (ix *recordIndex) firstUnused(candidates []int) int {
	for _, i := range candidates {
		if !ix.used[i] {
			return i
		}
	}
	return -1
}

func (ix *recordIndex) findExact(r Record) int {
	if r.BizNo != "" {
		if i := ix.firstUnused(ix.byBiz[r.BizNo]); i >= 0 {
			return i
		}
	}
	if r.TradeNo != "" {
		if i := ix.firstUnused(ix.byTrade[r.TradeNo]); i >= 0 {
			return i
		}
	}
	return -1
}

func (ix *recordIndex) findFuzzy(r Record) int {
	if r.TradeTime.IsZero() {
		return -1
	}

	candidate := -1
	for i := range ix.recs {
		if ix.used[i] {
			continue
		}
		s := ix.recs[i]
		if s.TradeTime.IsZero() || s.Amount != r.Amount {
			continue
		}
		d := s.TradeTime.Sub(r.TradeTime)
		if d < 0 {
			d = -d
		}
		if d > fuzzyWindow {
			continue
		}
		if candidate >= 0 {
			return -1
		}
		candidate = i
	}
	return candidate
}
