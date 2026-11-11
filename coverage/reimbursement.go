package coverage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const reimbursementWorker = "reimbursement:reimbursement/cmd/worker/main.go"

// appendReimbursementEntries 用途：盘点报销实际源码、命令和有限任务，不沿用整进程豁免。
// 参数：root 为工作区根目录，r 为待补充的源码清单。
// 返回值：报销缺关键入口、包含符号链接或解析失败时返回错误。
func appendReimbursementEntries(root string, r *Registry) error {
	directory := filepath.Join(root, "reimbursement")
	if info, err := os.Lstat(directory); os.IsNotExist(err) {
		return nil // 原六系统工作区仍可独立复核。
	} else if err != nil {
		return err
	} else if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("reimbursement source root must be a regular directory")
	}
	for _, name := range []string{"cmd", "internal", "migrations", "authz"} {
		err := filepath.WalkDir(filepath.Join(directory, name), func(path string, entry fs.DirEntry, err error) error {
			if os.IsNotExist(err) && (name == "migrations" || name == "authz") {
				return nil
			}
			if err != nil {
				return err
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("reimbursement source symlink rejected")
			}
			if entry.IsDir() || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			return inventoryReimbursementSource(root, path, r)
		})
		if err != nil {
			return err
		}
	}
	for _, source := range []string{"cmd/api/main.go", "cmd/worker/main.go", "cmd/migrate/main.go",
		"cmd/authz-catalog/main.go", "internal/bootstrap/runtime.go", "internal/jobs/processor.go",
		"internal/transport/httpapi/router.go"} {
		if r.SourceDigests["reimbursement/"+source] == "" {
			return fmt.Errorf("reimbursement execution source missing: %s", source)
		}
	}
	return nil
}

// inventoryReimbursementSource 用途：绑定源码摘要并发现报销 HTTP、命令和后台分派。
// 参数：root 为工作区，path 为实际文件，r 为输出清单。
// 返回值：文件读取、相对路径或 Go 解析错误。
func inventoryReimbursementSource(root, path string, r *Registry) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	source := filepath.ToSlash(relative)
	sum := sha256.Sum256(raw)
	r.SourceDigests[source] = hex.EncodeToString(sum[:])
	if !strings.HasSuffix(source, ".go") {
		return nil
	}
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, source, raw, 0)
	if err != nil {
		return err
	}
	appendReimbursementRoutes(set, file, source, r)
	if strings.HasPrefix(source, "reimbursement/cmd/") && strings.HasSuffix(source, "/main.go") {
		appendReimbursementCommand(source, r)
	}
	if source == "reimbursement/internal/jobs/processor.go" {
		return appendReimbursementTasks(set, file, source, r)
	}
	if source == "reimbursement/internal/jobs/reconciliation.go" {
		return appendReimbursementReconciliation(set, file, source, r)
	}
	if source == "reimbursement/internal/jobs/archive_download_retention.go" {
		return appendReimbursementDownloadRetention(set, file, source, r)
	}
	if source == "reimbursement/internal/jobs/approval_reminders.go" {
		return appendReimbursementReminders(set, file, source, r)
	}
	return nil
}

// appendReimbursementDownloadRetention 用途：登记临时导出包的有界存储清理任务。
// 参数：set/file 为任务语法树，source 为精确来源，r 为候选清单。
// 返回值：已审接收者或清理方法缺失时返回错误，不豁免封存档案及财务删除。
func appendReimbursementDownloadRetention(set *token.FileSet, file *ast.File, source string, r *Registry) error {
	remaining := map[string]bool{"RunArchiveDownloadRetention": true, "CleanupArchiveDownloads": true}
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
			Registration: "*Processor." + method.Name.Name, Operation: "ESSENTIAL_SERVICE",
			Reason: "Bounded retention of regenerable exp_archive_download export blobs only; preserves sealed archive manifests, source documents, financial facts and jobs/audit. No business deletion or new export request.",
		})
	}
	if len(remaining) != 0 {
		return fmt.Errorf("reviewed reimbursement download retention task removed")
	}
	return nil
}

