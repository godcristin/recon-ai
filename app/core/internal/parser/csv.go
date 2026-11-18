package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"recon-ai/app/core/internal/domain"
)

type Error struct {
	Line    int
	Message string
}

var timeLayouts = []string{
	"2006-01-02 15:04:05",
	"2006/01/02 15:04:05",
	time.RFC3339,
}

func ParseTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range timeLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %q", s)
}

func ParseYuanToFen(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty")
	}

	neg := false
	switch s[0] {
	case '-':
		neg = true
		s = s[1:]
	case '+':
		s = s[1:]
	}

	intPart, fracPart, _ := strings.Cut(s, ".")
	if intPart == "" {
		intPart = "0"
	}
	yuan, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %q", s)
	}
	if len(fracPart) > 2 {
		return 0, fmt.Errorf("more than 2 decimal places: %q", s)
	}
	fracPart += strings.Repeat("0", 2-len(fracPart))
	cents, err := strconv.ParseInt(fracPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid decimal part: %q", s)
	}

	v := yuan*100 + cents
	if neg {
		v = -v
	}
	return v, nil
}

func ParseChannel(r io.Reader) ([]domain.Record, []Error, error) {
	return parse(r, domain.SourceChannel)
}

func ParseSystem(r io.Reader) ([]domain.Record, []Error, error) {
	return parse(r, domain.SourceSystem)
}

func parse(r io.Reader, source int32) ([]domain.Record, []Error, error) {
	bizColumn := "out_trade_no"
	amountInYuan := true
	if source == domain.SourceSystem {
		bizColumn = "pay_order_no"
		amountInYuan = false
	}

	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true

	header, err := cr.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("read header: %w", err)
	}

	col := make(map[string]int, len(header))
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, required := range []string{"trade_no", bizColumn, "amount", "trade_time", "status"} {
		if _, ok := col[required]; !ok {
			return nil, nil, fmt.Errorf("missing required column %q", required)
		}
	}

	var (
		records []domain.Record
		errs    []Error
		line    = 1 // 表头是第 1 行
	)

	for {
		row, readErr := cr.Read()
		if readErr == io.EOF {
			break
		}
		line++
		if readErr != nil {
			errs = append(errs, Error{Line: line, Message: readErr.Error()})
			continue
		}

		get := func(name string) string {
			i, ok := col[name]
			if !ok || i >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[i])
		}

		amount, err := parseMoney(get("amount"), amountInYuan)
		if err != nil {
			errs = append(errs, Error{Line: line, Message: "amount: " + err.Error()})
			continue
		}

		fee, err := parseMoney(get("fee"), amountInYuan)
		if err != nil {
			errs = append(errs, Error{Line: line, Message: "fee: " + err.Error()})
			continue
		}

		tradeTime, err := ParseTime(get("trade_time"))
		if err != nil {
			errs = append(errs, Error{Line: line, Message: "trade_time: " + err.Error()})
			continue
		}

		records = append(records, domain.Record{
			Source:    source,
			TradeNo:   get("trade_no"),
			BizNo:     get(bizColumn),
			Amount:    amount,
			Fee:       fee,
			TradeTime: tradeTime,
			Status:    get("status"),
			Remark:    get("remark"),
		})
	}

	return records, errs, nil
}

func parseMoney(s string, inYuan bool) (int64, error) {
	if strings.TrimSpace(s) == "" {
		return 0, nil
	}
	if inYuan {
		return ParseYuanToFen(s)
	}
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %q", s)
	}
	return v, nil
}
