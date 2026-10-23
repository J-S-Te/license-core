package coverage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMixedPolicyDoesNotExemptNewTaskOrDestination(t *testing.T) {
	root := t.TempDir()
	source := "Settlement/cmd/worker/main.go"
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, source)), 0700); err != nil {
		t.Fatal(err)
	}
	base := `package main
func main() {
 scanner.RunOnce(ctx); taxReconciler.RunOnce(ctx); exporter.RunOnce(ctx); worker.Run(ctx); commercialGate.Run(ctx, nil); creditWorker.RunOnce(ctx)
 _ = outbox.Destination{Name:"PLATFORM_AUDIT"}
 _ = outbox.Destination{Name:"PLATFORM_NOTIFICATION"}
 _ = outbox.Destination{Name:"TAX_INVOICE_COMMAND"}
 EXTRA
}`
	for _, extra := range []string{"", `newBusiness.RunOnce(ctx)`, `newBusiness.Run(ctx)`, `_ = outbox.Destination{Name:"UNKNOWN_BUSINESS"}`} {
		if err := os.WriteFile(filepath.Join(root, source), []byte(strings.ReplaceAll(base, "EXTRA", extra)), 0600); err != nil {
			t.Fatal(err)
		}
		r := Registry{SourceDigests: map[string]string{}}
		parent := Entry{ID: "mixed", Application: "settlement", Source: source, Operation: "MIXED_EXECUTION"}
		if err := discoverMixedTasks(root, &r, parent); err != nil {
			t.Fatal(err)
		}
		if extra == "" && len(r.Entries) != 0 {
			t.Fatal("known task unresolved")
		}
		if extra != "" && (len(r.Entries) != 1 || r.Entries[0].Operation != "UNCLASSIFIED") {
			t.Fatal("new task silently exempted")
		}
	}
}

func TestMixedWithoutChildrenAndTaskPolicyTamperingRejected(t *testing.T) {
	r := Registry{Entries: []Entry{{ID: "mixed", Operation: "MIXED_EXECUTION"}}}
	if Check(r, r) == nil {
		t.Fatal("mixed exemption without tasks")
	}
	r.Entries = []Entry{{ID: "task", Kind: "task-policy", Operation: "MUTATE_BUSINESS", Reason: "review"}}
	saved := r
	saved.Entries = append([]Entry(nil), r.Entries...)
	saved.Entries[0].Operation = "ESSENTIAL_SERVICE"
	if Check(r, saved) == nil {
		t.Fatal("task policy tampering accepted")
	}
}

func TestUnclassifiedIsNotGreen(t *testing.T) {
	r := Registry{SourceDigests: map[string]string{"route": "abc"}, Entries: []Entry{{ID: "a", Registration: "GET/new", Operation: "UNCLASSIFIED"}}}
	if err := Check(r, r); err == nil || !strings.Contains(err.Error(), "coverage incomplete") {
		t.Fatalf("wanted incomplete, got %v", err)
	}
}
func TestDriftAndDuplicateFail(t *testing.T) {
	r := Registry{SourceDigests: map[string]string{"route": "abc"}}
	n := Registry{SourceDigests: map[string]string{"route": "def"}}
	if Check(r, n) == nil {
		t.Fatal("drift accepted")
	}
	r.Entries = []Entry{{ID: "a"}, {ID: "a"}}
	if Check(r, r) == nil {
		t.Fatal("duplicate accepted")
	}
}
func TestNoHTTPMethodBlanketPolicy(t *testing.T) {
	for _, method := range []string{"GET", "POST"} {
		e := Entry{Operation: "UNCLASSIFIED"}
		classify(&e, method, `"/new-feature"`, "h.unreviewed")
		if e.Operation != "UNCLASSIFIED" {
			t.Fatal("unreviewed route allowed")
		}
	}
	e := Entry{Source: "contract_management/internal/transport/httpapi/router.go", Operation: "UNCLASSIFIED"}
	classify(&e, "POST", `"/contract-templates/:id/preview"`, "h.previewTemplate")
	if e.Operation != "READ_HISTORY" {
		t.Fatal("POST read preview misclassified")
	}
}
func TestExpiredBIEmbeddingPolicyIsSourceAndRouteScoped(t *testing.T) {
	for _, registration := range []string{`api.GET("/embed/:dashboard") -> <anonymous>`, `embedProxy.GET("/:token") -> <anonymous>`, `embedProxy.GET("/:token/*resource") -> <anonymous>`} {
		e := Entry{Source: "data_analysis/internal/bootstrap/app.go", Registration: registration, Operation: "UNCLASSIFIED"}
		classify(&e, "GET", "", "<anonymous>")
		if e.Operation != "MUTATE_BUSINESS" {
			t.Fatal("embedding treated as historical query")
		}
		e.Source = "other/routes.go"
		e.Operation = "UNCLASSIFIED"
		classify(&e, "GET", "", "<anonymous>")
		if e.Operation != "UNCLASSIFIED" {
			t.Fatal("unreviewed source inherited embedding classification")
		}
	}
}

