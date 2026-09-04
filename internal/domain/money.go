package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// Amount is NGN in minor units (kobo). Stored as int64 to keep pool math
// exact — no floats anywhere near prize money.
type Amount int64

const AmountPerNaira = int64(100)

// FromNairaString parses a decimal string such as "25.00" or "500" into kobo.
func FromNairaString(s string) (Amount, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}
	parts := strings.SplitN(s, ".", 2)
	whole := strings.TrimLeft(parts[0], "0")
	if whole == "" {
		whole = "0"
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if len(fraction) > 2 {
		return 0, fmt.Errorf("more than two decimal places: %q", s)
	}
	fraction = (fraction + "00")[:2] // pad to 2 places

	w, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid naira amount %q: %w", s, err)
	}
	f, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid naira amount %q: %w", s, err)
	}
	v := Amount(w*AmountPerNaira + f)
	if neg {
		v = -v
	}
	return v, nil
}

// NairaString renders the amount as a two-decimal NGN string for BMONI,
// e.g. 2500 kobo -> "25.00".
func (a Amount) NairaString() string {
	neg := a < 0
	abs := int64(a)
	if neg {
		abs = -abs
	}
	n := abs / AmountPerNaira
	f := abs % AmountPerNaira
	out := fmt.Sprintf("%d.%02d", n, f)
	if neg {
		return "-" + out
	}
	return out
}

// SplitPodium divides pool into N descending integer shares. Winner rank 1
// (index 0) receives the largest share; the remainder after integer division
// is given to first place so the sum is exactly the pool.
func SplitPodium(pool Amount, n int) []Amount {
	if n <= 0 {
		return nil
	}
	if n == 1 {
		return []Amount{pool}
	}
	// weight_i = n - i, so the top of an N-winner podium gets N shares down
	// to 1. Sum of 1..n = n(n+1)/2.
	sumWeights := int64(n) * int64(n+1) / 2
	weights := make([]int64, n)
	for i := range weights {
		weights[i] = int64(n - i)
	}
	out := make([]Amount, n)
	var assigned int64
	for i := range out {
		out[i] = Amount(int64(pool) * weights[i] / sumWeights)
		assigned += int64(out[i])
	}
	// Hand the rounding remainder to first place so shares sum exactly.
	out[0] += Amount(int64(pool) - assigned)
	return out
}

// DisplayString renders the amount as an integer naira string without decimals
// for UI display (e.g. 15000000 kobo -> "150000"). Negative amounts keep the
// leading minus sign.
func (a Amount) DisplayString() string {
	neg := a < 0
	abs := int64(a)
	if neg {
		abs = -abs
	}
	n := abs / AmountPerNaira
	out := strconv.FormatInt(n, 10)
	if neg {
		return "-" + out
	}
	return out
}
