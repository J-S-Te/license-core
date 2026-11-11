package coverage

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

// TestReimbursementPhaseOneOperations 用途：新一期入口按精确副作用受许可约束，未知入口仍失败关闭。
// 参数：t为源码入口门禁测试上下文。
// 返回值：无；写借阅误豁免、下载误分类或未知路由被放行时失败。
func TestReimbursementPhaseOneOperations(t *testing.T) {
	source := "reimbursement/internal/transport/httpapi/phase1.go"
	for _, row := range []struct{ registration, want string }{
		{`group.GET("/statistics") -> h.statistics`, "READ_HISTORY"},
		{`group.GET("/parse-exceptions") -> h.parseExceptions`, "READ_HISTORY"},
		{`group.POST("/claims/:id/preview") -> h.claimPreview`, "MUTATE_BUSINESS"},
		{`group.GET("/archive-projects") -> h.archiveProjects`, "READ_HISTORY"},
		{`group.POST("/invoice-expense-suggestions") -> h.invoiceExpenseSuggestions`, "MUTATE_BUSINESS"},
		{`group.POST("/archives/:id/borrow-requests") -> h.borrowRequest`, "MUTATE_BUSINESS"},
		{`group.GET("/archive-borrows") -> h.borrowList`, "READ_HISTORY"},
		{`group.GET("/archive-borrows/:id") -> h.borrowDetail`, "READ_HISTORY"},
		{`group.POST("/archive-borrows/:id/decisions") -> h.borrowDecision`, "MUTATE_BUSINESS"},
		{`group.POST("/archive-borrows/:id/download") -> h.borrowDownload`, "EXPORT_HISTORY"},
		{`group.POST("/archive-borrows/:id/unknown") -> h.unknown`, "UNCLASSIFIED"},
	} {
		if got := reimbursementRouteOperation(source, row.registration); got != row.want {
			t.Fatalf("%s: %s != %s", row.registration, got, row.want)
		}
	}
}

// TestReimbursementReminderPolicies 仅登记精确催办接收者和方法，不豁免生成新事件。
// 参数：t为合成源码的许可盘点上下文。
// 返回值：无；方法改名、缺失或改接收者仍被认可时失败。
func TestReimbursementReminderPolicies(t *testing.T) {
	for _, tc := range []struct {
		declaration string
		valid       bool
	}{
		{`func(p *Processor) RunApprovalReminders(){}; func(p *Processor) ApprovalReminderBatch(){}`, true},
		{`func(p *Other) RunApprovalReminders(){}; func(p *Processor) ApprovalReminderBatch(){}`, false},
		{`func(p *Processor) RunApprovalReminders(){}; func(p *Processor) FutureReminderBatch(){}`, false},
		{`func RunApprovalReminders(){}; func ApprovalReminderBatch(){}`, false},
	} {
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, "approval_reminders.go", "package jobs;"+tc.declaration, 0)
		if err != nil {
			t.Fatal(err)
		}
		registry := Registry{}
		err = appendReimbursementReminders(set, file, "reimbursement/internal/jobs/approval_reminders.go", &registry)
		if (err == nil) != tc.valid {
			t.Fatalf("%s: unexpected review result %v", tc.declaration, err)
		}
		if !tc.valid {
			continue
		}
		if len(registry.Entries) != 2 {
			t.Fatalf("missing policies: %+v", registry.Entries)
		}
		for _, entry := range registry.Entries {
			if entry.Operation != "MUTATE_BUSINESS" {
				t.Fatalf("reminder write exempted: %+v", entry)
			}
		}
	}
}

// TestReimbursementReminderInventory 确认新任务文件实际进入清单而非仅测试分类函数。
// 参数：t为隔离源码目录测试上下文。
// 返回值：无；缺失任务策略或源码摘要时失败。
func TestReimbursementReminderInventory(t *testing.T) {
	root := reimbursementCoverageFixture(t, "")
	source := "reimbursement/internal/jobs/approval_reminders.go"
	raw := `package jobs; func(p *Processor) RunApprovalReminders(){}; func(p *Processor) ApprovalReminderBatch(){}`
	if err := os.WriteFile(filepath.Join(root, source), []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	registry := Registry{Schema: 1, SourceDigests: map[string]string{}}
	if err := appendReimbursementEntries(root, &registry); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range registry.Entries {
		if entry.Source == source && entry.Kind == "task-policy" && entry.Operation == "MUTATE_BUSINESS" {
			count++
		}
	}
	if count != 2 || registry.SourceDigests[source] == "" {
		t.Fatalf("reminder source escaped review: policies=%d", count)
	}
}
