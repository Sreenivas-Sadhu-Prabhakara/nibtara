package backend

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// Due is an outstanding customer balance and when it started. No interest is
// charged — this is an aging register plus a settlement worksheet.
type Due struct {
	Customer  string  `json:"customer"`
	Amount    float64 `json:"amount"`
	SinceDate string  `json:"sinceDate"` // ISO date the balance started
}

// Validate reports whether the Due is well formed.
func (d Due) Validate() error {
	if d.Customer == "" {
		return fmt.Errorf("customer is required")
	}
	if d.Amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}
	if d.SinceDate != "" {
		if _, err := time.Parse("2006-01-02", d.SinceDate); err != nil {
			return fmt.Errorf("since date must be YYYY-MM-DD")
		}
	}
	return nil
}

// DaysOutstanding is whole days between SinceDate and `today`.
func (d Due) DaysOutstanding(today time.Time) int {
	since, err := time.Parse("2006-01-02", d.SinceDate)
	if err != nil {
		return 0
	}
	return int(today.Sub(since).Hours() / 24)
}

// Bucket classifies aging: current (<60), aging (60–90), stale (>90).
func (d Due) Bucket(today time.Time) string {
	days := d.DaysOutstanding(today)
	switch {
	case days > 90:
		return "stale"
	case days >= 60:
		return "aging"
	default:
		return "current"
	}
}

// SettlementOption is one offer on the settlement slip.
type SettlementOption struct {
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
}

// SettlementOptions builds a defensible set of offers for an amount: pay in
// full, a small goodwill waiver if cleared now, and a three-part instalment.
func SettlementOptions(amount float64) []SettlementOption {
	round := func(v float64) float64 { return math.Round(v*100) / 100 }
	return []SettlementOption{
		{Label: "Pay in full now", Amount: round(amount)},
		{Label: "Clear now, 10% goodwill waiver", Amount: round(amount * 0.90)},
		{Label: "3 monthly instalments", Amount: round(amount / 3)},
	}
}

// Summary totals the outstanding ledger.
type Summary struct {
	Count           int     `json:"count"`
	TotalOutstanding float64 `json:"totalOutstanding"`
}

// Summarize totals outstanding dues.
func Summarize(records []Record) Summary {
	var s Summary
	for _, r := range records {
		s.Count++
		s.TotalOutstanding += r.Headline
	}
	return s
}

// parseEntry decodes+validates a due; headline is the amount, label the customer.
func parseEntry(raw []byte) (float64, string, error) {
	var d Due
	if err := json.Unmarshal(raw, &d); err != nil {
		return 0, "", fmt.Errorf("invalid json")
	}
	if err := d.Validate(); err != nil {
		return 0, "", err
	}
	return d.Amount, d.Customer, nil
}