// appendReimbursementReconciliation 用途：登记实际运行的只观察型对账任务。
// 参数：set/file 为真实语法树，source 为精确实现来源，r 为候选清单。
// 返回值：任务方法消失时返回错误；不豁免人工重建、重投或财务命令。
func appendReimbursementReconciliation(set *token.FileSet, file *ast.File, source string, r *Registry) error {
	for _, decl := range file.Decls {
		method, ok := decl.(*ast.FuncDecl)
		if !ok || method.Recv == nil || method.Body == nil || method.Name.Name != "RunReconciliation" {
			continue
		}
		if expression(set, method.Recv.List[0].Type) != "*Processor" {
			continue
		}
		r.Entries = append(r.Entries, Entry{
			ID:          reimbursementWorker + ":policy:RunReconciliation",
			Application: "reimbursement", Source: source, Kind: "task-policy",
			Registration: "*Processor.RunReconciliation", Operation: "ESSENTIAL_SERVICE",
			Reason: "Observes existing immutable financial snapshots via GET and stores only bounded technical cursor/findings; does not repair projections, queue work or change financial facts. Explicit rebuild remains MUTATE_BUSINESS.",
		})
		return nil
	}
	return fmt.Errorf("reviewed reimbursement reconciliation task removed")
}

// appendReimbursementRoutes 用途：发现所有报销源码中的 HTTP 注册，未知分文件仍进入分母。
// 参数：set/file 为语法树，source 为源码路径，r 为输出清单。
// 返回值：无；未知注册保留未分类，后续审核门禁拒绝。
func appendReimbursementRoutes(set *token.FileSet, file *ast.File, source string, r *Registry) {
	ordinal := 0
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !strings.Contains(
			"|GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS|Any|Handle|HandleFunc|"+
				"Match|Static|StaticFS|StaticFile|StaticFileFS|NoRoute|NoMethod|",
			"|"+selector.Sel.Name+"|") {
			return true
		}
		arguments := []string{}
		for _, argument := range call.Args[:len(call.Args)-1] {
			arguments = append(arguments, expression(set, argument))
		}
		handler := expression(set, call.Args[len(call.Args)-1])
		if _, anonymous := call.Args[len(call.Args)-1].(*ast.FuncLit); anonymous {
			handler = "<anonymous>"
		}
		registration := expression(set, call.Fun) + "(" + strings.Join(arguments, ", ") + ") -> " + handler
		ordinal++
		r.Entries = append(r.Entries, Entry{
			ID: fmt.Sprintf("reimbursement:%s:%03d", source, ordinal), Application: "reimbursement",
			Source: source, Kind: "http", Registration: registration,
			Operation: reimbursementRouteOperation(source, registration),
			Reason:    "Exact reviewed registration; finite loop tables and handler bodies remain source-digest bound",
		})
		return true
	})
}

// reimbursementRouteOperation 用途：按精确源码及注册项分类，不按 GET 或处理器名称泛化。
// 参数：source 为源码路径，registration 为完整注册表达式。
// 返回值：已复核操作类型；未知入口返回 UNCLASSIFIED。
func reimbursementRouteOperation(source, registration string) string {
	if source == "reimbursement/internal/transport/httpapi/phase1.go" {
		if operation := reimbursementPhaseOneRoutes[registration]; operation != "" {
			return operation
		}
	}
	if source == "reimbursement/internal/bootstrap/runtime.go" {
		for _, name := range []string{
			`router.GET("/healthz") -> <anonymous>`, `router.GET("/readyz") -> <anonymous>`,
			`health.HandleFunc("/healthz") -> <anonymous>`, `health.HandleFunc("/readyz") -> <anonymous>`,
			`health.HandleFunc("/readyz") -> workerReadinessHandler(runtime)`,
			`health.HandleFunc("/metrics") -> metricsHandler(runtime)`,
			`router.GET("/auth/login") -> gin.WrapF(auth.Login)`,
			`router.GET("/auth/callback") -> gin.WrapF(auth.Callback)`,
			`router.POST("/auth/backchannel-logout") -> gin.WrapF(auth.BackchannelLogout)`,
			`router.GET("/auth/logout") -> gin.WrapF(auth.Logout)`,
			`router.POST("/auth/local-logout") -> <anonymous>`,
			`router.GET("/logged-out") -> <anonymous>`, `router.GET("/auth/session") -> <anonymous>`,
			`router.POST("/auth/activity") -> gin.WrapF(auth.Activity)`,
		} {
			if registration == name {
				return "ESSENTIAL_SERVICE"
			}
		}
	}
	if source == "reimbursement/internal/transport/httpapi/router.go" {
		if operation := reimbursementBusinessRoutes[registration]; operation != "" {
			return operation
		}
	}
	if source == "reimbursement/internal/transport/httpapi/recovery.go" {
		if operation := reimbursementRecoveryRoutes[registration]; operation != "" {
			return operation
		}
	}
	return "UNCLASSIFIED"
}

