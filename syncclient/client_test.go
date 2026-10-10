package syncclient

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	core "github.com/J-S-Te/license-core"
	runtime "github.com/J-S-Te/license-core/runtime"
)

func fixture(t *testing.T) (*runtime.FileRuntime, Snapshot, string) {
	t.Helper()
	pk, private, _ := ed25519.GenerateKey(rand.Reader)
	vk, vendor, _ := ed25519.GenerateKey(rand.Reader)
	now := time.Now().Unix()
	license := core.License{ProtocolVersion: 1, Issuer: core.Issuer, ID: "license", Version: 1, CustomerID: "customer", ProductID: core.Product, Environment: "test", InstanceID: "instance", IssuedAt: now - 10, NotBefore: now - 10, Applications: []core.Application{{Code: "contract_management", NotBefore: now - 10, ExpiresAt: now + 60, Kind: "FULL"}}}
	raw, err := core.Sign(license, "vendor", vendor)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := runtime.SignSnapshot(runtime.PlatformSnapshot{Protocol: 1, InstanceID: "instance", Environment: "test", Application: "contract_management", ServiceID: "contract-api", Revision: 2, EnforcementState: runtime.Applying, CurrentLicenseJWS: raw, HighestLicenseVersion: 1, SnapshotIssuedAt: now}, "platform", private)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte(snapshot))
	local, err := runtime.NewFileRuntime(filepath.Join(t.TempDir(), "state.json"), runtime.Binding{InstanceID: "instance", Environment: "test", Application: "contract_management", ServiceID: "contract-api"}, map[string]ed25519.PublicKey{"platform": pk}, map[string]ed25519.PublicKey{"vendor": vk})
	if err != nil {
		t.Fatal(err)
	}
	return local, Snapshot{snapshot, hex.EncodeToString(hash[:]), 2}, strings.Repeat("a", 64)
}
func respond(w http.ResponseWriter, data any) {
	_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": data})
}

func TestSyncAppliesBeforeAckAndOfflineExpiry(t *testing.T) {
	local, snapshot, digest := fixture(t)
	var actions []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing machine authentication")
		}
		action := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		actions = append(actions, action)
		switch action {
		case "ready":
			var input map[string]any
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				t.Error(err)
			}
			if input["protocol"] != float64(1) || input["coverage_digest"] != "sha256:"+digest {
				t.Error("bad readiness")
			}
			respond(w, map[string]bool{"ready": true})
		case "snapshot":
			respond(w, snapshot)
		case "ack":
			state, err := local.ReadState(r.Context())
			if err != nil || state.ControlRevision != 2 {
				t.Error("ack before durable apply")
			}
			var ack acknowledgement
			_ = json.NewDecoder(r.Body).Decode(&ack)
			if ack.Digest != snapshot.Digest || ack.Revision != 2 {
				t.Error("wrong ack")
			}
			respond(w, map[string]bool{"acknowledged": true})
		default:
			t.Error("unexpected path")
		}
	}))
	defer server.Close()
	client, err := New(Options{BaseURL: server.URL, AllowHTTP: true, ServiceID: "contract-api", CoverageDigest: "sha256:" + digest, ImageDigest: "sha256:" + digest, Token: func(context.Context) (string, error) { return "test-token", nil }, Runtime: local})
	if err != nil {
		t.Fatal(err)
	}
	if err = client.SyncOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if strings.Join(actions, ",") != "ready,snapshot,ack" {
		t.Fatal(actions)
	}
	server.Close()
	if err = client.SyncOnce(context.Background()); err == nil {
		t.Fatal("offline sync succeeded")
	}
	if err = local.Evaluate(context.Background(), core.MUTATE_BUSINESS, time.Now().Add(2*time.Minute)); !errors.Is(err, core.ErrDenied) {
		t.Fatalf("offline expiration: %v", err)
	}
	if err = local.Evaluate(context.Background(), core.READ_HISTORY, time.Now().Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
}
func TestRejectedSnapshotsNeverAck(t *testing.T) {
	for _, mode := range []string{"digest", "signature", "revision", "oversize"} {
		t.Run(mode, func(t *testing.T) {
			local, snapshot, digest := fixture(t)
			acked := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/ready") {
					respond(w, map[string]bool{"ready": true})
					return
				}
				if strings.HasSuffix(r.URL.Path, "/ack") {
					acked = true
					respond(w, map[string]bool{"acknowledged": true})
					return
				}
				switch mode {
				case "digest":
					snapshot.Digest = digest
				case "signature":
					snapshot.RawJWS = "forged"
					h := sha256.Sum256([]byte(snapshot.RawJWS))
					snapshot.Digest = hex.EncodeToString(h[:])
				case "revision":
					snapshot.Revision = 99
				case "oversize":
					w.Write([]byte(strings.Repeat(" ", maxResponse+1)))
					return
				}
				respond(w, snapshot)
			}))
			defer server.Close()
			client, err := New(Options{BaseURL: server.URL, AllowHTTP: true, ServiceID: "contract-api", CoverageDigest: "sha256:" + digest, ImageDigest: "sha256:" + digest, Token: func(context.Context) (string, error) { return "secret", nil }, Runtime: local})
			if err != nil {
				t.Fatal(err)
			}
			if client.SyncOnce(context.Background()) == nil || acked {
				t.Fatal("invalid snapshot acknowledged")
			}
		})
	}
}
func TestRedirectDoesNotLeakCredentials(t *testing.T) {
	local, _, digest := fixture(t)
	hit := false
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hit = true }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, 307) }))
	defer server.Close()
	options := Options{BaseURL: server.URL, AllowHTTP: true, ServiceID: "contract-api", CoverageDigest: "sha256:" + digest, ImageDigest: "sha256:" + digest, Token: func(context.Context) (string, error) { return "secret", nil }, Runtime: local}
	client, err := New(options)
	if err != nil {
		t.Fatal(err)
	}
	if client.SyncOnce(context.Background()) == nil || hit {
		t.Fatal("redirect followed")
	}
	options.AllowHTTP = false
	if _, err = New(options); err == nil {
		t.Fatal("insecure URL accepted without approval")
	}
	for _, suffix := range []string{"?key=bad", "#frag", "/unexpected"} {
		options.AllowHTTP = true
		options.BaseURL = server.URL + suffix
		if _, err = New(options); err == nil {
			t.Fatal("uncontrolled URL accepted")
		}
	}
}
func TestCancelledLoopStops(t *testing.T) {
	local, _, digest := fixture(t)
	client, err := New(Options{BaseURL: "https://platform.invalid", ServiceID: "contract-api", CoverageDigest: "sha256:" + digest, ImageDigest: "sha256:" + digest, Token: func(context.Context) (string, error) { return "", ErrSync }, Runtime: local})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err = client.Run(ctx, func(error) { calls++; cancel() })
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}
