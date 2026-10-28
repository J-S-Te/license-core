package coverage

import (
	"fmt"
	"go/ast"
	"go/token"
)

type dispatchPolicy struct{ registration, operation, reason string }

// 仅这些原始case与方法分支经过逐项审核；未知case不能继承同名handler策略。
var reviewedDispatch20261028 = map[string][]dispatchPolicy{
	"Settlement/internal/httpapi/role_finance.go|path == \"settings\"": {
		{"GET /api/v1/settings", "READ_HISTORY", "读取租户已有规则；settings.manage，GET无写规则。"},
		{"PUT /api/v1/settings", "MUTATE_BUSINESS", "版本化保存规则并写审计。"},
	},
	"Settlement/internal/httpapi/role_finance.go|strings.HasPrefix(path, \"invoice-requests/\") && (strings.HasSuffix(path, \"/special-approve\") || strings.HasSuffix(path, \"/special-reject\"))": {
		{"POST /api/v1/invoice-requests/:id/special-approve", "MUTATE_BUSINESS", "总监加签，创建开票attempt；special_approve、合同范围、版本、申请人隔离。"},
		{"POST /api/v1/invoice-requests/:id/special-reject", "MUTATE_BUSINESS", "拒绝加签并释放预留，保持独立审核边界。"},
	},
	"Settlement/internal/httpapi/role_finance.go|path == \"receipt-differences\" || strings.HasPrefix(path, \"receipt-differences/\")": {
		{"GET /api/v1/receipt-differences", "READ_HISTORY", "读取本租户差异，manage及全范围，最多200。"},
		{"POST /api/v1/receipt-differences", "MUTATE_BUSINESS", "幂等创建差异并暂挂实际余额。"},
		{"POST /api/v1/receipt-differences/:id/resolve", "MUTATE_BUSINESS", "版本化提交差异处置。"},
		{"POST /api/v1/receipt-differences/:id/approve", "MUTATE_BUSINESS", "difference.approve与独立复核后写实际财务事实。"},
	},
	"Settlement/internal/httpapi/role_finance.go|path == \"accounting-vouchers\" || strings.HasPrefix(path, \"accounting-vouchers/\")": {
		{"GET /api/v1/accounting-vouchers", "READ_HISTORY", "读取已有凭证状态，voucher.manage及全范围，最多200。"},
		{"POST /api/v1/accounting-vouchers", "MUTATE_BUSINESS", "创建并尝试发布本地凭证，未配置/失败返回真实状态。"},
		{"POST /api/v1/accounting-vouchers/:id/retry", "MUTATE_BUSINESS", "重新尝试凭证Provider副作用并保存状态。"},
	},
	"Settlement/internal/httpapi/role_finance.go|strings.HasPrefix(path, \"receipts/\") && strings.Contains(path, \"/documents\")": {
		{"GET /api/v1/receipts/:receipt/documents", "READ_HISTORY", "receipt.read及全范围，本租户回单影像列表最多200。"},
		{"POST /api/v1/receipts/:receipt/documents", "MUTATE_BUSINESS", "receipt.record及全范围，持久意图、网关上传/绑定与归档。"},
		{"GET /api/v1/receipts/:receipt/documents/:id/download", "EXPORT_HISTORY", "本租户回款+文档定位，下载实际内容并审计，未创建新财务事实。"},
	},
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/api/v1/project-candidates\" && r.Method == http.MethodGet": {
		{"GET /api/v1/project-candidates", "MUTATE_BUSINESS", "当前commercialOperation默认MUTATE；实时权威项目候选用于业务分配，保留较严格现行边界；manage及全范围。"},
	},
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/api/v1/project-assistants\" && r.Method == http.MethodGet": {
		{"GET /api/v1/project-assistants", "MUTATE_BUSINESS", "当前commercialOperation默认MUTATE；查有效助理角色候选用于新分配，保留现行边界；manage及全范围。"},
	},
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/internal/v1/settlement/events/contract-responsibility\" && r.Method == http.MethodPost": {
		{"POST /internal/v1/settlement/events/contract-responsibility", "MUTATE_BUSINESS", "机器Bearer可信租户、事件hash与源版本校验后更改责任投影。"},
	},
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/api/v1/contract-references\" && r.Method == http.MethodGet": {
		{"GET /api/v1/contract-references", "READ_HISTORY", "receivable.read，queryScoped在分页前加入销售责任/助理协作边界，读已有有效合同。"},
	},
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/api/v1/project-assignments\" && (r.Method == http.MethodGet || r.Method == http.MethodPost)": {
		{"GET /api/v1/project-assignments", "READ_HISTORY", "collaboration.read，非全范围主体只读本人分配并分页。"},
		{"POST /api/v1/project-assignments", "MUTATE_BUSINESS", "manage及全范围，核验有效助理/权威项目引用、禁止主动自分配、版本化保存或撤销。"},
	},
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/api/v1/project-mappings\" && r.Method == http.MethodPut": {
		{"PUT /api/v1/project-mappings", "MUTATE_BUSINESS", "mapping.manage，仅本人active分配及有效权威引用，按版本改说明并审计。"},
	},
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/api/v1/receivable-plans\" && r.Method == http.MethodPost": {
		{"POST /api/v1/receivable-plans", "MUTATE_BUSINESS", "receivable.create，合同责任当前读、幂等与预算后创建待确认计划。"},
	},
	"Settlement/internal/httpapi/role_collaboration.go|strings.HasPrefix(r.URL.Path, \"/api/v1/receivable-plans/\") && strings.HasSuffix(r.URL.Path, \"/adjust\") && r.Method == http.MethodPost": {
		{"POST /api/v1/receivable-plans/:id/adjust", "MUTATE_BUSINESS", "receivable.adjust，当前责任/版本/合同预算/已核销开票及预留下限后写计划和应收。"},
	},
}