var reimbursementBusinessRoutes = map[string]string{
	`group.POST("/approval-tasks/:id/recover") -> h.command("approval.recover")`:                 "MUTATE_BUSINESS",
	`group.GET("/claims/:id/approval-rounds") -> h.approvalRounds`:                               "READ_HISTORY",
	`group.GET("/invoices/:id/controls") -> h.invoiceControls`:                                   "READ_HISTORY",
	`group.POST("/invoices/:id/corrections") -> h.command("invoice.correction.request")`:         "MUTATE_BUSINESS",
	`group.POST("/invoice-corrections/:id/decisions") -> h.command("invoice.correction.decide")`: "MUTATE_BUSINESS",
	`group.POST("/claims/:id/invoice-reconfirmations") -> h.command("claim.invoice.reconfirm")`:  "MUTATE_BUSINESS",
	`group.POST("/invoices/:id/sources") -> h.command("invoice.source.append")`:                  "MUTATE_BUSINESS",
	`group.POST("/invoices/:id/risks") -> h.command("invoice.risk.record")`:                      "MUTATE_BUSINESS",
	`group.POST("/invoice-risks/:id/resolutions") -> h.command("invoice.risk.close")`:            "MUTATE_BUSINESS",
	`group.POST("/archives/:id/packages") -> h.archivePackage`:                                   "EXPORT_HISTORY",
	`group.GET("/archive-packages/:id") -> h.archivePackageStatus`:                               "EXPORT_HISTORY",
	`group.POST("/archive-packages/:id/download") -> h.archiveDownload`:                          "EXPORT_HISTORY",

	`group.GET(route.path) -> h.query(route.kind)`:                                           "READ_HISTORY",
	`group.Handle(route.method, route.path) -> h.command(route.operation)`:                   "MUTATE_BUSINESS",
	`group.POST("/claims/:id/reopen") -> h.command("claim.reopen")`:                          "MUTATE_BUSINESS",
	`group.POST("/claims/:id/approved-void") -> h.command("claim.approved_void")`:            "MUTATE_BUSINESS",
	`group.POST("/claims/:id/payment-exceptions") -> h.command("payment.exception.record")`:  "MUTATE_BUSINESS",
	`group.GET("/claims/:id/payment-exceptions") -> h.query("payment-exceptions")`:           "READ_HISTORY",
	`group.GET("/payment-exceptions/:id") -> h.query("payment-exception")`:                   "READ_HISTORY",
	`group.POST("/payment-exceptions/:id/reviews") -> h.command("payment.exception.review")`: "MUTATE_BUSINESS",
	`group.GET("/configurations/:kind") -> h.configQuery`:                                    "READ_HISTORY",
	`group.POST("/configurations/:kind") -> h.configSave`:                                    "MUTATE_BUSINESS",
	`group.POST("/configurations/:kind/:id/publish") -> h.configPublish`:                     "MUTATE_BUSINESS",
	`group.POST("/configurations/:kind/:id/retire") -> h.configLifecycle("config.retire")`:   "MUTATE_BUSINESS",
	`group.POST("/configurations/:kind/:id/replace") -> h.configLifecycle("config.replace")`: "MUTATE_BUSINESS",
	`group.GET("/approval-tasks/:id/history") -> h.query("approval-history")`:                "READ_HISTORY",
	`group.POST("/claims/:id/evidence/:file_id/download-ticket") -> h.downloadEvidence`:      "EXPORT_HISTORY",
	`group.GET("/project-options") -> h.projects`:                                            "READ_HISTORY",
	`group.POST("/project-references/validate") -> h.validateProjects`:                       "READ_HISTORY",
	`group.POST("/invoice-documents/upload-sessions") -> h.upload`:                           "MUTATE_BUSINESS",
	`group.POST("/invoice-documents/:id/complete") -> h.complete`:                            "MUTATE_BUSINESS",
	`group.POST("/invoice-documents/:id/parse") -> h.parse`:                                  "MUTATE_BUSINESS",
	`group.GET("/invoices/:id/sources") -> h.sources`:                                        "READ_HISTORY",
	`group.POST("/invoices/:id/files/:file_id/download-ticket") -> h.download`:               "EXPORT_HISTORY",
	`group.POST("/exports") -> h.export`:                                                     "EXPORT_HISTORY",
}

