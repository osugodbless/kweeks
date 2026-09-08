package domain

import (
	"strings"
)

// NormalizePhone converts the ways a host actually types their number into the
// canonical E.164 form stored on the instructor and used as their BMONI user
// identity. kweeks is a Nigerian product (NGN wallets, NG bank payouts), so a
// bare number without an international prefix is assumed Nigerian and given
// the +234 country code.
//
// Recognised inputs:
//
//	+2348012345678      -> +2348012345678   (E.164 as typed)
//	+234 801 234 5678   -> +2348012345678   (spaces/dashes/parens dropped)
//	002348012345678     -> +2348012345678   (ITU 00 international prefix)
//	2348012345678       -> +2348012345678   (country code, no plus)
//	08012345678         -> +2348012345678   (NG local: leading trunk 0)
//	8012345678          -> +2348012345678   (NG subscriber number, no trunk)
func NormalizePhone(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", ErrBadPhone
	}
	hasPlus := strings.HasPrefix(s, "+")
	digits := stripNonDigits(s)
	if digits == "" {
		return "", ErrBadPhone
	}

	switch {
	case hasPlus:
		return finishE164(digits)
	case strings.HasPrefix(s, "00"):
		// ITU prefix: 00<country code><number>. Drop the two leading zeros.
		if len(digits) < 2+8 {
			return "", ErrBadPhone
		}
		return finishE164(digits[2:])
	case strings.HasPrefix(digits, "234"):
		switch {
		case len(digits) == 13:
			// +234 followed by the 10-digit subscriber number.
			return "+" + digits, nil
		case len(digits) == 14 && digits[3] == '0':
			// 234 + trunk 0 + subscriber number (redundant trunk).
			return "+" + digits[:3] + digits[4:], nil
		}
		return "", ErrBadPhone
	case strings.HasPrefix(digits, "0"):
		// Nigerian local format: 0 + 10-digit subscriber number.
		if len(digits) != 11 {
			return "", ErrBadPhone
		}
		return "+234" + digits[1:], nil
	default:
		// Bare subscriber number without trunk or country code.
		if len(digits) != 10 {
			return "", ErrBadPhone
		}
		return "+234" + digits, nil
	}
}

func finishE164(digits string) (string, error) {
	if len(digits) < 8 || len(digits) > 15 {
		return "", ErrBadPhone
	}
	return "+" + digits, nil
}

func stripNonDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
