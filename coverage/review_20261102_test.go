package coverage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewReviewedPolicyCannotEscapeExactRegistration 验证相似名称、路径和来源不能继承本轮审核。
// 参数：t 为有限策略测试上下文。
// 返回值：无；未审核来源、处理器、动词或任务被豁免时失败。
func TestNewReviewedPolicyCannotEscapeExactRegistration(t *testing.T) {
	for _, tc := range []struct{ source, registration, want string }{
		{"customer_and_opportunity/internal/bootstrap/app.go", `api.POST("/auth/activity") -> authHandler.Activity`, "ESSENTIAL_SERVICE"},
		{"other/app.go", `api.POST("/auth/activity") -> authHandler.Activity`, "UNCLASSIFIED"},
		{"customer_and_opportunity/internal/bootstrap/app.go", `api.POST("/auth/activity") -> authHandler.NewBusiness`, "UNCLASSIFIED"},
		{"customer_and_opportunity/internal/bootstrap/app.go", `api.GET("/auth/activity") -> authHandler.Activity`, "UNCLASSIFIED"},
		{"Settlement/cmd/worker/main.go", "projectReferenceSync.RunOnce", "MUTATE_BUSINESS"},
		{"Settlement/internal/outbox/destinations.go", "projectReferenceSync.RunOnce", "UNCLASSIFIED"},
		{"Settlement/cmd/worker/main.go", "newBusiness.RunOnce", "UNCLASSIFIED"},
	} {
		r := Registry{Entries: []Entry{{ID: "test", Source: tc.source,
			Registration: tc.registration, Kind: "task-policy", Operation: "UNCLASSIFIED"}}}
		applyReviewedEntries20261102(&r)
		if r.Entries[0].Operation != tc.want {
			t.Fatalf("%+v unexpectedly inherited policy: %+v", tc, r.Entries[0])
		}
		if tc.want == "UNCLASSIFIED" && Check(r, r) == nil {
			t.Fatal("unknown boundary passed the unchanged source gate")
		}
	}
}

// TestReviewedImplementationClosureCannotDisappear 验证新增实现文件缺失仍阻断源码指纹门禁。
// 参数：t 为隔离文件系统测试上下文，不修改实际工程源码。
// 返回值：无；漏指纹、缺文件被忽略或绑定内容有误时失败。
func TestReviewedImplementationClosureCannotDisappear(t *testing.T) {
	root := t.TempDir()
	sources := append(append([]string{}, executionSources...), reviewedExecutionSources20261028...)
	sources = append(sources, reviewedExecutionSources20261102...)
	for _, source := range sources {
		path := filepath.Join(root, source)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("synthetic-source-for-digest-test"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	r := Registry{SourceDigests: map[string]string{}}
	if err := fingerprintExecutionSources(root, &r); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"data_analysis/internal/importreport/http.go",
		"Settlement/internal/service/project_reference_sync.go",
		"customer_and_opportunity/internal/modules/portal/report/outbox_ops.go"} {
		if r.SourceDigests[source] == "" {
			t.Fatalf("reviewed implementation missing from source closure: %s", source)
		}
	}
	const missing = "data_analysis/internal/importreport/http.go"
	if err := os.Remove(filepath.Join(root, missing)); err != nil {
		t.Fatal(err)
	}
	if err := fingerprintExecutionSources(root, &r); err == nil || !strings.Contains(err.Error(), missing) {
		t.Fatalf("missing reviewed implementation did not block: %v", err)
	}
}

// TestNewReviewedBoundariesHaveActualPolicies 验证新增真实业务边界按已审核副作用分类。
// 参数：t 为工作区源码集成测试上下文，独立克隆缺应用源码时显式跳过。
// 返回值：无；查询、导出、会话或业务重试被错误分类时失败。
func TestNewReviewedBoundariesHaveActualPolicies(t *testing.T) {
	root := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(root, ".ai", "PROJECT_CONTEXT.md")); os.IsNotExist(err) {
		t.Skip("workspace source-closure integration requires application sources")
	}
	r, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ source, registration, want string }{
		{"contract_management/internal/transport/httpapi/router.go", `api.POST("/integration-outbox/failed/:kind/:id/retry") -> h.retryFailedIntegrationEvent`, "MUTATE_BUSINESS"},
		{"contract_management/internal/transport/httpapi/router.go", `api.GET("/contracts/:contractID/source") -> h.downloadContractSource`, "EXPORT_HISTORY"},
		{"customer_and_opportunity/internal/bootstrap/app.go", `api.POST("/auth/activity") -> authHandler.Activity`, "ESSENTIAL_SERVICE"},
		{"data_analysis/internal/bootstrap/app.go", `api.GET("/dashboard/contract-import-history") -> importreport.HistoryHandler(db)`, "READ_HISTORY"},
		{"Settlement/cmd/worker/main.go", "projectReferenceSync.RunOnce", "MUTATE_BUSINESS"},
		{"Settlement/internal/httpapi/role_finance.go", "POST /api/v1/receipt-differences/:id/confirm-refund", "MUTATE_BUSINESS"},
	} {
		found := false
		for _, entry := range r.Entries {
			if entry.Source == tc.source && entry.Registration == tc.registration {
				found = true
				if entry.Operation != tc.want {
					t.Errorf("%s: got %s, want %s", tc.registration, entry.Operation, tc.want)
				}
			}
		}
		if !found {
			t.Errorf("reviewed execution boundary missing: %s", tc.registration)
		}
	}
}