func TestReviewedHandlersAreSourceScoped(t *testing.T) {
	for _, tc := range []struct{ source, handler, op string }{
		{"customer_and_opportunity/internal/modules/customer/routes.go", "handler.PreviewImport", "MUTATE_BUSINESS"},
		{"customer_and_opportunity/internal/modules/customer/routes.go", "handler.CreateExport", "EXPORT_HISTORY"},
		{"data_analysis/internal/bootstrap/app.go", "dashboardHandler.Project", "READ_HISTORY"},
		{"data_analysis/internal/bootstrap/app.go", `dashboardHandler.Business("report")`, "READ_HISTORY"},
		{"data_analysis/internal/bootstrap/app.go", `dashboardHandler.Business("finance")`, "READ_HISTORY"},
		{"data_analysis/internal/bootstrap/app.go", `dashboardHandler.Business("unreviewed")`, "UNCLASSIFIED"},
		{"other/routes.go", `dashboardHandler.Business("report")`, "UNCLASSIFIED"},
		{"Settlement/internal/httpapi/api.go", "a.recoverInvoiceFailure", "MUTATE_BUSINESS"},
		{"other/routes.go", "a.recoverInvoiceFailure", "UNCLASSIFIED"},
		{"data_analysis/internal/bootstrap/app.go", "adminHandler.TriggerAllSources", "MUTATE_BUSINESS"},
		{"customer_and_opportunity/internal/modules/notification/routes.go", "handler.MarkRead", "ESSENTIAL_SERVICE"},
		{"customer_and_opportunity/internal/modules/portalinvite/routes.go", "handler.Verify", "ESSENTIAL_SERVICE"},
		{"new/module/routes.go", "handler.MarkRead", "UNCLASSIFIED"},
	} {
		e := Entry{Source: tc.source, Operation: "UNCLASSIFIED"}
		semanticClassify(&e, tc.handler)
		if e.Operation != tc.op {
			t.Fatalf("%s %s: got %s want %s", tc.source, tc.handler, e.Operation, tc.op)
		}
	}
}

func TestProjectMixedDeliveryEffectsRequiresExactReviewedReceiver(t *testing.T) {
	root := t.TempDir()
	source := "project_management/cmd/sla-notifier/main.go"
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, source)), 0700); err != nil {
		t.Fatal(err)
	}
	base := `package main
func main() {
 service.ScanSlaNotifications(ctx); service.DispatchDeliveryEffects(ctx); service.DispatchNotificationOutbox(ctx)
 EXTRA
}`
	for _, extra := range []string{"", "unknown.DispatchDeliveryEffects(ctx)", "newBusiness.RunOnce(ctx)"} {
		if err := os.WriteFile(filepath.Join(root, source), []byte(strings.ReplaceAll(base, "EXTRA", extra)), 0600); err != nil {
			t.Fatal(err)
		}
		r := Registry{SourceDigests: map[string]string{}}
		parent := Entry{ID: "project-mixed", Application: "project_management", Source: source, Operation: "MIXED_EXECUTION"}
		if err := discoverMixedTasks(root, &r, parent); err != nil {
			t.Fatal(err)
		}
		if extra == "" && len(r.Entries) != 0 {
			t.Fatalf("reviewed project dispatch unexpectedly unresolved: %+v", r.Entries)
		}
		if extra != "" && (len(r.Entries) != 1 || r.Entries[0].Operation != "UNCLASSIFIED") {
			t.Fatalf("new receiver/task inherited reviewed policy: %+v", r.Entries)
		}
	}
	if err := os.WriteFile(filepath.Join(root, source), []byte(strings.ReplaceAll(strings.ReplaceAll(base, "EXTRA", ""), "service.DispatchDeliveryEffects(ctx);", "")), 0600); err != nil {
		t.Fatal(err)
	}
	r := Registry{SourceDigests: map[string]string{}}
	if err := discoverMixedTasks(root, &r, Entry{ID: "project-mixed", Application: "project_management", Source: source, Operation: "MIXED_EXECUTION"}); err == nil || !strings.Contains(err.Error(), "reviewed mixed task removed") {
		t.Fatalf("removed reviewed effects dispatch not rejected: %v", err)
	}
}
