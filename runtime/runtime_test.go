package runtime

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	core "github.com/J-S-Te/license-core"
)

type fixture struct {
	t                              *testing.T
	platformPrivate, vendorPrivate ed25519.PrivateKey
	platformKeys, vendorKeys       map[string]ed25519.PublicKey
	binding                        Binding
	path                           string
	now                            time.Time
	r                              *FileRuntime
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	pp, ps, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	vp, vs, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	f := &fixture{t: t, platformPrivate: ps, vendorPrivate: vs, platformKeys: map[string]ed25519.PublicKey{"platform": pp}, vendorKeys: map[string]ed25519.PublicKey{"vendor": vp}, binding: Binding{"instance", "prod", "contract_management", "contract-api"}, path: filepath.Join(t.TempDir(), "state.json"), now: time.Unix(1800000000, 0)}
	f.r = f.restart()
	return f
}
func (f *fixture) restart() *FileRuntime {
	f.t.Helper()
	r, err := NewFileRuntime(f.path, f.binding, f.platformKeys, f.vendorKeys)
	if err != nil {
		f.t.Fatal(err)
	}
	return r
}
func (f *fixture) license(version uint64, start, end int64, app string) string {
	f.t.Helper()
	l := core.License{ProtocolVersion: core.ProtocolVersion, Issuer: core.Issuer, ID: "license", Version: version, CustomerID: "customer", ProductID: core.Product, Environment: f.binding.Environment, InstanceID: f.binding.InstanceID, IssuedAt: f.now.Unix() - 1000, NotBefore: start, Applications: []core.Application{{Code: app, NotBefore: start, ExpiresAt: end, Kind: "FULL"}}}
	raw, err := core.Sign(l, "vendor", f.vendorPrivate)
	if err != nil {
		f.t.Fatal(err)
	}
	return raw
}
func (f *fixture) snapshot() PlatformSnapshot {
	return PlatformSnapshot{Protocol: Protocol, InstanceID: f.binding.InstanceID, Environment: f.binding.Environment, Application: f.binding.Application, ServiceID: f.binding.ServiceID, Revision: 1, EnforcementState: Enforced, CurrentLicenseJWS: f.license(1, f.now.Unix()-100, f.now.Unix()+600, f.binding.Application), HighestLicenseVersion: 1, SnapshotIssuedAt: f.now.Unix()}
}
func (f *fixture) sign(s PlatformSnapshot) string {
	f.t.Helper()
	raw, err := SignSnapshot(s, "platform", f.platformPrivate)
	if err != nil {
		f.t.Fatal(err)
	}
	return raw
}
func (f *fixture) apply(s PlatformSnapshot) {
	f.t.Helper()
	if err := f.r.ApplySnapshot(context.Background(), f.sign(s), f.now); err != nil {
		f.t.Fatal(err)
	}
}