var reviewedDispatchHandlers20261028 = map[string]string{
	"Settlement/internal/httpapi/role_finance.go|path == \"settings\"": "a.roleSettings",
	"Settlement/internal/httpapi/role_finance.go|strings.HasPrefix(path, \"invoice-requests/\") && (strings.HasSuffix(path, \"/special-approve\") || strings.HasSuffix(path, \"/special-reject\"))": "a.roleSpecialReview",
	"Settlement/internal/httpapi/role_finance.go|path == \"receipt-differences\" || strings.HasPrefix(path, \"receipt-differences/\")":                                                              "a.roleDifferences",
	"Settlement/internal/httpapi/role_finance.go|path == \"accounting-vouchers\" || strings.HasPrefix(path, \"accounting-vouchers/\")":                                                              "a.roleVouchers",
	"Settlement/internal/httpapi/role_finance.go|strings.HasPrefix(path, \"receipts/\") && strings.Contains(path, \"/documents\")":                                                                  "a.roleReceiptDocuments",
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/api/v1/project-candidates\" && r.Method == http.MethodGet":                                                                  "a.projectCandidates",
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/api/v1/project-assistants\" && r.Method == http.MethodGet":                                                                  "a.projectAssistants",
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/internal/v1/settlement/events/contract-responsibility\" && r.Method == http.MethodPost":                                     "a.contractResponsibility",
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/api/v1/contract-references\" && r.Method == http.MethodGet":                                                                 "a.contractReferences",
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/api/v1/project-assignments\" && (r.Method == http.MethodGet || r.Method == http.MethodPost)":                                "a.projectAssignments",
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/api/v1/project-mappings\" && r.Method == http.MethodPut":                                                                    "a.projectMappingNote",
	"Settlement/internal/httpapi/role_collaboration.go|r.URL.Path == \"/api/v1/receivable-plans\" && r.Method == http.MethodPost":                                                                   "a.createManualPlan",
	"Settlement/internal/httpapi/role_collaboration.go|strings.HasPrefix(r.URL.Path, \"/api/v1/receivable-plans/\") && strings.HasSuffix(r.URL.Path, \"/adjust\") && r.Method == http.MethodPost":   "a.adjustPlan",
}

