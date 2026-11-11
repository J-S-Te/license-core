package coverage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// reimbursementCoverageFixture 用途：建立不含真实许可、凭据和业务数据的报销源码夹具。
// 参数：t 为测试上下文，extra 为额外任务分支。
// 返回值：由测试自动清理的临时源码根目录。
func reimbursementCoverageFixture(t *testing.T, extra string) string {
	t.Helper()
	root := t.TempDir()
	sources := map[string]string{
		"internal/transport/httpapi/router.go": `package httpapi
func routes(){ group.GET(route.path,h.query(route.kind)); group.POST("/exports",h.export) }`,
		"internal/bootstrap/runtime.go": `package bootstrap
func boot(){ router.GET("/healthz",func(){}) }`,
		"internal/jobs/processor.go": `package jobs
func (p *Processor) Job(){switch row.Kind {case "INVOICE_PARSE":p.parse();case "ARCHIVE":p.archive();case "EXPORT":p.export(); EXTRA}}
func (p *Processor) Outbox(){switch row.Topic {case "project","project_spend":p.deliver();case "audit":p.audit();case "notification":p.notify()}}`,
		"cmd/api/main.go":           "package main;func main(){}",
		"cmd/worker/main.go":        "package main;func main(){}",
		"cmd/migrate/main.go":       "package main;func main(){}",
		"cmd/authz-catalog/main.go": "package main;func main(){}",
	}
	for relative, source := range sources {
		path := filepath.Join(root, "reimbursement", relative)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(strings.ReplaceAll(source, "EXTRA", extra)), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestReimbursementCoverageIncludesFiniteTasks 用途：确认路由、四命令及分任务许可均入清单。
// 参数：t 为测试上下文。
// 返回值：无；源码缺项、任务误分类或父边界缺失时失败。
func TestReimbursementCoverageIncludesFiniteTasks(t *testing.T) {
	root := reimbursementCoverageFixture(t, "")
	registry := Registry{Schema: 1, SourceDigests: map[string]string{}}
	if err := appendReimbursementEntries(root, &registry); err != nil {
		t.Fatal(err)
	}
	if len(registry.Entries) < 14 || len(registry.SourceDigests) != 7 {
		t.Fatalf("missing reimbursement execution boundaries: %d/%d", len(registry.Entries), len(registry.SourceDigests))
	}
	if err := Check(registry, registry); err != nil {
		t.Fatal(err)
	}
}

// TestReimbursementCoverageRejectsUnknownTasks 用途：防止新增任务继承已审核任务的许可分类。
// 参数：t 为测试上下文。
// 返回值：无；未知任务未阻断源码门禁时失败。
func TestReimbursementCoverageRejectsUnknownTasks(t *testing.T) {
	root := reimbursementCoverageFixture(t, `case "NEW_BUSINESS":p.execute()`)
	registry := Registry{Schema: 1, SourceDigests: map[string]string{}}
	if err := appendReimbursementEntries(root, &registry); err != nil {
		t.Fatal(err)
	}
	if err := Check(registry, registry); err == nil {
		t.Fatal("new reimbursement task silently approved")
	}
}

// TestReimbursementCoverageTracksNewRoutes 用途：防止拆文件后的新路由逃离覆盖分母。
// 参数：t 为测试上下文。
// 返回值：无；未知新路由被放行或未绑定源码摘要时失败。
func TestReimbursementCoverageTracksNewRoutes(t *testing.T) {
	for _, registration := range []string{
		`group.GET("/unreviewed",h.newFeature)`,
		`group.Match([]string{"POST"},"/unreviewed",h.newFeature)`,
		`group.Handle("TRACE","/unreviewed",h.newFeature)`,
		`group.Static("/unreviewed","/tmp/test-only")`,
		`group.StaticFS("/unreviewed",testFilesystem)`,
		`group.StaticFile("/unreviewed","/tmp/test-only")`,
		`group.StaticFileFS("/unreviewed","test-only",testFilesystem)`,
		`router.NoRoute(h.newFeature)`,
		`router.NoMethod(h.newFeature)`,
	} {
		root := reimbursementCoverageFixture(t, "")
		path := filepath.Join(root, "reimbursement/internal/transport/httpapi/new.go")
		if err := os.WriteFile(path, []byte("package httpapi;func routes(){"+registration+"}"), 0600); err != nil {
			t.Fatal(err)
		}
		registry := Registry{Schema: 1, SourceDigests: map[string]string{}}
		if err := appendReimbursementEntries(root, &registry); err != nil {
			t.Fatal(err)
		}
		if len(registry.SourceDigests) != 8 || Check(registry, registry) == nil {
			t.Fatalf("new source route escaped review: %s", registration)
		}
	}
}

// TestReimbursementCoverageRejectsSourceSymlink 用途：禁止报销根目录通过符号链接逃离源码边界。
// 参数：t 为测试上下文。
// 返回值：无；扫描器接受链接目录时失败。
func TestReimbursementCoverageRejectsSourceSymlink(t *testing.T) {
	fixture := reimbursementCoverageFixture(t, "")
	root := t.TempDir()
	if err := os.Symlink(filepath.Join(fixture, "reimbursement"), filepath.Join(root, "reimbursement")); err != nil {
		t.Fatal(err)
	}
	registry := Registry{Schema: 1, SourceDigests: map[string]string{}}
	if err := appendReimbursementEntries(root, &registry); err == nil {
		t.Fatal("reimbursement root symlink accepted")
	}
}
