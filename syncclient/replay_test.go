package syncclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestExactPersistedSnapshotRetriesFailedAcknowledgement(t *testing.T) {
	local, snapshot, digest := fixture(t)
	var acknowledgements atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		switch action {
		case "ready":
			respond(w, map[string]bool{"ready": true})
		case "snapshot":
			respond(w, snapshot)
		case "ack":
			state, err := local.ReadState(r.Context())
			if err != nil || state.ControlRevision != snapshot.Revision || state.SnapshotJWS != snapshot.RawJWS {
				t.Error("ACK lacks exact durable signed snapshot", err)
			}
			var ack acknowledgement
			if err := json.NewDecoder(r.Body).Decode(&ack); err != nil || ack.Revision != snapshot.Revision || ack.Digest != snapshot.Digest {
				t.Error("ACK binding mismatch", err)
			}
			if acknowledgements.Add(1) == 1 {
				http.Error(w, "temporary failure", http.StatusServiceUnavailable)
				return
			}
			respond(w, map[string]bool{"acknowledged": true})
		default:
			t.Error("unexpected request")
		}
	}))
	defer server.Close()
	client, err := New(Options{BaseURL: server.URL, AllowHTTP: true, ServiceID: "contract-api", CoverageDigest: "sha256:" + digest, ImageDigest: "sha256:" + digest, Token: func(context.Context) (string, error) { return "test-token", nil }, Runtime: local})
	if err != nil {
		t.Fatal(err)
	}
	if err = client.SyncOnce(context.Background()); err == nil {
		t.Fatal("ACK failure swallowed")
	}
	for i := 0; i < 2; i++ {
		if err = client.SyncOnce(context.Background()); err != nil {
			t.Fatal("exact snapshot prevented ACK retry", err)
		}
	}
	if acknowledgements.Load() != 3 {
		t.Fatal("repeated sync did not retry exact ACK")
	}
}
