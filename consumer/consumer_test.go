package consumer

import (
	"context"
	"errors"
	core "github.com/J-S-Te/license-core"
	"os"
	"path/filepath"
	"testing"
)

func TestUnconfiguredPreservesExistingDeploymentButUnknownOperationDenied(t *testing.T) {
	g, err := FromLookup("project_management", func(string) string { return "" })
	if err != nil || g.Check(context.Background(), core.MUTATE_BUSINESS) != nil {
		t.Fatal(err)
	}
	if !errors.Is(g.Check(context.Background(), "unknown"), core.ErrDenied) {
		t.Fatal("unknown operation permitted")
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

func TestCompatibilityCannotDiscardEnrollment(t *testing.T) {
	empty := func(string) string { return "" }
	dir := t.TempDir()
	if err := compatibilityAllowed(empty, dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "signed-state.json"), []byte("previous enrollment"), 0600); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(compatibilityAllowed(empty, dir), ErrConfiguration) {
		t.Fatal("persisted enrollment reverted to compatibility")
	}
	for _, field := range []string{"INSTANCE_ID", "SERVICE_ID", "CLIENT_SECRET", "STATE_PATH"} {
		get := func(name string) string {
			if name == "COMMERCIAL_LICENSE_"+field {
				return "binding"
			}
			return ""
		}
		if !errors.Is(compatibilityAllowed(get, t.TempDir()), ErrConfiguration) {
			t.Fatal("partial enrollment ignored")
		}
	}
}
