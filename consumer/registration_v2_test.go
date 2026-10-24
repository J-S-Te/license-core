package consumer

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/J-S-Te/license-core/catalog"
)

func lookup(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func TestRegistrationConfigurationKeepsV1DefaultAndV2Strict(t *testing.T) {
	empty := func(string) string { return "" }
	protocol, version, err := registrationConfiguration(empty)
	if err != nil || protocol != catalog.RegistrationProtocolV1 || version != 0 {
		t.Fatalf("default registration = %d/%d/%v", protocol, version, err)
	}
	v1 := lookup(map[string]string{"COMMERCIAL_LICENSE_REGISTRATION_PROTOCOL": "1"})
	if protocol, version, err = registrationConfiguration(v1); err != nil || protocol != catalog.RegistrationProtocolV1 || version != 0 {
		t.Fatalf("explicit V1 = %d/%d/%v", protocol, version, err)
	}
	v2 := lookup(map[string]string{"COMMERCIAL_LICENSE_REGISTRATION_PROTOCOL": "2", "COMMERCIAL_LICENSE_CATALOG_VERSION": "1"})
	if protocol, version, err = registrationConfiguration(v2); err != nil || protocol != catalog.RegistrationProtocolV2 || version != catalog.Version {
		t.Fatalf("V2 = %d/%d/%v", protocol, version, err)
	}
	for _, values := range []map[string]string{
		{"COMMERCIAL_LICENSE_REGISTRATION_PROTOCOL": "2"},
		{"COMMERCIAL_LICENSE_CATALOG_VERSION": "1"},
		{"COMMERCIAL_LICENSE_REGISTRATION_PROTOCOL": "3", "COMMERCIAL_LICENSE_CATALOG_VERSION": "1"},
		{"COMMERCIAL_LICENSE_REGISTRATION_PROTOCOL": "2", "COMMERCIAL_LICENSE_CATALOG_VERSION": "2"},
		{"COMMERCIAL_LICENSE_REGISTRATION_PROTOCOL": "1", "COMMERCIAL_LICENSE_CATALOG_VERSION": "1"},
	} {
		if _, _, err = registrationConfiguration(lookup(values)); !errors.Is(err, ErrConfiguration) {
			t.Fatalf("partial/unknown config %v error = %v", values, err)
		}
	}
}

func TestV2ConsumerStartsWithoutDigestsButV1StillRequiresThem(t *testing.T) {
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := x509.MarshalPKIXPublicKey(private.Public())
	if err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(t.TempDir(), "platform.pem")
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: raw}), 0600); err != nil {
		t.Fatal(err)
	}
	base := map[string]string{
		"COMMERCIAL_LICENSE_ENABLED":                  "true",
		"COMMERCIAL_LICENSE_INSTANCE_ID":              "instance",
		"COMMERCIAL_LICENSE_ENVIRONMENT":              "test",
		"COMMERCIAL_LICENSE_SERVICE_ID":               "contract-api",
		"COMMERCIAL_LICENSE_STATE_PATH":               filepath.Join(t.TempDir(), "state.json"),
		"COMMERCIAL_LICENSE_PLATFORM_PUBLIC_KEY_PATH": keyPath,
		"COMMERCIAL_LICENSE_PLATFORM_BASE_URL":        "https://platform.example",
		"COMMERCIAL_LICENSE_CLIENT_ID":                "client",
		"COMMERCIAL_LICENSE_CLIENT_SECRET":            "secret",
	}
	baseLookup := lookup(base)
	v2 := func(name string) string {
		switch name {
		case "COMMERCIAL_LICENSE_REGISTRATION_PROTOCOL":
			return "2"
		case "COMMERCIAL_LICENSE_CATALOG_VERSION":
			return "1"
		default:
			return baseLookup(name)
		}
	}
	if _, err := FromLookup("contract_management", v2); err != nil {
		t.Fatalf("V2 consumer without digests rejected: %v", err)
	}
	legacy := func(name string) string {
		switch name {
		case "COMMERCIAL_LICENSE_COVERAGE_DIGEST":
			return "sha256:" + strings.Repeat("a", 64)
		case "COMMERCIAL_LICENSE_IMAGE_DIGEST":
			return "sha256:" + strings.Repeat("b", 64)
		default:
			return baseLookup(name)
		}
	}
	if _, err := FromLookup("contract_management", legacy); err != nil {
		t.Fatalf("legacy consumer with digests rejected: %v", err)
	}
	v1WithoutDigests := func(name string) string {
		switch name {
		case "COMMERCIAL_LICENSE_REGISTRATION_PROTOCOL":
			return "1"
		case "COMMERCIAL_LICENSE_CATALOG_VERSION":
			return ""
		case "COMMERCIAL_LICENSE_COVERAGE_DIGEST", "COMMERCIAL_LICENSE_IMAGE_DIGEST":
			return ""
		default:
			return baseLookup(name)
		}
	}
	if _, err := FromLookup("contract_management", v1WithoutDigests); !errors.Is(err, ErrConfiguration) {
		t.Fatalf("V1 consumer without digests error = %v", err)
	}
}

func TestCompatibilityRejectsV2RegistrationSwitches(t *testing.T) {
	for _, field := range []string{"REGISTRATION_PROTOCOL", "CATALOG_VERSION"} {
		get := lookup(map[string]string{"COMMERCIAL_LICENSE_" + field: "1"})
		if err := compatibilityAllowed(get, t.TempDir()); !errors.Is(err, ErrConfiguration) {
			t.Fatalf("compatibility ignored %s", field)
		}
	}
}
