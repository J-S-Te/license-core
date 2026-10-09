package licensecore

import (
	"crypto/ed25519"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func recoveryFixture() Recovery {
	return Recovery{ProtocolVersion: 1, Issuer: Issuer, ID: "recover-1", ProductID: Product, InstanceID: "instance-1", Environment: "prod", LicenseID: "license-1", LicenseVersion: 1, IssuedAt: 400, ExpiresAt: 500, AnchorAt: 450}
}
func TestRecoveryRoundTripAndTypeIsolation(t *testing.T) {
	pub, key := keys(t)
	trusted := map[string]ed25519.PublicKey{"vendor": pub}
	r := recoveryFixture()
	raw, err := SignRecovery(r, "vendor", key)
	if err != nil {
		t.Fatal(err)
	}
	got, err := VerifyRecovery(raw, trusted)
	if err != nil || got != r {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err = Verify(raw, trusted); err == nil {
		t.Fatal("recovery accepted as license")
	}
	license, err := Sign(fixture(), "vendor", key)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = VerifyRecovery(license, trusted); err == nil {
		t.Fatal("license accepted as recovery")
	}
	if _, err = VerifyRecovery(raw, nil); err == nil {
		t.Fatal("unknown key accepted")
	}
	parts := strings.Split(raw, ".")
	parts[1] = encoded([]byte(`{}`))
	if _, err = VerifyRecovery(strings.Join(parts, "."), trusted); err == nil {
		t.Fatal("tampering accepted")
	}
}
func TestRecoveryStrictProtocol(t *testing.T) {
	pub, key := keys(t)
	trusted := map[string]ed25519.PublicKey{"vendor": pub}
	p, _ := json.Marshal(recoveryFixture())
	h := `{"alg":"EdDSA","typ":"commercial-clock-recovery+jws","kid":"vendor"}`
	for _, bad := range []string{strings.Replace(string(p), `"recovery_id"`, `"RECOVERY_ID"`, 1), strings.Replace(string(p), `"license_version":1`, `"license_version":1,"license_version":1`, 1), strings.Replace(string(p), `"anchor_at":450`, `"anchor_at":450,"extra":true`, 1), string(p) + `{}`} {
		if _, err := VerifyRecovery(tokenWith(t, key, h, bad), trusted); err == nil {
			t.Fatal("invalid recovery JSON accepted")
		}
	}
	for _, bad := range []string{strings.Replace(h, `"EdDSA"`, `"none"`, 1), strings.Replace(h, `"kid":"vendor"`, `"kid":"vendor","jwk":{}`, 1)} {
		if _, err := VerifyRecovery(tokenWith(t, key, bad, string(p)), trusted); err == nil {
			t.Fatal("invalid recovery header accepted")
		}
	}
}
func TestRecoveryBindingAndTime(t *testing.T) {
	r := recoveryFixture()
	l := fixture()
	for _, at := range []int64{400, 499} {
		if err := ValidateRecovery(r, l, time.Unix(at, 0)); err != nil {
			t.Fatal(err)
		}
	}
	for _, at := range []int64{399, 500} {
		if ValidateRecovery(r, l, time.Unix(at, 0)) == nil {
			t.Fatal("time boundary accepted")
		}
	}
	for _, change := range []func(*Recovery){func(r *Recovery) { r.InstanceID = "other" }, func(r *Recovery) { r.Environment = "dev" }, func(r *Recovery) { r.LicenseID = "other" }, func(r *Recovery) { r.LicenseVersion = 2 }, func(r *Recovery) { r.ProductID = "other" }, func(r *Recovery) { r.AnchorAt = 399 }, func(r *Recovery) { r.AnchorAt = 501 }, func(r *Recovery) { r.ExpiresAt = 400 }, func(r *Recovery) { r.IssuedAt = 0 }, func(r *Recovery) { r.ProtocolVersion = 2 }, func(r *Recovery) { r.ID = "" }} {
		bad := r
		change(&bad)
		if ValidateRecovery(bad, l, time.Unix(450, 0)) == nil {
			t.Fatal("invalid/binding recovery accepted")
		}
	}
	_, key := keys(t)
	bad := r
	bad.AnchorAt = 1
	if _, err := SignRecovery(bad, "vendor", key); err == nil {
		t.Fatal("invalid recovery signed")
	}
	if _, err := SignRecovery(r, "vendor", nil); err == nil {
		t.Fatal("invalid key accepted")
	}
}
