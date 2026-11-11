package coverage

import (
	"fmt"
	"go/ast"
	"go/token"
)

var reimbursementPhaseOneRoutes = map[string]string{
	`group.GET("/parse-exceptions") -> h.parseExceptions`:                       "READ_HISTORY",
	`group.POST("/claims/:id/preview") -> h.claimPreview`:                       "MUTATE_BUSINESS",
	`group.GET("/statistics") -> h.statistics`:                                  "READ_HISTORY",
	`group.GET("/archive-projects") -> h.archiveProjects`:                       "READ_HISTORY",
	`group.POST("/invoice-expense-suggestions") -> h.invoiceExpenseSuggestions`: "MUTATE_BUSINESS",
	`group.POST("/archives/:id/borrow-requests") -> h.borrowRequest`:            "MUTATE_BUSINESS",
	`group.GET("/archive-borrows") -> h.borrowList`:                             "READ_HISTORY",
	`group.GET("/archive-borrows/:id") -> h.borrowDetail`:                       "READ_HISTORY",
	`group.POST("/archive-borrows/:id/decisions") -> h.borrowDecision`:          "MUTATE_BUSINESS",
	`group.POST("/archive-borrows/:id/download") -> h.borrowDownload`:           "EXPORT_HISTORY",
}

// appendReimbursementReminders 登记产生新催办事件的有限业务写入口。
// 参数：set/file为真实语法树，source为准确来源，r为待审清单。
// 返回值：已审方法缺失或接收者改变时返回错误，不将整个worker豁免为基础服务。
func appendReimbursementReminders(set *token.FileSet, file *ast.File, source string, r *Registry) error {
	remaining := map[string]bool{"RunApprovalReminders": true, "ApprovalReminderBatch": true}
	for _, decl := range file.Decls {
		method, ok := decl.(*ast.FuncDecl)
		if !ok || method.Recv == nil || method.Body == nil ||
			expression(set, method.Recv.List[0].Type) != "*Processor" || !remaining[method.Name.Name] {
			continue
		}
		delete(remaining, method.Name.Name)
		r.Entries = append(r.Entries, Entry{
			ID:          reimbursementWorker + ":policy:" + method.Name.Name,
			Application: "reimbursement", Source: source, Kind: "task-policy",
			Registration: "*Processor." + method.Name.Name, Operation: "MUTATE_BUSINESS",
			Reason: "Generates new approval reminder events; checks current business license before scanning and again before transactional enqueue. Existing immutable notification delivery retains its separate essential-service boundary.",
		})
	}
	if len(remaining) != 0 {
		return fmt.Errorf("reviewed reimbursement reminder task removed")
	}
	return nil
}