// appendReimbursementCommand 用途：登记四个已知命令，worker 必须保留分任务许可。
// 参数：source 为命令入口路径，r 为输出清单。
// 返回值：无；未知可执行入口保留未分类。
func appendReimbursementCommand(source string, r *Registry) {
	name := filepath.Base(filepath.Dir(source))
	operation := "UNCLASSIFIED"
	switch name {
	case "api", "migrate", "authz-catalog":
		operation = "ESSENTIAL_SERVICE"
	case "worker":
		operation = "MIXED_EXECUTION"
	}
	r.Entries = append(r.Entries, Entry{ID: "reimbursement:" + source, Application: "reimbursement",
		Source: source, Kind: "executable", Registration: name, Operation: operation,
		Reason: "Bootstrap or controlled migration only; business HTTP and worker tasks remain separately gated"})
}

// appendReimbursementTasks 用途：从实际 Job/Outbox 分派提取有限任务许可，未知任务不放行。
// 参数：set/file 为后台语法树，source 为文件路径，r 为输出清单。
// 返回值：Job 或 Outbox 分派缺失时返回错误。
func appendReimbursementTasks(set *token.FileSet, file *ast.File, source string, r *Registry) error {
	found := map[string]bool{}
	for _, declaration := range file.Decls {
		method, ok := declaration.(*ast.FuncDecl)
		if !ok || method.Recv == nil || expression(set, method.Recv.List[0].Type) != "*Processor" {
			continue
		}
		field := map[string]string{"Job": "row.Kind", "Outbox": "row.Topic"}[method.Name.Name]
		if field == "" {
			continue
		}
		ast.Inspect(method.Body, func(node ast.Node) bool {
			dispatch, ok := node.(*ast.SwitchStmt)
			if ok && dispatch.Tag != nil && expression(set, dispatch.Tag) == field {
				found[method.Name.Name] = true
				appendReimbursementCases(set, dispatch, method.Name.Name, source, r)
			}
			return true
		})
	}
	if !found["Job"] || !found["Outbox"] {
		return fmt.Errorf("reimbursement Job/Outbox dispatch missing; review before approval")
	}
	return nil
}

// appendReimbursementCases 用途：按实际分支逐项建立任务子策略，禁止未知任务继承豁免。
// 参数：set/dispatch 为语法树，method 为分派名，source 为源码路径，r 为清单。
// 返回值：无；动态或未审核分支保留 UNCLASSIFIED。
func appendReimbursementCases(set *token.FileSet, dispatch *ast.SwitchStmt, method, source string, r *Registry) {
	policies := map[string]map[string]string{
		"Job": {"INVOICE_PARSE": "MUTATE_BUSINESS", "ARCHIVE": "EXPORT_HISTORY", "EXPORT": "EXPORT_HISTORY", "ARCHIVE_PACKAGE": "EXPORT_HISTORY"},
		"Outbox": {"project": "ESSENTIAL_SERVICE", "project_spend": "ESSENTIAL_SERVICE", "project_rebuild": "ESSENTIAL_SERVICE",
			"audit": "ESSENTIAL_SERVICE", "notification": "ESSENTIAL_SERVICE"},
	}
	for _, statement := range dispatch.Body.List {
		for _, label := range statement.(*ast.CaseClause).List {
			name, operation := expression(set, label), "UNCLASSIFIED"
			if literal, ok := label.(*ast.BasicLit); ok {
				if text, err := strconv.Unquote(literal.Value); err == nil {
					name = text
					if reviewed := policies[method][name]; reviewed != "" {
						operation = reviewed
					}
				}
			}
			r.Entries = append(r.Entries, Entry{ID: reimbursementWorker + ":policy:" + method + ":" + name,
				Application: "reimbursement", Source: source, Kind: "task-policy",
				Registration: method + ":" + name, Operation: operation,
				Reason: "Finite dispatch: new parsing is business; frozen history export and immutable outbox delivery have separate gates"})
		}
	}
}

var reimbursementRecoveryRoutes = map[string]string{
	`group.GET("/operations/:kind") -> h.operationsQuery`:                  "READ_HISTORY",
	`group.POST("/operations/:kind/:id/retry") -> h.operationsRetry`:       "MUTATE_BUSINESS",
	`group.GET("/claims/:id/sync-status") -> h.syncStatus`:                 "READ_HISTORY",
	`group.POST("/claims/:id/spend-replays") -> h.spendReplay`:             "MUTATE_BUSINESS",
	`group.GET("/claims/:id/spend-capacity") -> h.spendCapacity`:           "READ_HISTORY",
	`group.GET("/approval-recovery-tasks") -> h.approvalRecoveryQuery`:     "READ_HISTORY",
	`group.GET("/approval-recovery-tasks/:id") -> h.approvalRecoveryQuery`: "READ_HISTORY",
}