func TestSnapshotStrictTrustAndBinding(t *testing.T) {
	f := newFixture(t)
	s := f.snapshot()
	raw := f.sign(s)
	if _, err := VerifySnapshot(raw, f.binding, f.platformKeys, f.vendorKeys); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"instance", "environment", "application", "service"} {
		b := f.binding
		switch field {
		case "instance":
			b.InstanceID = "other"
		case "environment":
			b.Environment = "other"
		case "application":
			b.Application = "settlement"
		case "service":
			b.ServiceID = "other"
		}
		if _, err := VerifySnapshot(raw, b, f.platformKeys, f.vendorKeys); err == nil {
			t.Fatal("accepted wrong", field)
		}
	}
	if _, err := VerifySnapshot(s.CurrentLicenseJWS, f.binding, f.platformKeys, f.vendorKeys); err == nil {
		t.Fatal("commercial license accepted as snapshot")
	}
	if _, err := VerifySnapshot(raw, f.binding, f.vendorKeys, f.vendorKeys); err == nil {
		t.Fatal("same trust accepted")
	}
	parts := strings.Split(raw, ".")
	payload, err := decode(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	for name, bad := range map[string]string{
		"duplicate": strings.TrimSuffix(string(payload), "}") + `,"revision":2}`,
		"unknown":   strings.TrimSuffix(string(payload), "}") + `,"jwk":{}}`,
		"alias":     strings.Replace(string(payload), `"serviceID"`, `"ServiceID"`, 1),
		"missing":   strings.Replace(string(payload), `"migrationEligible":false,`, "", 1),
		"null":      strings.Replace(string(payload), `"migrationEligible":false`, `"migrationEligible":null`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			input := parts[0] + "." + encode([]byte(bad))
			badRaw := input + "." + encode(ed25519.Sign(f.platformPrivate, []byte(input)))
			if _, err := VerifySnapshot(badRaw, f.binding, f.platformKeys, f.vendorKeys); err == nil {
				t.Fatal("accepted malformed signed JSON")
			}
		})
	}
	s.CurrentLicenseJWS = f.license(1, f.now.Unix()-100, f.now.Unix()+100, "settlement")
	// A valid license omitting this app is a denial, not a signature error.
	f.apply(s)
	if err := f.r.Evaluate(context.Background(), core.MUTATE_BUSINESS, f.now); !errors.Is(err, core.ErrDenied) {
		t.Fatal(err)
	}
	commercial, err := core.Verify(s.CurrentLicenseJWS, f.vendorKeys)
	if err != nil {
		t.Fatal(err)
	}
	commercial.InstanceID = "other"
	s.CurrentLicenseJWS, err = core.Sign(commercial, "vendor", f.vendorPrivate)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySnapshot(f.sign(s), f.binding, f.platformKeys, f.vendorKeys); err == nil {
		t.Fatal("accepted embedded binding mismatch")
	}
	s.CurrentLicenseJWS = raw
	if _, err := VerifySnapshot(f.sign(s), f.binding, f.platformKeys, f.vendorKeys); err == nil {
		t.Fatal("platform token accepted as commercial license")
	}
}

