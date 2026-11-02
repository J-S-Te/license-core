package coverage

// 本轮新增边界已逐项核对真实持久化和副作用；不扩大到相似路径或处理器。
type reviewedSourcePolicy struct{ operation, reason string }

var reviewedEntries20261102 = map[string]reviewedSourcePolicy{
	`Settlement/cmd/worker/main.go|projectReferenceSync.RunOnce`: {
		operation: "MUTATE_BUSINESS",
		reason:    "周期拉权威引用后写本地合同/客户映射、active及updated_at水位；即使无变化或永久冲突也更新水位，非历史只读/既有不可变投递。两个ID同一执行边界，保留各自父策略关系。",
	},
	`Settlement/internal/httpapi/api.go|r.URL.Path == "/api/v1/contract-revision-holds" && r.Method == http.MethodGet`: {
		operation: "READ_HISTORY",
		reason:    "revision.read及责任谓词限定已有人工复核工单，状态/合同筛选及最多200条分页，只查询已有差异与处置结果。",
	},
	`Settlement/internal/httpapi/api.go|r.URL.Path == "/api/v1/plan-coverage" && r.Method == http.MethodGet`: {
		operation: "READ_HISTORY",
		reason:    "有具体合同则核验合同责任后读已有快照和计划合计；缺省按责任谓词列最多200差额合同。GET路径仅计算阈值，不调用auditPlanCoverageDeviation或创建/调整计划。",
	},
	`Settlement/internal/httpapi/api.go|r.URL.Path == "/api/v1/tax-command-dead-letters" && r.Method == http.MethodGet`: {
		operation: "READ_HISTORY",
		reason:    "invoice.issue权限与可信租户约束下SELECT已有税控死信、尝试和申请投影，过滤后最多200条；无重置/认领/外部税控调用。",
	},
	`Settlement/internal/httpapi/api.go|r.URL.Path == "/metrics" && r.Method == http.MethodGet`: {
		operation: "ESSENTIAL_SERVICE",
		reason:    "仅读目的地配置布尔值和未投递计数，输出运行指标，不推进业务状态。仅此精确GET谓词。",
	},
	`Settlement/internal/httpapi/api.go|strings.HasPrefix(r.URL.Path, "/api/v1/contract-revision-holds/") && strings.HasSuffix(r.URL.Path, "/resolve") && r.Method == http.MethodPost`: {
		operation: "MUTATE_BUSINESS",
		reason:    "revision.resolve专用权限与合同事务范围校验后写工单RESOLVED、人工说明和审计；属于财务复核业务状态推进，不能当技术回执豁免。",
	},
	`Settlement/internal/httpapi/api.go|strings.HasPrefix(r.URL.Path, "/api/v1/invoice-requests/") && strings.HasSuffix(r.URL.Path, "/cancel") && r.Method == http.MethodPost`: {
		operation: "MUTATE_BUSINESS",
		reason:    "以invoice.issue、租户/责任、版本/幂等及税控未受理证据为前提，将尝试ABANDONED、申请CANCELLED并释放未开票预留，新增明确财务事实。",
	},
	`Settlement/internal/httpapi/api.go|strings.HasPrefix(r.URL.Path, "/api/v1/tax-command-dead-letters/") && strings.HasSuffix(r.URL.Path, "/retry") && r.Method == http.MethodPost`: {
		operation: "MUTATE_BUSINESS",
		reason:    "持有invoice.issue权限、租户和幂等键，锁定死信/尝试/申请后将TAX_INVOICE_COMMAND重置PENDING，使开票或红冲命令可重新发送；不因重试已有命令而变必要服务。",
	},
	`contract_management/cmd/contract-number-audit/main.go|contract-number-audit`: {
		operation: "ESSENTIAL_SERVICE",
		reason:    "Read-only operational number consistency audit. Single connection READ ONLY session; SELECT-only migration auditor; no repair.",
	},
	`contract_management/internal/transport/httpapi/router.go|api.GET("/approvals/:approvalID/contract-source") -> h.downloadApprovalSource`: {
		operation: "EXPORT_HISTORY",
		reason:    "Downloads exact immutable approval snapshot after participant authorization; shared downloadSource writes persistent access audit.",
	},
	`contract_management/internal/transport/httpapi/router.go|api.GET("/contract-number-rule") -> h.getContractNumberRule`: {
		operation: "READ_HISTORY",
		reason:    "Reads global rule and annual sequence water level; sample number neither consumes sequence nor reserves registry.",
	},
	`contract_management/internal/transport/httpapi/router.go|api.GET("/contract-source-capabilities") -> h.contractSourceCapabilities`: {
		operation: "READ_HISTORY",
		reason:    "Existing PDF/batch/mode capability flags and constant size limits; permission checked; no import/upload created.",
	},
	`contract_management/internal/transport/httpapi/router.go|api.GET("/contracts/:contractID/document-descriptor") -> h.describeContractDocument`: {
		operation: "READ_HISTORY",
		reason:    "GetContract authorization and scope before document metadata/hash description; no original bytes or write.",
	},
	`contract_management/internal/transport/httpapi/router.go|api.GET("/contracts/:contractID/source") -> h.downloadContractSource`: {
		operation: "EXPORT_HISTORY",
		reason:    "Downloads authorized stored original. downloadSource records persistent access audit before bytes; mandatory technical audit does not turn export into new business.",
	},
	`contract_management/internal/transport/httpapi/router.go|api.GET("/integration-outbox/failed") -> h.listFailedIntegrationEvents`: {
		operation: "READ_HISTORY",
		reason:    "Bounded current-tenant business dead-letter listing, approval.manage required; no payload or retry.",
	},
	`contract_management/internal/transport/httpapi/router.go|api.POST("/auth/activity") -> h.activity`: {
		operation: "ESSENTIAL_SERVICE",
		reason:    "Only updates still-valid session interaction and last-seen timestamps; absolute expiry preserved.",
	},
	`contract_management/internal/transport/httpapi/router.go|api.POST("/integration-outbox/failed/:kind/:id/retry") -> h.retryFailedIntegrationEvent`: {
		operation: "MUTATE_BUSINESS",
		reason:    "Resets financial/responsibility/project/CRM business dead letters to pending and resets attempts/lease with transactional audit. Enables external business replay, unlike immutable audit delivery.",
	},
	`contract_management/internal/transport/httpapi/router.go|api.PUT("/contract-number-rule") -> h.updateContractNumberRule`: {
		operation: "MUTATE_BUSINESS",
		reason:    "Writes global single-row numbering rule affecting future issuance across tenants. Dedicated permission and business license checks remain required.",
	},
	`contract_management/internal/transport/httpapi/router.go|api.PUT("/contracts/:contractID/financial-terms") -> h.changeContractFinancialTerms`: {
		operation: "MUTATE_BUSINESS",
		reason:    "Changes approved contract amount and payment plan with row lock/version/state/origin checks; financial event, lifecycle and audit written atomically.",
	},
	`contract_management/internal/transport/httpapi/router.go|internal.GET("") -> h.contractSummary`: {
		operation: "READ_HISTORY",
		reason:    "Verified machine tenant and summary permission; SELECT existing contract, review and audit evidence; no association writes.",
	},
	`contract_management/internal/transport/httpapi/router.go|r.GET("/metrics") -> h.integrationOutboxMetrics`: {
		operation: "ESSENTIAL_SERVICE",
		reason:    "Deployment-level fixed-kind dead-letter COUNT metrics. No tenant payload, claim, retry, or event production.",
	},
	`customer_and_opportunity/internal/bootstrap/app.go|api.POST("/auth/activity") -> authHandler.Activity`: {
		operation: "ESSENTIAL_SERVICE",
		reason:    "Activity仅推进已认证本地会话的last_seen_at与last_interactive_at，不创建业务记录、不改变绝对到期；普通业务认证和Origin/CSRF仍保留。",
	},
	`customer_and_opportunity/internal/portalbootstrap/router.go|internal.GET("/report-outbox/failed") -> listReportDeliveryFailures(deps)`: {
		operation: "READ_HISTORY",
		reason:    "仅按机器主体验签租户分页读取既有PORTAL_REPORT_SUBMITTED死信；Count/Scan没有状态回写或新投递。",
	},
	`customer_and_opportunity/internal/portalbootstrap/router.go|internal.POST("/report-outbox/failed/:id/retry") -> retryReportDeliveryFailure(deps)`: {
		operation: "MUTATE_BUSINESS",
		reason:    "把报告提交死信改为RETRY_WAIT并清空重试/租约字段，之后worker会向项目系统POST报告申请并推进审批；并非单纯既有通知或审计投递。",
	},
	`data_analysis/internal/bootstrap/app.go|api.GET("/dashboard/contract-import-history") -> importreport.HistoryHandler(db)`: {
		operation: "READ_HISTORY",
		reason:    "History在只读RR事务读取已发布contract_import_history和run状态，按可信主体范围做SQL聚合；不触发同步、不保存派生结果。",
	},
	`data_analysis/internal/bootstrap/app.go|api.POST("/auth/activity") -> authHandler.Activity`: {
		operation: "ESSENTIAL_SERVICE",
		reason:    "TouchActivity只维护已存在本地身份会话时间戳，校验撤销/绝对/空闲有效期；不生成聚合、告警或业务事件。",
	},
	`data_analysis/internal/bootstrap/app.go|router.GET(InternalReportStatusPath) -> status`: {
		operation: "READ_HISTORY",
		reason:    "精确版本已应用证明查询；附带bi_audit_outbox审计插入是技术性审计副作用，不重新发布报表或启动聚合。不能说它完全无持久化写入。",
	},
	`project_management/cmd/evidence-ledger-inventory/main.go|evidence-ledger-inventory`: {
		operation: "ESSENTIAL_SERVICE",
		reason:    "Read-only operational ledger inventory. Rejects --apply; single-connection READ ONLY session, two SELECT queries and in-memory aggregation; no migration/backfill.",
	},
	`project_management/internal/httpapi/api.go|api.GET("/approved-contract-options") -> h.searchApprovedContractOptions`: {
		operation: "READ_HISTORY",
		reason:    "Authorized project.create selector searches existing approved contracts by keyword/cursor and reads project references for exclusion; no activation.",
	},
	`project_management/internal/httpapi/api.go|api.GET("/service-items/:id/field-evidence") -> h.getServiceItemFieldEvidence`: {
		operation: "READ_HISTORY",
		reason:    "project.read/resource-scope checks, bounded complete service-item event SELECT and in-memory gap computation; overflow fails closed, no completion or backfill.",
	},
	`project_management/internal/httpapi/api.go|api.POST("/auth/activity") -> h.activity`: {
		operation: "ESSENTIAL_SERVICE",
		reason:    "Conditional update of valid session interaction/last-seen only; revoked, idle-expired or absolutely expired sessions cannot revive.",
	},
	`project_management/internal/httpapi/api.go|api.POST("/contracts/activate") -> h.activateContractFromUser`: {
		operation: "MUTATE_BUSINESS",
		reason:    "Re-reads authoritative contract and service catalog, reconstructs trusted fields, then activateContract persists project/service items or stamped-state synchronization.",
	},
	`project_management/internal/httpapi/api.go|api.POST("/projects/:id/supplement-closure") -> h.closeSupplement`: {
		operation: "MUTATE_BUSINESS",
		reason:    "Checks decomposition permission, REQUIRED state, reason and same-customer approved supplement; applyEvent persists SUPPLEMENT_CLOSED and accepted scope baseline.",
	},
	`project_management/internal/httpapi/api.go|imports.GET("/capabilities") -> <anonymous>`: {
		operation: "READ_HISTORY",
		reason:    "Anonymous route only calls contractImportCapabilities, returns static supported protocol/modes and current flag under machine authentication.",
	},
	`project_management/internal/httpapi/api.go|imports.GET("/project-bindings/:origin_import_item_id") -> h.getContractImportBinding`: {
		operation: "READ_HISTORY",
		reason:    "Reads persistent source-item binding receipt by verified tenant; SELECT and JSON decode only, no create or repair.",
	},
	`project_management/internal/httpapi/api.go|imports.POST("/project-bindings") -> <anonymous>`: {
		operation: "MUTATE_BUSINESS",
		reason:    "Anonymous route calls BindImportedContract. Exact replay may read, but first command atomically creates/links project, stores binding and receipt; classify maximum real side effect.",
	},
}