// settlementDispatchEntries 盘点有限的真实辅助分派函数，并保留混合方法的子策略。
// 参数：fs、file 为解析树；source 为审核来源；base 为此前注册序数。
// 返回值：原始谓词和有限方法策略，未知谓词保持未分类。
func settlementDispatchEntries(fs *token.FileSet, file *ast.File, source string, base int) []Entry {
	functions := map[string]string{"Settlement/internal/httpapi/api.go": "ServeHTTP", "Settlement/internal/httpapi/role_finance.go": "serveRoleFinance", "Settlement/internal/httpapi/role_collaboration.go": "serveRoleCollaboration"}
	name := functions[source]
	if name == "" {
		return nil
	}
	entries := []Entry{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != name {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			sw, ok := node.(*ast.SwitchStmt)
			if !ok {
				return true
			}
			if sw.Tag != nil {
				return false
			}
			for _, item := range sw.Body.List {
				clause := item.(*ast.CaseClause)
				for _, condition := range clause.List {
					base++
					e := Entry{ID: fmt.Sprintf("settlement:%s:%03d", source, base), Application: "settlement", Source: source, Kind: "http-dispatch", Registration: expression(fs, condition), Operation: "UNCLASSIFIED", Reason: "Exact dispatch predicate requires semantic review"}
					handler := ""
					for _, statement := range clause.Body {
						ast.Inspect(statement, func(n ast.Node) bool {
							call, ok := n.(*ast.CallExpr)
							if !ok {
								return true
							}
							sel, ok := call.Fun.(*ast.SelectorExpr)
							if ok && expression(fs, sel.X) == "a" {
								actual := "a." + sel.Sel.Name
								if handler != "" && handler != actual {
									handler = "<multiple>"
								} else {
									handler = actual
								}
								semanticClassify(&e, actual)
							}
							return true
						})
					}
					for _, exact := range []string{`r.URL.Path == "/healthz"`, `r.URL.Path == "/readyz"`, `r.URL.Path == "/auth/callback"`, `r.URL.Path == "/auth/logout"`, `r.URL.Path == "/auth/local-logout" && r.Method == http.MethodPost`, `r.URL.Path == "/auth/backchannel-logout" && r.Method == http.MethodPost`, `r.URL.Path == "/logged-out"`} {
						if e.Registration == exact {
							e.Operation = "ESSENTIAL_SERVICE"
							e.Reason = "Exact health/auth lifecycle predicate"
						}
					}
					key := source + "|" + e.Registration
					policies := reviewedDispatch20261028[key]
					if expected, ok := reviewedDispatchHandlers20261028[key]; ok && handler != expected {
						policies = nil
					}
					if len(policies) == 1 {
						e.Operation = policies[0].operation
						e.Reason = policies[0].reason
					}
					if len(policies) > 1 {
						e.Operation = "MIXED_EXECUTION"
						e.Reason = "Finite reviewed method/path policies below; no whole-handler exemption"
					}
					entries = append(entries, e)
					if len(policies) > 1 {
						for i, policy := range policies {
							entries = append(entries, Entry{ID: fmt.Sprintf("%s:policy:%02d", e.ID, i+1), Application: e.Application, Source: source, Kind: "task-policy", Registration: policy.registration, Operation: policy.operation, Reason: policy.reason})
						}
					}
				}
			}
			return false
		})
	}
	return entries
}

// temporalRegistrationEntries 记录带选项活动注册与结构体活动，不把循环当完整运行覆盖。
// 参数：fs、file 为解析树；app 为应用；source 为真实注册文件。
// 返回值：确定性注册表达式；未展开活动仍需独立运行比对。
func temporalRegistrationEntries(fs *token.FileSet, file *ast.File, app, source string) []Entry {
	entries := []Entry{}
	ordinal := 0
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name != "RegisterWorkflowWithOptions" && sel.Sel.Name != "RegisterActivity" && sel.Sel.Name != "RegisterActivityWithOptions" {
			return true
		}
		ordinal++
		e := Entry{ID: fmt.Sprintf("%s:%s:temporal:%03d", app, source, ordinal), Application: app, Source: source, Kind: "temporal-registration", Registration: expression(fs, call), Operation: "MUTATE_BUSINESS", Reason: "Registered workflow/activity may mutate business; execution checks and expanded runtime comparison remain required"}
		if app == "project_management" && source == "project_management/internal/background/runtime.go" {
			e.Operation = "MIXED_EXECUTION"
			e.Reason = "Reviewed finite sla/effects/notifications orchestration; concrete activity name policies and independent runtime comparison required"
		}
		entries = append(entries, e)
		return true
	})
	return entries
}

// review20261028Classify 对新项目只读入口应用精确来源约束。
// 参数：e 为待分类条目；handler 为处理器表达式。
// 返回值：匹配已审核条目时真。
func review20261028Classify(e *Entry, handler string) bool {
	if e.Source != "project_management/internal/httpapi/api.go" {
		return false
	}
	switch handler {
	case "h.settlementProjectReference", "h.projectSnapshot(snapshotLimits)", "h.listSupplementContracts":
		e.Operation = "READ_HISTORY"
		e.Reason = "Reviewed handler/service/repository existing data query only; tenant/auth scope and dependency errors remain enforced; no new business projection"
		return true
	}
	return false
}