func TestDurableMigrationLatchAndMissingState(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now); !errors.Is(err, ErrState) {
		t.Fatal(err)
	}
	if err := f.r.Evaluate(ctx, core.ESSENTIAL_SERVICE, f.now); err != nil {
		t.Fatal(err)
	}
	s := f.snapshot()
	s.EnforcementState = Pending
	s.MigrationEligible = true
	s.CurrentLicenseJWS = ""
	s.HighestLicenseVersion = 0
	f.apply(s)
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now); err != nil {
		t.Fatal(err)
	}
	s.Revision++
	s.EnforcementState = Applying
	s.MigrationEligible = false
	f.apply(s)
	f.r = f.restart()
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now); !errors.Is(err, core.ErrDenied) {
		t.Fatal(err)
	}
	s.Revision++
	s.EnforcementState = Pending
	s.MigrationEligible = true
	if err := f.r.ApplySnapshot(ctx, f.sign(s), f.now); !errors.Is(err, ErrRollback) {
		t.Fatal(err)
	}
	s.EnforcementState = Enforced
	s.MigrationEligible = false
	f.apply(s)
	s.Revision++
	s.EnforcementState = Applying
	if err := f.r.ApplySnapshot(ctx, f.sign(s), f.now); !errors.Is(err, ErrRollback) {
		t.Fatal(err)
	}
	info, err := os.Stat(f.path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("incorrect durable permissions", err)
	}
	state, err := f.r.ReadState(ctx)
	if err != nil || !state.Enforced || !state.EnforcementStarted || state.ControlRevision != 3 {
		t.Fatal(state, err)
	}
	if err := os.WriteFile(f.path, []byte(`{"bad":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now); !errors.Is(err, ErrState) {
		t.Fatal(err)
	}
	if err := f.r.ApplySnapshot(ctx, f.sign(f.snapshot()), f.now); !errors.Is(err, ErrState) {
		t.Fatal("corrupt state silently reset", err)
	}
	if err := f.r.Evaluate(ctx, core.ESSENTIAL_SERVICE, f.now); err != nil {
		t.Fatal(err)
	}
}

func TestWholePendingLicenseAndAbsoluteExpiry(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	s := f.snapshot()
	s.PendingLicenseJWS = f.license(2, f.now.Unix()+100, f.now.Unix()+200, "settlement")
	s.HighestLicenseVersion = 2
	f.apply(s)
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(99*time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(100*time.Second)); !errors.Is(err, core.ErrDenied) {
		t.Fatal("pending removal did not take effect", err)
	}
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(50*time.Second)); !errors.Is(err, core.ErrDenied) {
		t.Fatal("clock jitter reactivated old whole license", err)
	}
	s.Revision++
	s.PendingLicenseJWS = ""
	if err := f.r.ApplySnapshot(ctx, f.sign(s), f.now.Add(100*time.Second)); !errors.Is(err, ErrRollback) {
		t.Fatal("pending downgrade accepted", err)
	}
	s.CurrentLicenseJWS = f.license(2, f.now.Unix()+100, f.now.Unix()+200, "settlement")
	f.apply(s)
	f = newFixture(t)
	s = f.snapshot()
	s.CurrentLicenseJWS = f.license(1, f.now.Unix()-100, f.now.Unix()+10, f.binding.Application)
	f.apply(s)
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(10*time.Second)); !errors.Is(err, core.ErrDenied) {
		t.Fatal("absolute expiry extended", err)
	}
	f.r = f.restart()
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(5*time.Second)); !errors.Is(err, core.ErrDenied) {
		t.Fatal("jitter resurrected expired license", err)
	}
	for _, operation := range []core.Operation{core.READ_HISTORY, core.EXPORT_HISTORY} {
		if err := f.r.Evaluate(ctx, operation, f.now.Add(10*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.r.Evaluate(ctx, core.Operation("UNKNOWN"), f.now); !errors.Is(err, core.ErrDenied) {
		t.Fatal(err)
	}
}

func TestVersionAndRevisionRollback(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	s := f.snapshot()
	f.apply(s)
	if err := f.r.ApplySnapshot(ctx, f.sign(s), f.now); !errors.Is(err, ErrRollback) {
		t.Fatal(err)
	}
	s.Revision = 2
	s.CurrentLicenseJWS = f.license(1, f.now.Unix()-100, f.now.Unix()+999, f.binding.Application)
	if err := f.r.ApplySnapshot(ctx, f.sign(s), f.now); !errors.Is(err, ErrRollback) {
		t.Fatal("same-version equivocation", err)
	}
	s.CurrentLicenseJWS = f.license(2, f.now.Unix()-50, f.now.Unix()+999, f.binding.Application)
	s.HighestLicenseVersion = 2
	f.apply(s)
	f.r = f.restart()
	s.Revision = 3
	s.HighestLicenseVersion = 1
	s.CurrentLicenseJWS = f.snapshot().CurrentLicenseJWS
	if err := f.r.ApplySnapshot(ctx, f.sign(s), f.now); !errors.Is(err, ErrRollback) {
		t.Fatal("highest rollback", err)
	}
	s.HighestLicenseVersion = 2
	if err := f.r.ApplySnapshot(ctx, f.sign(s), f.now); !errors.Is(err, ErrRollback) {
		t.Fatal("current rollback", err)
	}
}

func TestClockLatchAndSignedRecovery(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	s := f.snapshot()
	f.apply(s)
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(-300*time.Second)); !errors.Is(err, core.ErrDenied) {
		t.Fatal("300 sec jitter should not latch", err)
	}
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(-301*time.Second)); !errors.Is(err, ErrClock) {
		t.Fatal(err)
	}
	f.r = f.restart()
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now); !errors.Is(err, ErrClock) {
		t.Fatal("restart cleared clock latch", err)
	}
	s.Revision++
	if err := f.r.ApplySnapshot(ctx, f.sign(s), f.now); !errors.Is(err, ErrClock) {
		t.Fatal("snapshot cleared clock latch", err)
	}
	if err := f.r.Evaluate(ctx, core.ESSENTIAL_SERVICE, f.now); err != nil {
		t.Fatal(err)
	}
	recovery := core.Recovery{ProtocolVersion: core.ProtocolVersion, Issuer: core.Issuer, ID: "recovery-1", ProductID: core.Product, InstanceID: f.binding.InstanceID, Environment: f.binding.Environment, LicenseID: "license", LicenseVersion: 1, IssuedAt: f.now.Unix() - 10, ExpiresAt: f.now.Unix() + 10, AnchorAt: f.now.Unix()}
	raw, err := core.SignRecovery(recovery, "vendor", f.vendorPrivate)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.r.RecoverClock(ctx, raw, f.now); err != nil {
		t.Fatal(err)
	}
	f.r = f.restart()
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now); err != nil {
		t.Fatal(err)
	}
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(-301*time.Second)); !errors.Is(err, ErrClock) {
		t.Fatal(err)
	}
	if err := f.r.RecoverClock(ctx, raw, f.now); !errors.Is(err, ErrRollback) {
		t.Fatal("recovery replay accepted", err)
	}
}

func TestConcurrentIndependentRuntimes(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.apply(f.snapshot())
	var wg sync.WaitGroup
	failures := make(chan error, 30)
	for i := 0; i < 30; i++ {
		r := f.restart()
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			failures <- r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(time.Duration(i)*time.Second))
		}(i)
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	state, err := f.r.ReadState(ctx)
	if err != nil || state.ClockHighWater != f.now.Unix()+29 {
		t.Fatal(state, err)
	}
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(-272*time.Second)); !errors.Is(err, ErrClock) {
		t.Fatal("lost max across concurrent writers", err)
	}
}

// A real child process validates that flock, serialization and key trust survive
// process boundaries; parent Go object locks are insufficient for this test.
func TestProcessRestart(t *testing.T) {
	if os.Getenv("LICENSE_RUNTIME_CHILD") == "1" {
		var s struct {
			Path             string
			Binding          Binding
			Platform, Vendor map[string]ed25519.PublicKey
			Now              int64
		}
		if err := json.Unmarshal([]byte(os.Getenv("LICENSE_RUNTIME_CONFIG")), &s); err != nil {
			t.Fatal(err)
		}
		r, err := NewFileRuntime(s.Path, s.Binding, s.Platform, s.Vendor)
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Evaluate(context.Background(), core.MUTATE_BUSINESS, time.Unix(s.Now, 0)); err != nil {
			t.Fatal(err)
		}
		return
	}
	f := newFixture(t)
	f.apply(f.snapshot())
	data, err := json.Marshal(struct {
		Path             string
		Binding          Binding
		Platform, Vendor map[string]ed25519.PublicKey
		Now              int64
	}{f.path, f.binding, f.platformKeys, f.vendorKeys, f.now.Unix() + 60})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		cmd := exec.Command(os.Args[0], "-test.run=^TestProcessRestart$")
		cmd.Env = append(os.Environ(), "LICENSE_RUNTIME_CHILD=1", "LICENSE_RUNTIME_CONFIG="+string(data))
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("child: %v %s", err, output)
		}
	}
	state, err := f.r.ReadState(context.Background())
	if err != nil || state.ClockHighWater != f.now.Unix()+60 {
		t.Fatal(state, err)
	}
}

func TestFileTamperingAndCanceledLock(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.apply(f.snapshot())
	if err := os.Chmod(f.path, 0644); err != nil {
		t.Fatal(err)
	}
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now); !errors.Is(err, ErrState) {
		t.Fatal("insecure state permissions accepted", err)
	}
	if err := os.Chmod(f.path, 0600); err != nil {
		t.Fatal(err)
	}
	locked, err := lock(ctx, f.path+".lock")
	if err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	if err := f.r.Evaluate(canceled, core.MUTATE_BUSINESS, f.now); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	if err := unlock(locked); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(filepath.Dir(f.path), "target.json")
	if err := os.Rename(f.path, target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, f.path); err != nil {
		t.Fatal(err)
	}
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now); !errors.Is(err, ErrState) {
		t.Fatal("symlink state accepted", err)
	}
}
