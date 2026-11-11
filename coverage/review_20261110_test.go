package coverage

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReviewNovemberBoundaries 用途：核对会话、业务写与迁移入口的真实分类。
// 参数：t 为工作区源码测试上下文。
// 返回值：无；漏扫、未分类或把业务写登记为基础服务时失败。
func TestReviewNovemberBoundaries(t *testing.T) {
	root := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(root, ".ai/PROJECT_CONTEXT.md")); os.IsNotExist(err) {
		t.Skip("requires workspace application sources")
	}
	r, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ source, registration, want string }{
		{"contract_management/internal/transport/httpapi/router.go", `r.POST("/auth/activity") -> <anonymous>`, "ESSENTIAL_SERVICE"},
		{"project_management/internal/httpapi/api.go", `read.POST("/project-references/search") -> h.expenseSearch`, "READ_HISTORY"},
		{"project_management/internal/httpapi/api.go", `write.POST("/spend-snapshots") -> h.expenseSpendSnapshot`, "MUTATE_BUSINESS"},
		{"customer_and_opportunity/internal/modules/importprovision/routes.go", `router.POST("/import-provision/customers/lookup") -> handler.LookupCustomer`, "READ_HISTORY"},
		{"data_analysis/internal/bootstrap/embed.go", `api.GET("/embed/:dashboard") -> <anonymous>`, "MUTATE_BUSINESS"},
		{"reimbursement/internal/transport/httpapi/recovery.go", `group.GET("/claims/:id/spend-capacity") -> h.spendCapacity`, "READ_HISTORY"},
		{"reimbursement/internal/transport/httpapi/recovery.go", `group.POST("/operations/:kind/:id/retry") -> h.operationsRetry`, "MUTATE_BUSINESS"},
		{"reimbursement/internal/transport/httpapi/router.go", `group.POST("/approval-tasks/:id/recover") -> h.command("approval.recover")`, "MUTATE_BUSINESS"},
	} {
		found := false
		for _, e := range r.Entries {
			if e.Source == tc.source && e.Registration == tc.registration {
				found = true
				if e.Operation != tc.want {
					t.Errorf("%s: %s want %s", tc.registration, e.Operation, tc.want)
				}
			}
		}
		if !found {
			t.Errorf("missing reviewed boundary: %s", tc.registration)
		}
	}
	for _, source := range []string{
		"data_analysis/internal/bootstrap/embed.go",
		"customer_and_opportunity/internal/modules/portal/localidentity/handler.go",
		"contract_management/internal/transport/httpapi/local_catalog.go",
	} {
		if r.SourceDigests[source] == "" {
			t.Errorf("missing implementation digest: %s", source)
		}
	}
}

// TestNovemberPolicyRemainsExact 用途：禁止相似来源、方法或处理器继承本轮分类。
// 参数：t 为精确边界负向测试上下文。
// 返回值：无；任何未审核表达式被授予已知分类时失败。
func TestNovemberPolicyRemainsExact(t *testing.T) {
	const source = "data_analysis/internal/bootstrap/embed.go"
	const registration = `api.GET("/embed/:dashboard") -> <anonymous>`
	for _, tc := range []struct{ source, registration string }{
		{"new/embed.go", registration},
		{source, `api.GET("/new") -> <anonymous>`},
		{source, `api.POST("/embed/:dashboard") -> <anonymous>`},
		{source, `api.GET("/embed/:dashboard") -> h.newFeature`},
	} {
		r := Registry{Entries: []Entry{{ID: "test", Source: tc.source,
			Registration: tc.registration, Operation: "UNCLASSIFIED"}}}
		applyReviewedEntries20261110(&r)
		if r.Entries[0].Operation != "UNCLASSIFIED" || Check(r, r) == nil {
			t.Fatalf("unreviewed expression inherited policy: %+v", r.Entries[0])
		}
	}
}
