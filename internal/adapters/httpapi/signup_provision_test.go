package httpapi

import (
	"context"
	"errors"
	"testing"

	"github.com/osugodbless/kweeks/internal/domain"
)

// flakyMoney wraps wizardMoney and fails create-user until armed=false. It
// models a rail that is configured but temporarily failing at signup.
type flakyMoney struct {
	*wizardMoney
	armFailure bool
}

func (f *flakyMoney) CreateUser(ctx context.Context, id domain.UserIdentity) (string, error) {
	if f.armFailure {
		return "", errors.New("bmoni POST /v1/users: status 500: upstream down")
	}
	return f.wizardMoney.CreateUser(ctx, id)
}

// The BMONI identity must use the explicit first/last names the host typed,
// verbatim - never a re-split of a single name field (surname-first input is
// the classic split trap).
func TestSignupPassesExplicitNamePartsToBmoni(t *testing.T) {
	money := &wizardMoney{}
	api, _ := buildWizardServer(t, money)
	rr := authDo(api, "POST", "/api/auth/signup", map[string]string{
		"firstName": "Chiamaka", "lastName": "Okafor-Osei",
		"email": "c.osei@kweeks.ng", "phone": "+2348000000099", "password": "secret1",
	}, "")
	if rr.Code != 200 {
		t.Fatalf("signup: %d %s", rr.Code, rr.Body.String())
	}
	if !money.createdUser {
		t.Fatalf("signup did not create the BMONI user")
	}
	if money.idSeen.FirstName != "Chiamaka" || money.idSeen.LastName != "Okafor-Osei" {
		t.Fatalf("BMONI identity split the name: first=%q last=%q", money.idSeen.FirstName, money.idSeen.LastName)
	}

	var body map[string]any
	decodeBody(t, rr, &body)
	ins := body["instructor"].(map[string]any)
	if ins["firstName"] != "Chiamaka" || ins["lastName"] != "Okafor-Osei" || ins["name"] != "Chiamaka Okafor-Osei" {
		t.Fatalf("instructor shape wrong: %v", ins)
	}
}

// When BMONI user creation fails at signup the host must still be signed in
// (token present) so they can retry from the wizard instead of being locked out.
func TestSignupStillAuthenticatesWhenUserCreationFails(t *testing.T) {
	money := &flakyMoney{wizardMoney: &wizardMoney{}, armFailure: true}
	api, _ := buildWizardServer(t, money)
	rr := authDo(api, "POST", "/api/auth/signup", map[string]string{
		"firstName": "Tunde", "lastName": "Bakare",
		"email": "tunde@kweeks.ng", "phone": "+2348000000002", "password": "secret1",
	}, "")
	if rr.Code != 200 {
		t.Fatalf("signup: %d %s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	decodeBody(t, rr, &body)
	token, _ := body["token"].(string)
	if token == "" {
		t.Fatalf("signup returned no token when user creation failed")
	}
	if rr := authDo(api, "GET", "/api/auth/me", nil, token); rr.Code != 200 {
		t.Fatalf("me after failed provisioning: %d %s", rr.Code, rr.Body.String())
	}
	// Setup reports unprovisioned but rail reachable, so the wizard offers retry.
	rr = authDo(api, "GET", "/api/wallet/setup", nil, token)
	if rr.Code != 200 {
		t.Fatalf("setup: %d", rr.Code)
	}
	var st map[string]any
	decodeBody(t, rr, &st)
	if st["stage"] != "unprovisioned" || st["railConfigured"] != true || st["bmoniUserId"] != "" {
		t.Fatalf("setup wrong for failed provisioning: %v", st)
	}
}

// The retry endpoint must re-run create-user idempotently: after a signup-time
// failure, calling it provisions the user and advances the wizard to KYC.
func TestRetryCreateBmoniUserAfterSignupFailure(t *testing.T) {
	money := &flakyMoney{wizardMoney: &wizardMoney{}, armFailure: true}
	api, _ := buildWizardServer(t, money)
	rr := authDo(api, "POST", "/api/auth/signup", map[string]string{
		"firstName": "Zainab", "lastName": "Lawal",
		"email": "zainab@kweeks.ng", "phone": "+2348000000003", "password": "secret1",
	}, "")
	if rr.Code != 200 {
		t.Fatalf("signup: %d %s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	decodeBody(t, rr, &body)
	token := body["token"].(string)

	// Rail recovers; the host clicks retry in the wizard.
	money.armFailure = false
	rr = authDo(api, "POST", "/api/wallet/create-user", nil, token)
	if rr.Code != 200 {
		t.Fatalf("create-user retry: %d %s", rr.Code, rr.Body.String())
	}
	if !money.createdUser {
		t.Fatalf("create-user retry did not call the rail")
	}
	rr = authDo(api, "GET", "/api/wallet/setup", nil, token)
	var st map[string]any
	decodeBody(t, rr, &st)
	if st["stage"] != "kyc" || st["bmoniUserId"] != "usr_wiz" {
		t.Fatalf("stage after retry = %v, want kyc with user", st["stage"])
	}

	// Retrying again is a no-op (idempotent): still 200, still one user.
	rr = authDo(api, "POST", "/api/wallet/create-user", nil, token)
	if rr.Code != 200 {
		t.Fatalf("second create-user retry: %d %s", rr.Code, rr.Body.String())
	}
}
