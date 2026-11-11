package coverage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// settlementReviewFixture 用途：复制四个真实分派源形成隔离扫描夹具。
// 参数：t为当前测试上下文。
// 返回值：临时根路径；真实源不可用时跳过工作区测试。
func settlementReviewFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{"native_financial.go", "contract_import.go", "contract_import_query.go", "contract_import_export.go"} {
		rel := filepath.Join("Settlement/internal/httpapi", name)
		body, err := os.ReadFile(filepath.Join("../..", rel))
		if os.IsNotExist(err) {
			t.Skip("requires workspace application sources")
		}
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestSettlementNovemberReviewBoundaries 用途：验证分文件分派完整展开及读写分类。
// 参数：t为静态源码扫描上下文。
// 返回值：无；漏掉方法、错误豁免业务或缺摘要时失败。
func TestSettlementNovemberReviewBoundaries(t *testing.T) {
	root := settlementReviewFixture(t)
	r := Registry{SourceDigests: map[string]string{}}
	if err := appendSettlementReview20261110(root, &r); err != nil {
		t.Fatal(err)
	}
	if len(r.Entries) != 19 {
		t.Fatalf("got %d boundaries, want 19", len(r.Entries))
	}
	wants := map[string]string{
		"GET /internal/v1/settlement/contract-imports/":                          "READ_HISTORY",
		"GET /api/v1/financial-customers":                                        "READ_HISTORY",
		"POST /api/v1/financial-customers":                                       "MUTATE_BUSINESS",
		"POST /api/v1/financial-baselines/:id/void":                              "MUTATE_BUSINESS",
		"GET /api/v1/financial-baselines/capabilities":                           "READ_HISTORY",
		"POST /api/v1/settlement/contract-imports/:id/reconcile":                 "MUTATE_BUSINESS",
		"POST /internal/v1/settlement/contract-imports/:id/release":              "MUTATE_BUSINESS",
		"POST /api/v1/settlement/contract-imports/:id/exports":                   "EXPORT_HISTORY",
		"GET /api/v1/settlement/contract-imports/:id/exports/:exportID/download": "EXPORT_HISTORY",
	}
	for _, e := range r.Entries {
		if e.Operation == "UNCLASSIFIED" {
			t.Fatalf("reviewed entry unclassified: %+v", e)
		}
		if want, ok := wants[e.Registration]; ok {
			if e.Operation != want {
				t.Errorf("%s got %s want %s", e.Registration, e.Operation, want)
			}
			delete(wants, e.Registration)
		}
	}
	if len(wants) != 0 {
		t.Fatalf("missing expected entries: %v", wants)
	}
	if len(r.SourceDigests) != 4 {
		t.Fatalf("missing dispatch fingerprints: %v", r.SourceDigests)
	}
}

// TestSettlementNovemberDispatchDriftClosed 用途：拒绝删入口、换处理器及新增未知分派继承策略。
// 参数：t为独立磁盘夹具测试上下文。
// 返回值：无；漂移未留下未分类证据或未阻断Check时失败。
func TestSettlementNovemberDispatchDriftClosed(t *testing.T) {
	for _, tc := range []struct{ name, old, replacement, unknown string }{
		{"removed predicate", `path == "/api/v1/financial-baselines/capabilities"`, `path == "/api/v1/removed"`, "/api/v1/removed"},
		{"changed handler", "a.service.VoidNativeFinancialBaseline", "a.service.FutureMutation", "FutureMutation"},
		{"unknown predicate", "path := r.URL.Path", "path := r.URL.Path\n if r.URL.Path == \"/api/v1/future\" { a.future(w, r); return true }", "/api/v1/future"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := settlementReviewFixture(t)
			path := filepath.Join(root, "Settlement/internal/httpapi/native_financial.go")
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(body), tc.old) {
				t.Fatal("fixture anchor missing")
			}
			body = []byte(strings.Replace(string(body), tc.old, tc.replacement, 1))
			if err := os.WriteFile(path, body, 0600); err != nil {
				t.Fatal(err)
			}
			r := Registry{SourceDigests: map[string]string{}}
			if err := appendSettlementReview20261110(root, &r); err != nil {
				t.Fatal(err)
			}
			found := false
			for _, e := range r.Entries {
				if e.Source == "Settlement/internal/httpapi/native_financial.go" {
					if e.Operation != "UNCLASSIFIED" {
						t.Fatalf("changed source inherited policy: %+v", e)
					}
					if strings.Contains(e.Registration, tc.unknown) {
						found = true
					}
				}
			}
			if !found {
				t.Fatalf("missing changed dispatch/handler evidence %s", tc.unknown)
			}
			if Check(r, r) == nil {
				t.Fatal("unreviewed dispatch drift passed Check")
			}
		})
	}
}

// TestSettlementNovemberDispatchSourceRemoved 用途：验证删除已审核分派源时停止扫描。
// 参数：t为隔离源码夹具上下文。
// 返回值：无；缺少真实源却继续登记成功时失败。
func TestSettlementNovemberDispatchSourceRemoved(t *testing.T) {
	root := settlementReviewFixture(t)
	path := filepath.Join(root, "Settlement/internal/httpapi/contract_import.go")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	r := Registry{SourceDigests: map[string]string{}}
	if err := appendSettlementReview20261110(root, &r); err == nil {
		t.Fatal("removed reviewed source passed scanning")
	}
}

// TestSettlementNovemberDispatchFunctionsRemoved 用途：验证删空分派函数仍留下未知证据。
// 参数：t为隔离源码夹具上下文。
// 返回值：无；删除函数后清单缺失未知条目或Check放行时失败。
func TestSettlementNovemberDispatchFunctionsRemoved(t *testing.T) {
	root := settlementReviewFixture(t)
	source := "Settlement/internal/httpapi/native_financial.go"
	if err := os.WriteFile(filepath.Join(root, source), []byte("package httpapi\n"), 0600); err != nil {
		t.Fatal(err)
	}
	r := Registry{SourceDigests: map[string]string{}}
	if err := appendSettlementReview20261110(root, &r); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range r.Entries {
		if entry.Source == source && entry.Operation == "UNCLASSIFIED" {
			found = true
		}
	}
	if !found || Check(r, r) == nil {
		t.Fatal("removed dispatch functions passed without unclassified evidence")
	}
}
