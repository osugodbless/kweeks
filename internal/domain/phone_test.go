package domain

import "testing"

func TestNormalizePhone(t *testing.T) {
	cases := []struct {
		label, input, want string
	}{
		{"E164 as typed", "+2348012345678", "+2348012345678"},
		{"E164 another cc", "+14155550123", "+14155550123"},
		{"spaces", "+234 801 234 5678", "+2348012345678"},
		{"parens and dashes", "+234(801)-234-5678", "+2348012345678"},
		{"00 international prefix", "002348012345678", "+2348012345678"},
		{"country code no plus", "2348012345678", "+2348012345678"},
		{"NG local trunk zero", "08012345678", "+2348012345678"},
		{"NG local leading zero spaces", "0801 234 5678", "+2348012345678"},
		{"NG subscriber no trunk", "8012345678", "+2348012345678"},
		{"cc plus redundant trunk zero", "23408012345678", "+2348012345678"},
		{"leading whitespace trimmed", "  +2348012345678  ", "+2348012345678"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.label, func(t *testing.T) {
			got, err := NormalizePhone(tc.input)
			if err != nil {
				t.Fatalf("NormalizePhone(%q) error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizePhone(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizePhoneRejectsGarbage(t *testing.T) {
	bad := []string{
		"",
		"   ",
		"not-a-phone",
		"+123",               // plus but too short
		"+1234567890123456",  // 16 digits: over E.164 max
		"080",                // trunk zero but too short
		"1234567",            // neither NG subscriber (10) nor anything else
		"234567890123456789", // bare long run with 234 prefix that fits no case
	}
	for _, input := range bad {
		input := input
		t.Run(input, func(t *testing.T) {
			if got, err := NormalizePhone(input); err == nil {
				t.Fatalf("NormalizePhone(%q) = %q, want error", input, got)
			}
		})
	}
}
