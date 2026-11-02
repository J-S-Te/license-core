package coverage

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

// TestActivityPolicyDoesNotExemptUnknownSources 验证有限活动策略不随方法名或导出状态扩大豁免。
// 参数：t 为隔离语法树测试上下文。
// 返回值：无；新方法或移植方法继承豁免、私有方法误报时失败。
func TestActivityPolicyDoesNotExemptUnknownSources(t *testing.T) {
	for _, tc := range []struct{ source, method, receiver, want string }{
		{"contract_management/internal/workflows/number_audit.go", "AuditContractNumbers", "*Activities", "ESSENTIAL_SERVICE"},
		{"contract_management/internal/workflows/import_confirmation.go", "AdvanceImportConfirmation", "*Activities", "MUTATE_BUSINESS"},
		{"contract_management/internal/workflows/new.go", "AuditContractNumbers", "*Activities", "UNCLASSIFIED"},
		{"contract_management/internal/workflows/number_audit.go", "NewBusiness", "*Activities", "UNCLASSIFIED"},
		{"contract_management/internal/workflows/number_audit.go", "NewBusiness", "Activities", "UNCLASSIFIED"},
		{"contract_management/internal/workflows/number_audit.go", "privateHelper", "*Activities", ""},
	} {
		fs := token.NewFileSet()
		file, err := parser.ParseFile(fs, tc.source,
			"package workflows; func (a "+tc.receiver+") "+tc.method+"() {}", 0)
		if err != nil {
			t.Fatal(err)
		}
		entries := activityMethodEntries(fs, file, "contract_management", tc.source)
		if tc.want == "" {
			if len(entries) != 0 {
				t.Fatal("private method entered the activity denominator")
			}
			continue
		}
		if len(entries) != 1 || entries[0].Operation != tc.want {
			t.Fatalf("wrong source-scoped policy for %+v: %+v", tc, entries)
		}
		if tc.want == "UNCLASSIFIED" && Check(Registry{Entries: entries}, Registry{Entries: entries}) == nil {
			t.Fatal("unreviewed activity passed the unchanged source gate")
		}
	}
}

// TestSplitActivityDiscoveryIgnoresTestFiles 验证新的真实活动源会入分母，测试与私有方法不会。
// 参数：t 为隔离文件系统测试上下文。
// 返回值：无；漏掉新增活动源或扫描了测试专用活动时失败。
func TestSplitActivityDiscoveryIgnoresTestFiles(t *testing.T) {
	root := t.TempDir()
	directory := "contract_management/internal/workflows"
	if err := os.MkdirAll(filepath.Join(root, directory), 0700); err != nil {
		t.Fatal(err)
	}
	for name, method := range map[string]string{
		"new.go": "NewBusiness", "new_test.go": "TestActivity", "private.go": "privateHelper",
	} {
		code := "package workflows; func (a *Activities) " + method + "() {}"
		if err := os.WriteFile(filepath.Join(root, directory, name), []byte(code), 0600); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string][]string{}
	if err := appendSplitActivitySources(root, files); err != nil {
		t.Fatal(err)
	}
	if got := files["contract_management"]; len(got) != 1 || got[0] != directory+"/new.go" {
		t.Fatalf("wrong activity source denominator: %+v", files)
	}
}

// TestSplitContractActivityIsInventoried 核验结构体活动包含真实拆分方法且排除私有辅助方法。
// 参数：t 为当前测试上下文；独立克隆未包含业务源码时显式跳过集成断言。
// 返回值：无；漏掉只读巡检活动、遗漏源指纹或误报私有活动时失败。
func TestSplitContractActivityIsInventoried(t *testing.T) {
	root := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(root, ".ai", "PROJECT_CONTEXT.md")); os.IsNotExist(err) {
		t.Skip("workspace source-closure integration requires application sources")
	}
	r, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range r.Entries {
		if entry.Kind != "activity-method" {
			continue
		}
		if entry.Registration == "*Activities.checkLicense" {
			t.Error("private helper is not a registered Temporal activity")
		}
		if entry.Source == "contract_management/internal/workflows/number_audit.go" &&
			entry.Registration == "*Activities.AuditContractNumbers" {
			found = entry.Operation == "ESSENTIAL_SERVICE"
		}
	}
	if !found || r.SourceDigests["contract_management/internal/workflows/number_audit.go"] == "" {
		t.Fatal("real read-only split activity is missing from reviewed inventory")
	}
}
