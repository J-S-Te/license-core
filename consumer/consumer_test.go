package consumer

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	core "github.com/J-S-Te/license-core"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// absentState 返回一个不存在的状态目录，模拟从未登记组件的持久化挂载。
func absentState(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "commercial-license")
}

// enabledLookup 构造一份完整的受控授权配置（ENABLED=true），供“生产路径不变”用例使用。
func enabledLookup(t *testing.T, extra map[string]string) func(string) string {
	t.Helper()
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
	values := map[string]string{
		"COMMERCIAL_LICENSE_ENABLED":                  "true",
		"COMMERCIAL_LICENSE_INSTANCE_ID":              "instance",
		"COMMERCIAL_LICENSE_ENVIRONMENT":              "test",
		"COMMERCIAL_LICENSE_SERVICE_ID":               "contract-api",
		"COMMERCIAL_LICENSE_STATE_PATH":               filepath.Join(t.TempDir(), "state.json"),
		"COMMERCIAL_LICENSE_PLATFORM_PUBLIC_KEY_PATH": keyPath,
		"COMMERCIAL_LICENSE_PLATFORM_BASE_URL":        "https://platform.example",
		"COMMERCIAL_LICENSE_CLIENT_ID":                "client",
		"COMMERCIAL_LICENSE_CLIENT_SECRET":            "secret",
		"COMMERCIAL_LICENSE_COVERAGE_DIGEST":          "sha256:" + strings.Repeat("a", 64),
		"COMMERCIAL_LICENSE_IMAGE_DIGEST":             "sha256:" + strings.Repeat("b", 64),
	}
	for name, value := range extra {
		values[name] = value
	}
	return lookup(values)
}