// applyReviewedEntries20261102 应用本轮逐项审核的精确来源与注册策略。
// 参数：r 为当前真实源码扫描清单。
// 返回值：无；不匹配的新增来源、注册或任务保持原来的未分类状态。
func applyReviewedEntries20261102(r *Registry) {
	for i := range r.Entries {
		entry := &r.Entries[i]
		if policy, ok := reviewedEntries20261102[entry.Source+"|"+entry.Registration]; ok {
			entry.Operation, entry.Reason = policy.operation, policy.reason
		}
	}
}

// 新增实际实现闭包只扩展摘要覆盖，不授予运行许可或登记豁免。
var reviewedExecutionSources20261102 = []string{
	"contract_management/internal/infrastructure/mysql/import_confirmation.go",
	"Settlement/internal/businessclock/businessclock.go",
	"Settlement/internal/config/config.go",
	"Settlement/internal/config/project_reference_sync.go",
	"Settlement/internal/dunning/active_case_guard.go",
	"Settlement/internal/dunning/overlap.go",
	"Settlement/internal/httpapi/contract_revision_hold.go",
	"Settlement/internal/httpapi/integration_status.go",
	"Settlement/internal/httpapi/invoice_cancel.go",
	"Settlement/internal/outbox/import_gate.go",
	"Settlement/internal/platform/client_credentials_token.go",
	"Settlement/internal/platform/platform_application_token.go",
	"Settlement/internal/service/contract_revision.go",
	"Settlement/internal/service/finance_records.go",
	"Settlement/internal/service/invoice_cancel.go",
	"Settlement/internal/service/plan_coverage.go",
	"Settlement/internal/service/project_reference_sync.go",
	"Settlement/internal/service/refund_confirmation.go",
	"Settlement/internal/service/tax_callback.go",
	"Settlement/internal/service/tax_deadletter.go",
	"contract_management/cmd/file-backfill/source_descriptor.go",
	"contract_management/internal/application/contract_financial_change.go",
	"contract_management/internal/application/contract_number.go",
	"contract_management/internal/application/contract_summary.go",
	"contract_management/internal/application/integration_outbox.go",
	"contract_management/internal/application/service.go",
	"contract_management/internal/application/templates.go",
	"contract_management/internal/businessclock/businessclock.go",
	"contract_management/internal/domain/contract/number_rule.go",
	"contract_management/internal/infrastructure/mysql/contract_number.go",
	"contract_management/internal/infrastructure/mysql/contract_summary.go",
	"contract_management/internal/infrastructure/mysql/draft.go",
	"contract_management/internal/infrastructure/mysql/integration_outbox.go",
	"contract_management/internal/infrastructure/mysql/number_audit.go",
	"contract_management/internal/infrastructure/mysql/opportunity_link_guard.go",
	"contract_management/internal/infrastructure/mysql/payment_terms.go",
	"contract_management/internal/infrastructure/platform/oidc.go",
	"contract_management/internal/infrastructure/platform/oidc_store.go",
	"contract_management/internal/integration/crm/link.go",
	"contract_management/internal/migration/preflight.go",
	"contract_management/internal/transport/httpapi/contract_document.go",
	"contract_management/internal/transport/httpapi/contract_financial_change.go",
	"contract_management/internal/transport/httpapi/contract_number.go",
	"contract_management/internal/transport/httpapi/contract_summary.go",
	"contract_management/internal/transport/httpapi/integration_outbox.go",
	"contract_management/internal/transport/httpapi/metrics.go",
	"contract_management/internal/workflows/number_audit.go",
	"customer_and_opportunity/internal/modules/crmauth/handler.go",
	"customer_and_opportunity/internal/modules/crmauth/model.go",
	"customer_and_opportunity/internal/modules/crmauth/repository.go",
	"customer_and_opportunity/internal/modules/crmauth/service.go",
	"customer_and_opportunity/internal/modules/portal/report/model.go",
	"customer_and_opportunity/internal/modules/portal/report/outbox_ops.go",
	"customer_and_opportunity/internal/modules/portal/report/repository.go",
	"customer_and_opportunity/internal/modules/portal/report/service.go",
	"customer_and_opportunity/internal/portalreportworker/http_port.go",
	"customer_and_opportunity/internal/portalreportworker/store.go",
	"customer_and_opportunity/internal/portalreportworker/worker.go",
	"customer_and_opportunity/internal/shared/handoverguard/guard.go",
	"data_analysis/internal/aggregation/contract_approval.go",
	"data_analysis/internal/aggregation/import_report.go",
	"data_analysis/internal/aggregation/source_lock.go",
	"data_analysis/internal/aggregation/sync_run.go",
	"data_analysis/internal/auditoutbox/append.go",
	"data_analysis/internal/importreport/http.go",
	"data_analysis/internal/importreport/verifier.go",
	"data_analysis/internal/oidc/handler.go",
	"data_analysis/internal/oidc/model.go",
	"data_analysis/internal/oidc/service.go",
	"data_analysis/internal/shared/scope/scope.go",
	"project_management/internal/application/contract_import.go",
	"project_management/internal/application/field_checkin.go",
	"project_management/internal/application/service.go",
	"project_management/internal/application/supplement_exit.go",
	"project_management/internal/businessclock/businessclock.go",
	"project_management/internal/httpapi/contract_import.go",
	"project_management/internal/infrastructure/mysql/contract_import.go",
	"project_management/internal/infrastructure/mysql/delivery.go",
	"project_management/internal/infrastructure/mysql/delivery_events.go",
	"project_management/internal/infrastructure/mysql/delivery_queries.go",
	"project_management/internal/infrastructure/mysql/delivery_report.go",
	"project_management/internal/infrastructure/mysql/supplement_exit.go",
	"project_management/internal/platform/file_gateway.go",
	"project_management/internal/platform/oidc.go",
	"project_management/internal/platform/oidc_store.go",
}
