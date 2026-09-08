package httpapi

import (
	"net/http"
	"strings"
	"testing"
)

// A valid instructor phone typed the way Nigerians actually type it (local
// trunk, no leading +) must not bounce back as "invalid email or password".
// Regression for signup failing on every valid email because the phone format
// the client accepts differs from the one the server demands.
func TestSignupAcceptsLocalFormatPhones(t *testing.T) {
	api, _ := buildAuthServer(t)
	cases := []struct {
		label, phone, want string
	}{
		{"E164 with plus", "+2348012345678", "+2348012345678"},
		{"E164 with spaces", "+234 801 234 5678", "+2348012345678"},
		{"country code no plus", "2348012345678", "+2348012345678"},
		{"NG local trunk zero", "08012345678", "+2348012345678"},
		{"NG subscriber no trunk", "8012345678", "+2348012345678"},
		{"00 international prefix", "002348012345678", "+2348012345678"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.label, func(t *testing.T) {
			rr := authDo(api, "POST", "/api/auth/signup", map[string]string{
				"name": "Adeola Peters", "email": "new+" + tc.phone + "@kweeks.ng",
				"phone": tc.phone, "password": "secret1",
			}, "")
			if rr.Code != http.StatusOK {
				t.Fatalf("signup status %d body %s (phone %q)", rr.Code, rr.Body.String(), tc.phone)
			}
			var body map[string]any
			decodeBody(t, rr, &body)
			ins := body["instructor"].(map[string]any)
			if ins["phone"] != tc.want {
				t.Fatalf("stored phone = %v, want %s", ins["phone"], tc.want)
			}
		})
	}
}

// A genuinely invalid phone must fail with a phone-specific message and a 400,
// never the misleading "invalid email or password".
func TestSignupRejectsGarbagePhoneClearly(t *testing.T) {
	api, _ := buildAuthServer(t)
	rr := authDo(api, "POST", "/api/auth/signup", map[string]string{
		"name": "Adeola Peters", "email": "new2@kweeks.ng",
		"phone": "not-a-phone", "password": "secret1",
	}, "")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("signup status %d want 400", rr.Code)
	}
	var body map[string]string
	decodeBody(t, rr, &body)
	if msg := body["error"]; msg == "invalid email or password" || msg == "" {
		t.Fatalf("error %q must name the phone, not blame email/password", msg)
	} else if !strings.Contains(msg, "phone") && !strings.Contains(msg, "E.164") && !strings.Contains(msg, "+234") {
		t.Fatalf("error %q should point at the phone format", msg)
	}
}
