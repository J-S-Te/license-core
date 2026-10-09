package licensecore

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func fixture() License {
	return License{ProtocolVersion: 1, Issuer: Issuer, ID: "license-1", Version: 1, CustomerID: "customer-1", ProductID: Product, Environment: "prod", InstanceID: "instance-1", IssuedAt: 100, NotBefore: 200, Applications: []Application{{"contract_management", 200, 300, "FULL"}}}
}
func keys(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	return pub, key
}
func tokenWith(t *testing.T, key ed25519.PrivateKey, h, p string) string {
	t.Helper()
	s := encoded([]byte(h)) + "." + encoded([]byte(p))
	return s + "." + encoded(ed25519.Sign(key, []byte(s)))
}
func TestSignVerify(t *testing.T) {
	pub, key := keys(t)
	raw, err := Sign(fixture(), "vendor-1", key)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Verify(raw, map[string]ed25519.PublicKey{"vendor-1": pub})
	if err != nil || got.ID != "license-1" {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err = Verify(raw, nil); err == nil {
		t.Fatal("unknown kid accepted")
	}
	parts := strings.Split(raw, ".")
	parts[1] = encoded([]byte(`{}`))
	if _, err = Verify(strings.Join(parts, "."), map[string]ed25519.PublicKey{"vendor-1": pub}); err == nil {
		t.Fatal("tampering accepted")
	}
}
func TestStrictJWS(t *testing.T) {
	pub, key := keys(t)
	p, _ := json.Marshal(fixture())
	trusted := map[string]ed25519.PublicKey{"vendor": pub}
	for _, h := range []string{`{"alg":"none","typ":"commercial-license+jws","kid":"vendor"}`, `{"alg":"EdDSA","typ":"JWT","kid":"vendor"}`, `{"alg":"EdDSA","typ":"commercial-license+jws","kid":"vendor","jwk":{}}`, `{"alg":"EdDSA","alg":"EdDSA","typ":"commercial-license+jws","kid":"vendor"}`, `{"ALG":"EdDSA","typ":"commercial-license+jws","kid":"vendor"}`} {
		if _, err := Verify(tokenWith(t, key, h, string(p)), trusted); err == nil {
			t.Fatalf("accepted %s", h)
		}
	}
	h := `{"alg":"EdDSA","typ":"commercial-license+jws","kid":"vendor"}`
	for _, bad := range []string{strings.Replace(string(p), `"version":1`, `"version":1,"version":1`, 1), strings.Replace(string(p), `"kind":"FULL"`, `"kind":"FULL","extra":true`, 1), string(p) + `{}`, strings.Replace(string(p), `"customer_id"`, `"CUSTOMER_ID"`, 1)} {
		if _, err := Verify(tokenWith(t, key, h, bad), trusted); err == nil {
			t.Fatalf("accepted payload %s", bad)
		}
	}
	raw := tokenWith(t, key, h, string(p))
	for _, bad := range []string{raw + "=", " " + raw, strings.Repeat("a", MaxTokenBytes+1), "a.b", "a.b.c.d"} {
		if _, err := Verify(bad, trusted); err == nil {
			t.Fatal("malformed accepted")
		}
	}
}
func TestEvaluate(t *testing.T) {
	l := fixture()
	for _, tc := range []struct {
		op                 Operation
		at                 int64
		app, instance, env string
		want               bool
	}{{MUTATE_BUSINESS, 200, "contract_management", "instance-1", "prod", true}, {MUTATE_BUSINESS, 299, "contract_management", "instance-1", "prod", true}, {MUTATE_BUSINESS, 300, "contract_management", "instance-1", "prod", false}, {READ_HISTORY, 300, "contract_management", "instance-1", "prod", true}, {EXPORT_HISTORY, 999, "contract_management", "instance-1", "prod", true}, {READ_HISTORY, 199, "contract_management", "instance-1", "prod", false}, {READ_HISTORY, 250, "customer_portal", "instance-1", "prod", false}, {READ_HISTORY, 250, "contract_management", "other", "prod", false}, {READ_HISTORY, 250, "contract_management", "instance-1", "dev", false}, {ESSENTIAL_SERVICE, 1, "anything", "other", "dev", true}, {Operation("unknown"), 250, "contract_management", "instance-1", "prod", false}} {
		err := Evaluate(l, tc.app, tc.op, time.Unix(tc.at, 0), tc.instance, tc.env)
		if (err == nil) != tc.want {
			t.Fatalf("%+v: %v", tc, err)
		}
	}
	l.Version = 0
	if Evaluate(l, "", ESSENTIAL_SERVICE, time.Now(), "", "") == nil {
		t.Fatal("invalid essential accepted")
	}
	l = fixture()
	l.Applications[0].Kind = "TRIAL"
	if Evaluate(l, "contract_management", MUTATE_BUSINESS, time.Unix(300, 0), "instance-1", "prod") == nil {
		t.Fatal("expired trial accepted")
	}
}
func TestValidationAndVersions(t *testing.T) {
	for _, change := range []func(*License){func(l *License) { l.Version = 0 }, func(l *License) { l.ID = strings.Repeat("a", 129) }, func(l *License) { l.CustomerID = "bad\n" }, func(l *License) { l.ProductID = "other" }, func(l *License) { l.ProtocolVersion = 2 }, func(l *License) { l.IssuedAt = 0 }, func(l *License) { l.IssuedAt = 201 }, func(l *License) { l.Applications = nil }, func(l *License) { l.Applications = append(l.Applications, l.Applications[0]) }, func(l *License) { l.Applications[0].Code = "unknown" }, func(l *License) { l.Applications[0].Kind = "fake" }, func(l *License) { l.Applications[0].NotBefore = 0 }, func(l *License) { l.Applications[0].ExpiresAt = 200 }, func(l *License) { l.Applications[0].ExpiresAt = maxTimestamp + 1 }} {
		l := fixture()
		change(&l)
		if Validate(l) == nil {
			t.Fatalf("invalid accepted: %+v", l)
		}
	}
	current := fixture()
	next := fixture()
	next.Version = 2
	if CompareVersions(next, current) != nil {
		t.Fatal("upgrade rejected")
	}
	for _, change := range []func(*License){func(l *License) { l.Version = 1 }, func(l *License) { l.CustomerID = "other" }, func(l *License) { l.InstanceID = "other" }, func(l *License) { l.Environment = "dev" }} {
		l := next
		change(&l)
		if !errors.Is(CompareVersions(l, current), ErrRollback) {
			t.Fatal("rollback accepted")
		}
	}
	_, key := keys(t)
	if _, err := Sign(fixture(), "", key); err == nil {
		t.Fatal("empty kid accepted")
	}
	if _, err := Sign(fixture(), "vendor", nil); err == nil {
		t.Fatal("invalid key accepted")
	}
	key[63] ^= 1
	if _, err := Sign(fixture(), "vendor", key); err == nil {
		t.Fatal("malformed key accepted")
	}
}
func TestRequest(t *testing.T) {
	r := Request{1, Product, "prod", "instance-1", "customer-1"}
	raw, _ := json.Marshal(r)
	got, err := ParseRequest(raw)
	if err != nil || got != r {
		t.Fatalf("%+v %v", got, err)
	}
	for _, bad := range []string{`null`, `{}`, string(raw) + `{}`, strings.Replace(string(raw), `"protocol_version":1`, `"protocol_version":1,"protocol_version":1`, 1), strings.Replace(string(raw), `"customer_id"`, `"CUSTOMER_ID"`, 1), strings.Replace(string(raw), `"prod"`, `null`, 1)} {
		if _, err := ParseRequest([]byte(bad)); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
}

type memoryTime struct {
	value int64
	fail  bool
}

func (s *memoryTime) Load(context.Context) (int64, error) {
	if s.fail {
		return 0, errors.New("storage unavailable")
	}
	return s.value, nil
}
func (s *memoryTime) SaveMax(_ context.Context, n int64) error {
	if s.fail {
		return errors.New("storage unavailable")
	}
	if n > s.value {
		s.value = n
	}
	return nil
}
func TestTimeObserver(t *testing.T) {
	s := &memoryTime{}
	o := TimeObserver{Store: s, Tolerance: 2 * time.Second}
	if o.Observe(context.Background(), time.Unix(300, 0)) != nil {
		t.Fatal("observe failed")
	}
	if o.Observe(context.Background(), time.Unix(298, 0)) != nil || s.value != 300 {
		t.Fatal("tolerance/highwater failed")
	}
	if !errors.Is(o.Observe(context.Background(), time.Unix(297, 0)), ErrRollback) {
		t.Fatal("rollback accepted")
	}
	s.fail = true
	if o.Observe(context.Background(), time.Unix(301, 0)) == nil {
		t.Fatal("storage failure ignored")
	}
}

func TestRenewalPreservesExpiredApplication(t *testing.T) {
	l := fixture()
	l.IssuedAt = 400
	l.NotBefore = 500
	l.Version = 2
	if err := Validate(l); err != nil {
		t.Fatalf("renewal snapshot rejected: %v", err)
	}
	if Evaluate(l, "contract_management", READ_HISTORY, time.Unix(499, 0), l.InstanceID, l.Environment) == nil {
		t.Fatal("global activation bypassed")
	}
	if err := Evaluate(l, "contract_management", READ_HISTORY, time.Unix(500, 0), l.InstanceID, l.Environment); err != nil {
		t.Fatal(err)
	}
	if Evaluate(l, "contract_management", MUTATE_BUSINESS, time.Unix(500, 0), l.InstanceID, l.Environment) == nil {
		t.Fatal("expired preserved application permits writes")
	}
}
