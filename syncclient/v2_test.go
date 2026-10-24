package syncclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/J-S-Te/license-core/catalog"
)

func TestV2ReadyUsesCatalogIdentityWithoutDigests(t *testing.T) {
	local, snapshot, _ := fixture(t)
	var ready map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/ready"):
			if err := json.NewDecoder(r.Body).Decode(&ready); err != nil {
				t.Error(err)
			}
			respond(w, map[string]bool{"ready": true})
		case strings.HasSuffix(r.URL.Path, "/snapshot"):
			respond(w, snapshot)
		case strings.HasSuffix(r.URL.Path, "/ack"):
			respond(w, map[string]bool{"acknowledged": true})
		default:
			t.Error("unexpected path")
		}
	}))
	defer server.Close()
	client, err := New(Options{
		BaseURL: server.URL, AllowHTTP: true, ServiceID: "contract-api",
		RegistrationProtocol: catalog.RegistrationProtocolV2, CatalogVersion: catalog.Version,
		Token: func(context.Context) (string, error) { return "token", nil }, Runtime: local,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = client.SyncOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ready["protocol"] != float64(catalog.RegistrationProtocolV2) || ready["catalog_version"] != float64(catalog.Version) {
		t.Fatalf("V2 readiness = %#v", ready)
	}
	if _, exists := ready["coverage_digest"]; exists {
		t.Fatalf("V2 sent coverage digest: %#v", ready)
	}
	if _, exists := ready["image_digest"]; exists {
		t.Fatalf("V2 sent image digest: %#v", ready)
	}
}

func TestV2RejectsPartialOrUnknownRegistrationConfig(t *testing.T) {
	local, _, digest := fixture(t)
	base := Options{BaseURL: "https://platform.invalid", ServiceID: "contract-api", CoverageDigest: "sha256:" + digest, ImageDigest: "sha256:" + digest, Token: func(context.Context) (string, error) { return "token", nil }, Runtime: local}
	cases := []Options{
		func() Options { item := base; item.RegistrationProtocol = catalog.RegistrationProtocolV2; return item }(),
		func() Options { item := base; item.CatalogVersion = catalog.Version; return item }(),
		func() Options {
			item := base
			item.RegistrationProtocol = catalog.RegistrationProtocolV2
			item.CatalogVersion = 99
			return item
		}(),
		func() Options {
			item := base
			item.RegistrationProtocol = 99
			item.CatalogVersion = catalog.Version
			return item
		}(),
	}
	for index, options := range cases {
		if _, err := New(options); err == nil {
			t.Fatalf("case %d accepted partial/unknown registration", index)
		}
	}
}
