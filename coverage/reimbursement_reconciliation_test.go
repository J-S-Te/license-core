package coverage

import (
	"go/parser"
	"go/token"
	"testing"
)

// TestReconciliationTaskRequiresExactMethod 用途：禁止其他接收者继承已审核对账任务策略。
// 参数：t 为临时语法树测试上下文。
// 返回值：无；方法缺失、改名或换接收者后仍被登记为基础服务时失败。
func TestReconciliationTaskRequiresExactMethod(t *testing.T) {
	for _, tc := range []struct {
		declaration string
		valid       bool
	}{
		{`func(p *Processor) RunReconciliation(){}`, true},
		{`func(p *Other) RunReconciliation(){}`, false},
		{`func(p *Processor) FutureReconciliation(){}`, false},
		{`func RunReconciliation(){}`, false},
	} {
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, "reconciliation.go", "package jobs;"+tc.declaration, 0)
		if err != nil {
			t.Fatal(err)
		}
		r := Registry{}
		err = appendReimbursementReconciliation(set, file, "reimbursement/internal/jobs/reconciliation.go", &r)
		if (err == nil) != tc.valid {
			t.Fatalf("%s: unexpected review result %v", tc.declaration, err)
		}
		if tc.valid && (len(r.Entries) != 1 || r.Entries[0].Operation != "ESSENTIAL_SERVICE") {
			t.Fatalf("missing reviewed observation task: %+v", r.Entries)
		}
	}
}
