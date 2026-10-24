package coverage

import (
	"os"
	"path/filepath"
	"testing"
)

// 这些预期描述操作效果，不表示对某类 HTTP 方法整体豁免。
func TestUpgradeReviewedHandlersRemainFiniteAndSourceScoped(t *testing.T) {
	for _, tc := range []struct{ source, handler, want string }{
		{"contract_management/internal/transport/httpapi/router.go", "h.listFailedAuditEvents", "ESSENTIAL_SERVICE"},
		{"contract_management/internal/transport/httpapi/router.go", "h.retryFailedAuditEvent", "ESSENTIAL_SERVICE"},
		{"contract_management/internal/transport/httpapi/router.go", "h.recoverApproval", "MUTATE_BUSINESS"},
		{"contract_management/internal/transport/httpapi/handover.go", "h.handoverResponsibilities", "READ_HISTORY"},
		{"contract_management/internal/transport/httpapi/handover.go", "h.handoverReceipt", "READ_HISTORY"},
		{"contract_management/internal/transport/httpapi/handover.go", "h.handoverTransfer", "MUTATE_BUSINESS"},
		{"customer_and_opportunity/internal/modules/handover/handler.go", "handler.responsibilities", "READ_HISTORY"},
		{"customer_and_opportunity/internal/modules/handover/handler.go", "handler.transfer", "MUTATE_BUSINESS"},
		{"customer_and_opportunity/internal/modules/opportunity/routes.go", "handler.ListFailedContractTransfers", "READ_HISTORY"},
		{"customer_and_opportunity/internal/modules/opportunity/routes.go", "handler.RetryFailedContractTransfer", "MUTATE_BUSINESS"},
		{"customer_and_opportunity/internal/modules/presale/routes.go", "handler.ReplaceExecutorAssignments", "MUTATE_BUSINESS"},
		{"customer_and_opportunity/internal/modules/presale/routes.go", "handler.TakeoverExecutingTeamLead", "MUTATE_BUSINESS"},
		{"customer_and_opportunity/internal/modules/presale/routes.go", "handler.RejectExecutionAssignment", "MUTATE_BUSINESS"},
		{"project_management/internal/httpapi/api.go", "h.handoverFreeze", "MUTATE_BUSINESS"},
		{"project_management/internal/httpapi/api.go", "h.handoverReceipt", "READ_HISTORY"},
		{"project_management/internal/httpapi/api.go", "h.retryFailedAuditEvent", "ESSENTIAL_SERVICE"},
		{"Settlement/internal/httpapi/api.go", "a.handoverResponsibilities", "READ_HISTORY"},
		{"Settlement/internal/httpapi/api.go", "a.handoverTransfer", "MUTATE_BUSINESS"},
		{"Settlement/internal/httpapi/api.go", "a.listOutboxEvents", "READ_HISTORY"},
		{"Settlement/internal/httpapi/api.go", "a.retryOutboxEvent", "MUTATE_BUSINESS"},
		{"other/routes.go", "h.handoverReceipt", "UNCLASSIFIED"},
		{"contract_management/internal/transport/httpapi/handover.go", "h.newHandoverFeature", "UNCLASSIFIED"},
	} {
		e := Entry{Source: tc.source, Operation: "UNCLASSIFIED"}
		semanticClassify(&e, tc.handler)
		if e.Operation != tc.want {
			t.Errorf("%s %s: %s want %s", tc.source, tc.handler, e.Operation, tc.want)
		}
	}
}

func TestReviewedHandoverAuxiliaryRegistrationsAreInventoried(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "..", ".ai", "PROJECT_CONTEXT.md")); os.IsNotExist(err) {
		t.Skip("workspace source-closure integration: standalone license-core checkout has no application sources")
	}
	// 扫描实际源码依赖闭包：夹具保留辅助注册项，
	// 不使用固定路由数量或由分类器生成的预期结果。
	r, err := Scan(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"contract_management/internal/transport/httpapi/handover.go", "customer_and_opportunity/internal/modules/handover/handler.go"} {
		found := false
		for _, e := range r.Entries {
			if e.Source == source && e.Kind == "http" && (e.Registration == `group.GET("/responsibilities") -> h.handoverResponsibilities` || e.Registration == `group.GET("/handover/responsibilities") -> handler.responsibilities`) {
				found = true
			}
		}
		if !found {
			t.Errorf("actual auxiliary responsibility route omitted: %s", source)
		}
	}
	for _, source := range []string{
		"contract_management/internal/transport/httpapi/license.go",
		"contract_management/internal/transport/httpapi/audit_buffer.go",
		"project_management/internal/httpapi/commercial_license.go",
		"data_analysis/internal/modules/dashboard/application/business.go",
		"Settlement/internal/service/credit_sync_worker.go",
	} {
		if r.SourceDigests[source] == "" {
			t.Errorf("reviewed execution source not fingerprinted: %s", source)
		}
	}
}

func TestCreditMixedTaskDoesNotInheritUnknownReceiver(t *testing.T) {
	root := t.TempDir()
	source := "Settlement/cmd/worker/main.go"
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, source)), 0700); err != nil {
		t.Fatal(err)
	}
	code := `package main
func main(){ scanner.RunOnce(); taxReconciler.RunOnce(); exporter.RunOnce(); worker.Run(); commercialGate.Run(); creditWorker.RunOnce(); stranger.RunOnce(); _ = outbox.Destination{Name:"PLATFORM_AUDIT"}; _ = outbox.Destination{Name:"PLATFORM_NOTIFICATION"}; _ = outbox.Destination{Name:"TAX_INVOICE_COMMAND"} }`
	if err := os.WriteFile(filepath.Join(root, source), []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	r := Registry{SourceDigests: map[string]string{}}
	if err := discoverMixedTasks(root, &r, Entry{ID: "mixed", Source: source, Application: "settlement", Operation: "MIXED_EXECUTION"}); err != nil {
		t.Fatal(err)
	}
	if len(r.Entries) != 1 || r.Entries[0].Registration != "stranger.RunOnce" || r.Entries[0].Operation != "UNCLASSIFIED" {
		t.Fatalf("only exact reviewed credit receiver may pass: %+v", r.Entries)
	}
}

func TestClassificationFallbackDoesNotExemptNewSourcesOrRegistrations(t *testing.T) {
	for _, tc := range []struct{ source, registration, handler, path string }{
		{"new/routes.go", `r.GET("/healthz") -> <anonymous>`, "<anonymous>", `"/healthz"`},
		{"customer_and_opportunity/internal/modules/new/routes.go", `r.GET("/download") -> h.exportContract`, "h.exportContract", `"/download"`},
		{"project_management/internal/httpapi/api.go", `router.GET("/healthz") -> h.newBusiness`, "h.newBusiness", `"/healthz"`},
		{"project_management/internal/httpapi/api.go", `r.POST("/preview") -> h.previewTemplate`, "h.previewTemplate", `"/preview"`},
	} {
		e := Entry{Source: tc.source, Registration: tc.registration, Operation: "UNCLASSIFIED"}
		classify(&e, "GET", tc.path, tc.handler)
		semanticClassify(&e, tc.handler)
		if e.Operation != "UNCLASSIFIED" {
			t.Errorf("new source/handler inherited %s: %+v", e.Operation, tc)
		}
	}
}