func TestCompatibilityPreservesExistingDeploymentButUnknownOperationDenied(t *testing.T) {
	g, err := fromLookup("project_management", lookup(map[string]string{
		"COMMERCIAL_LICENSE_ENABLED":       "false",
		"COMMERCIAL_LICENSE_COMPATIBILITY": "true",
	}), absentState(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range []core.Operation{core.READ_HISTORY, core.EXPORT_HISTORY, core.MUTATE_BUSINESS, core.ESSENTIAL_SERVICE} {
		if err := g.Check(context.Background(), op); err != nil {
			t.Fatalf("compatibility denied %s: %v", op, err)
		}
	}
	if !errors.Is(g.Check(context.Background(), "unknown"), core.ErrDenied) {
		t.Fatal("unknown operation permitted")
	}
}

// 未启用受控授权的组件必须显式 opt-in；非法取值按未开启处理（fail-closed）。
func TestCompatibilityRequiresExplicitOptIn(t *testing.T) {
	for _, values := range []map[string]string{
		{},
		{"COMMERCIAL_LICENSE_ENABLED": ""},
		{"COMMERCIAL_LICENSE_ENABLED": "false"},
		{"COMMERCIAL_LICENSE_ENABLED": "false", "COMMERCIAL_LICENSE_COMPATIBILITY": ""},
		{"COMMERCIAL_LICENSE_ENABLED": "false", "COMMERCIAL_LICENSE_COMPATIBILITY": "1"},
		{"COMMERCIAL_LICENSE_ENABLED": "false", "COMMERCIAL_LICENSE_COMPATIBILITY": "yes"},
		{"COMMERCIAL_LICENSE_ENABLED": "false", "COMMERCIAL_LICENSE_COMPATIBILITY": "on"},
		{"COMMERCIAL_LICENSE_ENABLED": "false", "COMMERCIAL_LICENSE_COMPATIBILITY": "false"},
		{"COMMERCIAL_LICENSE_ENABLED": "false", "COMMERCIAL_LICENSE_COMPATIBILITY": "tru"},
		{"COMMERCIAL_LICENSE_COMPATIBILITY": "1"},
	} {
		if _, err := fromLookup("project_management", lookup(values), absentState(t)); !errors.Is(err, ErrConfiguration) {
			t.Fatalf("implicit compatibility permitted for %v: %v", values, err)
		}
	}
}

// 显式 opt-in 才放行，且大小写/首尾空白归一后必须是 true。
func TestCompatibilityAcceptsNormalizedTrue(t *testing.T) {
	for _, values := range []map[string]string{
		{"COMMERCIAL_LICENSE_ENABLED": "false", "COMMERCIAL_LICENSE_COMPATIBILITY": "true"},
		{"COMMERCIAL_LICENSE_ENABLED": "false", "COMMERCIAL_LICENSE_COMPATIBILITY": " TRUE "},
		{"COMMERCIAL_LICENSE_ENABLED": "", "COMMERCIAL_LICENSE_COMPATIBILITY": "True"},
		{"COMMERCIAL_LICENSE_COMPATIBILITY": "true"},
	} {
		g, err := fromLookup("project_management", lookup(values), absentState(t))
		if err != nil {
			t.Fatalf("explicit compatibility rejected for %v: %v", values, err)
		}
		if err := g.Check(context.Background(), core.MUTATE_BUSINESS); err != nil {
			t.Fatal(err)
		}
	}
}

// 显式 opt-in 也不能丢弃既有绑定或持久化登记。
func TestCompatibilityCannotDiscardEnrollment(t *testing.T) {
	for _, field := range []string{"INSTANCE_ID", "ENVIRONMENT", "SERVICE_ID", "STATE_PATH", "PLATFORM_PUBLIC_KEY_PATH", "PLATFORM_BASE_URL", "CLIENT_ID", "CLIENT_SECRET", "COVERAGE_DIGEST", "IMAGE_DIGEST", "REGISTRATION_PROTOCOL", "CATALOG_VERSION"} {
		get := lookup(map[string]string{
			"COMMERCIAL_LICENSE_ENABLED":       "false",
			"COMMERCIAL_LICENSE_COMPATIBILITY": "true",
			"COMMERCIAL_LICENSE_" + field:      "binding",
		})
		if _, err := fromLookup("project_management", get, absentState(t)); !errors.Is(err, ErrConfiguration) {
			t.Fatalf("compatibility ignored binding %s: %v", field, err)
		}
	}

	empty := lookup(map[string]string{
		"COMMERCIAL_LICENSE_ENABLED":       "false",
		"COMMERCIAL_LICENSE_COMPATIBILITY": "true",
	})
	emptyDirectory := t.TempDir()
	if err := compatibilityAllowed(empty, emptyDirectory); err != nil {
		t.Fatal(err)
	}
	if _, err := fromLookup("project_management", empty, emptyDirectory); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(emptyDirectory, "signed-state.json"), []byte("previous enrollment"), 0600); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(compatibilityAllowed(empty, emptyDirectory), ErrConfiguration) {
		t.Fatal("persisted enrollment reverted to compatibility")
	}
	if _, err := fromLookup("project_management", empty, emptyDirectory); !errors.Is(err, ErrConfiguration) {
		t.Fatalf("opt-in discarded persisted enrollment: %v", err)
	}
}

func TestControlledStartupHasNoFallback(t *testing.T) {
	for _, enabled := range []string{"TRUE", "1", "true"} {
		_, err := FromLookup("project_management", func(name string) string {
			if name == "COMMERCIAL_LICENSE_ENABLED" {
				return enabled
			}
			return ""
		})
		if err == nil {
			t.Fatal("incomplete controlled configuration permitted")
		}
	}
}

// ENABLED=true 的受控路径完全不变：不要求、也不依赖兼容开关。
func TestEnabledPathIgnoresCompatibilitySwitch(t *testing.T) {
	if _, err := FromLookup("contract_management", enabledLookup(t, nil)); err != nil {
		t.Fatalf("controlled configuration without compatibility flag rejected: %v", err)
	}
	if _, err := FromLookup("contract_management", enabledLookup(t, map[string]string{"COMMERCIAL_LICENSE_COMPATIBILITY": "true"})); err != nil {
		t.Fatalf("controlled configuration affected by compatibility flag: %v", err)
	}
}
