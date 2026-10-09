package licensecore

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestIssuerRequired(t *testing.T) {
	for _, issuer := range []string{"", "another-vendor"} {
		l := fixture()
		l.Issuer = issuer
		if Validate(l) == nil {
			t.Fatal("invalid issuer accepted")
		}
		r := recoveryFixture()
		r.Issuer = issuer
		if validateRecoveryFields(r) == nil {
			t.Fatal("invalid recovery issuer accepted")
		}
	}
}
func TestDigest(t *testing.T) {
	pub, key := keys(t)
	raw, err := Sign(fixture(), "vendor", key)
	if err != nil {
		t.Fatal(err)
	}
	trusted := map[string]ed25519.PublicKey{"vendor": pub}
	got, err := Digest(raw, trusted)
	sum := sha256.Sum256([]byte(raw))
	if err != nil || got != hex.EncodeToString(sum[:]) {
		t.Fatal("wrong authenticated digest")
	}
	if got, err := Digest(raw, nil); err == nil || got != "" {
		t.Fatal("unverified digest returned")
	}
}
func TestDiff(t *testing.T) {
	current := fixture()
	current.Applications = append(current.Applications, Application{"customer_portal", 200, 300, "FULL"})
	next := fixture()
	next.Version = 2
	next.Applications[0].ExpiresAt = 400
	next.Applications = append(next.Applications, Application{"data_analysis", 200, 300, "TRIAL"})
	got, err := Diff(next, current)
	if err != nil || len(got) != 3 {
		t.Fatalf("%+v %v", got, err)
	}
	if got[0].Code != "contract_management" || got[0].Kind != UPDATED || got[0].Old == nil || got[0].New == nil || got[1].Code != "customer_portal" || got[1].Kind != REMOVED || got[1].New != nil || got[2].Code != "data_analysis" || got[2].Kind != ADDED || got[2].Old != nil {
		t.Fatalf("wrong diff: %+v", got)
	}
	got[0].Old.ExpiresAt = 999
	if current.Applications[0].ExpiresAt != 300 {
		t.Fatal("diff mutates input")
	}
	same := fixture()
	same.Version = 2
	if changes, err := Diff(same, fixture()); err != nil || len(changes) != 0 {
		t.Fatal("unchanged apps produce diff")
	}
	if _, err := Diff(current, current); err == nil {
		t.Fatal("rollback comparison accepted")
	}
	next.InstanceID = "other"
	if _, err := Diff(next, current); err == nil {
		t.Fatal("cross-binding diff accepted")
	}
}
