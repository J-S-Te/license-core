package coverage

// 本表来自逐项调用链审核；精确来源和表达式不匹配时保持原分类。
var reviewedEntries20261110 = map[string]reviewedSourcePolicy{
	"reimbursement/internal/bootstrap/runtime.go|health.HandleFunc(\"/readyz\") -> workerReadinessHandler(runtime)":     {operation: "ESSENTIAL_SERVICE", reason: "Readiness only pings DB and evaluates license/configuration presence; no remote business work or queue claim."},
	"reimbursement/internal/bootstrap/runtime.go|health.HandleFunc(\"/metrics\") -> metricsHandler(runtime)":            {operation: "ESSENTIAL_SERVICE", reason: "Technical SELECT COUNT/MIN over queue statuses with finite labels; no business identifiers, claim, retry or mutation. observability/metrics.go and bootstrap/diagnostics.go reviewed."},
	"reimbursement/internal/transport/httpapi/router.go|group.GET(\"/archives/:id/packages\") -> h.archivePackages":     {operation: "READ_HISTORY", reason: "ArchivePackages rechecks current whole-claim authority and returns owned archive-package job metadata/digests only; does not read blob contents or queue generation."},
	"reimbursement/internal/transport/httpapi/router.go|group.GET(\"/invoices/:id/controls\") -> h.invoiceControls":     {operation: "READ_HISTORY", reason: "QueryInvoiceControls verifies current invoice visibility then reads bounded correction/risk/source/claim impact records; no decision, source append or financial state change."},
	"reimbursement/internal/transport/httpapi/router.go|group.GET(\"/claims/:id/approval-rounds\") -> h.approvalRounds": {operation: "READ_HISTORY", reason: "ApprovalRounds only COUNT/SELECT existing frozen rounds, tasks and actions under current claim/task scope; no reassignment or new approval. approval_rounds.go reviewed after in-flight writer completed."},
	"reimbursement/internal/transport/httpapi/router.go|group.POST(\"/invoices/:id/corrections\") -> h.command(\"invoice.correction.request\")": {
		operation: "MUTATE_BUSINESS",
		reason:    "Creates invoice correction request and control impact under tenant/version checks; invoice_corrections_write.go requestInvoiceCorrection.",
	},
	"reimbursement/internal/transport/httpapi/router.go|group.POST(\"/invoice-corrections/:id/decisions\") -> h.command(\"invoice.correction.decide\")": {
		operation: "MUTATE_BUSINESS",
		reason:    "Independent finance decision changes correction, invoice and impacted claim state; decideInvoiceCorrection.",
	},
	"reimbursement/internal/transport/httpapi/router.go|group.POST(\"/claims/:id/invoice-reconfirmations\") -> h.command(\"claim.invoice.reconfirm\")": {
		operation: "MUTATE_BUSINESS",
		reason:    "Applicant reconfirms changed invoice facts and advances impact/reapproval state; reconfirmInvoiceCorrection.",
	},
	"reimbursement/internal/transport/httpapi/router.go|group.POST(\"/invoices/:id/sources\") -> h.command(\"invoice.source.append\")": {
		operation: "MUTATE_BUSINESS",
		reason:    "Appends controlled source/document relations and immutable command audit; invoice_sources.go appendControlledInvoiceSources.",
	},
	"reimbursement/internal/transport/httpapi/router.go|group.POST(\"/invoices/:id/risks\") -> h.command(\"invoice.risk.record\")": {
		operation: "MUTATE_BUSINESS",
		reason:    "Creates invoice risk and affected claim control records; invoice_risks.go recordInvoiceRisk.",
	},
	"reimbursement/internal/transport/httpapi/router.go|group.POST(\"/invoice-risks/:id/resolutions\") -> h.command(\"invoice.risk.close\")": {
		operation: "MUTATE_BUSINESS",
		reason:    "Independent review changes risk resolution and control state; closeInvoiceRisk.",
	},
	"reimbursement/internal/transport/httpapi/router.go|group.POST(\"/archives/:id/packages\") -> h.archivePackage": {
		operation: "EXPORT_HISTORY",
		reason:    "Queues an export from a SEALED version with fixed manifest hash; no financial changes. Current POST runtime mapping remains stricter MUTATE_BUSINESS.",
	},
	"reimbursement/internal/transport/httpapi/router.go|group.GET(\"/archive-packages/:id\") -> h.archivePackageStatus": {
		operation: "READ_HISTORY",
		reason:    "Rechecks current whole-claim authority and reads owned package job status/digest only; ArchivePackageStatus.",
	},
	"reimbursement/internal/transport/httpapi/router.go|group.POST(\"/archive-packages/:id/download\") -> h.archiveDownload": {
		operation: "EXPORT_HISTORY",
		reason:    "Rechecks current whole-claim authority and artifact hash, downloads historical sealed ZIP and appends access audit only; ReadArchivePackage.",
	},
	"reimbursement/internal/jobs/processor.go|Job:ARCHIVE_PACKAGE": {
		operation: "EXPORT_HISTORY",
		reason:    "Processor.Job checks EXPORT_HISTORY before archivePackage; binds sealed manifest/hash, emits bounded ZIP and persists lease-bound export result/audit without financial mutation.",
	},
	"Settlement/internal/httpapi/api.go|r.URL.Path == \"/auth/activity\" && r.Method == http.MethodPost": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "api.go:251→Authenticator.Activity(platform/session_governance_http.go:82)→lifecycle:87→readCentral(interact=true)→CentralClient.Activity(sessiongovernance/client.go:157)；精确Origin/CSRF后记录真实会话活动，技术写入不改变应收/收款/发票。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"Settlement/internal/httpapi/api.go|r.URL.Path == \"/auth/session\" && r.Method == http.MethodGet": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "api.go:244→Authenticator.SessionState(platform/session_governance_http.go:75)→lifecycle:87→Authenticate/loadCentralSession(platform/session_governance.go:34)/readCentral:55→CentralClient.State(sessiongovernance/client.go:131)；只维护认证与中央会话，不写财务台账。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"Settlement/internal/httpapi/contract_import.go|POST /api/v1/settlement/contract-imports/:id/reconcile": {
		operation: "MUTATE_BUSINESS",
		reason:    "serveContractImports:20→service.ReconcileContractImport(import_reconciliation.go:119)创建对账记录并更新READY、信用决定及审批主体。",
	},
	"Settlement/internal/httpapi/contract_import.go|POST /internal/v1/settlement/contract-imports/:id/release": {
		operation: "MUTATE_BUSINESS",
		reason:    "serveContractImports:39→service.ReleaseContractImport(import_reconciliation.go:167)放行财务基线、导入及信用投递状态。",
	},
	"Settlement/internal/httpapi/contract_import.go|POST /internal/v1/settlement/events/contract-imports": {
		operation: "MUTATE_BUSINESS",
		reason:    "serveContractImports:38→service.PrepareContractImport(contract_import.go:111)经ingestContract事务创建HOLD基线、计划与导入记录。",
	},
	"Settlement/internal/httpapi/contract_import_export.go|GET /api/v1/settlement/contract-imports/:id/exports/:exportID": {
		operation: "READ_HISTORY",
		reason:    "serveImportExports:52→service.GetImportExport(import_export.go:250)读取历史导出状态及内容，过期清理仅导出技术数据。",
	},
	"Settlement/internal/httpapi/contract_import_export.go|GET /api/v1/settlement/contract-imports/:id/exports/:exportID/download": {
		operation: "EXPORT_HISTORY",
		reason:    "serveImportExports:52/63经同租户既有导出查询后只传输已成功CSV内容，不创建财务业务。",
	},
	"Settlement/internal/httpapi/contract_import_export.go|POST /api/v1/settlement/contract-imports/:id/exports": {
		operation: "EXPORT_HISTORY",
		reason:    "serveImportExports:29→service.CreateImportExport(import_export.go:46)/buildImportExport:121只读已有恢复映射生成CSV，必要写导出任务/内容/状态，不推进财务对账或放行。",
	},
	"Settlement/internal/httpapi/contract_import_query.go|GET /api/v1/settlement/contract-imports": {
		operation: "READ_HISTORY",
		reason:    "serveImportQueries:14在用户权限及全租户范围校验后分页读取已有导入状态，无恢复或放行写入。",
	},
	"Settlement/internal/httpapi/contract_import_query.go|GET /api/v1/settlement/contract-imports/:id": {
		operation: "READ_HISTORY",
		reason:    "serveImportQueries→importStatus:158读取既有基线、计划、责任及信用决定版本，不保存派生状态。",
	},
	"Settlement/internal/httpapi/contract_import_query.go|GET /api/v1/settlement/contract-imports/:id/facts": {
		operation: "READ_HISTORY",
		reason:    "serveImportQueries→importFacts:108分页查询同租户已恢复事实及目标业务ID映射。",
	},
	"Settlement/internal/httpapi/contract_import_query.go|GET /internal/v1/settlement/contract-imports/:id": {
		operation: "READ_HISTORY",
		reason:    "serveImportQueries→importStatus:158仅在可信机器租户读取既有导入、计划及责任ACK。",
	},
	"Settlement/internal/httpapi/contract_import_query.go|GET /internal/v1/settlement/contract-imports/capabilities": {
		operation: "READ_HISTORY",
		reason:    "serveImportQueries:53在机器身份可信租户下只输出既有开关和协议能力，不执行PREPARE/RELEASE。",
	},
	"Settlement/internal/httpapi/native_financial.go|GET /api/v1/financial-baselines": {
		operation: "READ_HISTORY",
		reason:    "nativeFinancialBaselines:107→service.ListNativeFinancialBaselines(native_financial_query.go:89)分页读取既有本地财务依据。",
	},
	"Settlement/internal/httpapi/native_financial.go|GET /api/v1/financial-baselines/:id": {
		operation: "READ_HISTORY",
		reason:    "serveNativeFinancial:44→service.GetNativeFinancialBaseline(native_financial_query.go:68)读取同租户既有依据。",
	},
	"Settlement/internal/httpapi/native_financial.go|GET /api/v1/financial-baselines/capabilities": {
		operation: "READ_HISTORY",
		reason:    "serveNativeFinancial:19读取开关和CanCreateNativeFinancial权限能力，不创建财务依据。",
	},
	"Settlement/internal/httpapi/native_financial.go|GET /api/v1/financial-customers": {
		operation: "READ_HISTORY",
		reason:    "nativeFinancialCustomers:74→service.ListNativeFinancialCustomers(native_financial_query.go:24)只查询同租户本地客户。",
	},
	"Settlement/internal/httpapi/native_financial.go|POST /api/v1/financial-baselines": {
		operation: "MUTATE_BUSINESS",
		reason:    "nativeFinancialBaselines:107→service.CreateNativeFinancialBaseline(native_financial.go:225)事务写本地依据、待确认计划、幂等和审计。",
	},
	"Settlement/internal/httpapi/native_financial.go|POST /api/v1/financial-baselines/:id/void": {
		operation: "MUTATE_BUSINESS",
		reason:    "serveNativeFinancial:53→service.VoidNativeFinancialBaseline(native_financial_query.go:115)作废依据并取消待确认计划，改变业务状态。",
	},
	"Settlement/internal/httpapi/native_financial.go|POST /api/v1/financial-customers": {
		operation: "MUTATE_BUSINESS",
		reason:    "nativeFinancialCustomers:74→service.CreateNativeFinancialCustomer(native_financial.go:132)事务创建本地客户、幂等及审计。",
	},
	"contract_management/internal/temporalworker/metrics.go|mux.Handle(\"GET /metrics\") -> registry": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "metrics.go:159→MetricsRegistry.ServeHTTP:122仅读锁保护的Temporal内存计数/耗时并渲染文本，不写合同或启动工作流。",
	},
	"contract_management/internal/temporalworker/metrics.go|mux.HandleFunc(\"GET /readyz\") -> <anonymous>": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "metrics.go:160只调用ready()原子状态回调返回200/503，不认领任务或改变审批/许可状态。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.GET(\"/contract-import-batches\") -> h.listImportBatches": {
		operation: "READ_HISTORY",
		reason:    "contract_import.go:27→import_confirmation.go:80：limit<=100/offset<=10000、真实范围内SELECT候选再逐行复核，无认领。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.GET(\"/contract-import-batches/:id\") -> h.getImportBatch": {
		operation: "READ_HISTORY",
		reason:    "contract_import.go:26→application/contract_import_preview.go:96：按租户/创建人或真实受派人读取已有台账，不推进状态。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.GET(\"/contract-import-batches/:id/items/:itemID/document-descriptor\") -> h.describeImportDocument": {
		operation: "READ_HISTORY",
		reason:    "contract_import.go:28→import_confirmation.go:172→application/import_document.go:14：核冻结SHA并持久访问审计后返回描述，无业务状态推进。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.GET(\"/contract-import-batches/:id/items/:itemID/source\") -> h.downloadImportDocument": {
		operation: "EXPORT_HISTORY",
		reason:    "contract_import.go:29→import_confirmation.go:184→GetImportSourceDocument：核验已有冻结原件SHA并先持久访问审计，然后输出原字节。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.GET(\"/contract-import/auto-approve\") -> h.getImportAutoApproveSetting": {
		operation: "READ_HISTORY",
		reason:    "contract_import.go:40→application/contract_import_auto_approve.go:38与mysql/contract_import_auto_approve_setting.go:89：仅admin读取租户设置及能力，无默认插入。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.POST(\"/contract-import-batches\") -> h.createImportBatch": {
		operation: "MUTATE_BUSINESS",
		reason:    "contract_import.go:25→application/contract_import_preview.go:51：许可及权限后持久创建DRAFT台账和TTL，是新合同导入起点。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.POST(\"/contract-import-batches/:id/commit\") -> h.commitImportBatch": {
		operation: "MUTATE_BUSINESS",
		reason:    "contract_import.go:36→application/contract_import_commit.go:13：持久操作/会话引用、升epoch并入QUEUED；后续创建合同/原件绑定/CRM开立。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.POST(\"/contract-import-batches/:id/confirmation-actions\") -> h.actImportConfirmation": {
		operation: "MUTATE_BUSINESS",
		reason:    "contract_import.go:31→application/contract_import_confirmation.go:268：核许可后保存审批命令；AUTO撤回直接写WITHDRAWN/PAUSED并升代。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.POST(\"/contract-import-batches/:id/files\") -> h.uploadImportSource": {
		operation: "MUTATE_BUSINESS",
		reason:    "contract_import.go:33→application/contract_import_preview.go:156：检查业务许可，上传新原件、绑定暂存引用并更新代次/清单版本。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.POST(\"/contract-import-batches/:id/pause\") -> h.controlImportBatch": {
		operation: "MUTATE_BUSINESS",
		reason:    "contract_import.go:37→application/contract_import_commit.go:83：持久PAUSED、升epoch、清租约，影响导入业务调度，非身份安全撤销。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.POST(\"/contract-import-batches/:id/preview\") -> h.previewImportBatch": {
		operation: "MUTATE_BUSINESS",
		reason:    "contract_import.go:35→application/contract_import_preview.go:505、603-621：核业务许可，持久保存NEEDS_PROVISION计划、逐行结果、状态、令牌和到期。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.POST(\"/contract-import-batches/:id/resume\") -> h.controlImportBatch": {
		operation: "MUTATE_BUSINESS",
		reason:    "contract_import.go:38→ControlImport：原创建人及会话验证后写QUEUED及可重试项READY，使新合同/CRM开立恢复。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.POST(\"/contract-import-batches/:id/submit-confirmation\") -> h.submitImportConfirmation": {
		operation: "MUTATE_BUSINESS",
		reason:    "contract_import.go:30→application/contract_import_confirmation.go:102：核业务许可，冻结AUTO/MANUAL及确认台账/QUEUED，人工分支启动Temporal。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.POST(\"/contract-import-items/:id/release\") -> h.releaseImportItem": {
		operation: "MUTATE_BUSINESS",
		reason:    "contract_import.go:32→application/contract_import_targets.go:28：许可和放行权限后排项目/财务/报表目标QUEUED/命令/围栏，驱动新业务。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.PUT(\"/contract-import-batches/:id/manifest\") -> h.setImportManifest": {
		operation: "MUTATE_BUSINESS",
		reason:    "contract_import.go:34→application/contract_import_preview.go:269：持久替换单项/清单，升revision/epoch、失效预览令牌或写纠错状态。",
	},
	"contract_management/internal/transport/httpapi/contract_import.go|api.PUT(\"/contract-import/auto-approve\") -> h.putImportAutoApproveSetting": {
		operation: "MUTATE_BUSINESS",
		reason:    "contract_import.go:41→application/contract_import_auto_approve.go:63→mysql设置:111：admin与能力校验后事务修改自动批准设置/版本及审计。",
	},
	"contract_management/internal/transport/httpapi/import_confirmation.go|group.GET(\"/report-source\") -> <anonymous>": {
		operation: "READ_HISTORY",
		reason:    "import_confirmation.go:145/152独立机器scope/caller/tenant/env/audience认证，mysql/import_report.go:15在只读RR事务读取已放行最多1000完整历史行并内存摘要。",
	},
	"contract_management/internal/transport/httpapi/local_catalog.go|api.GET(\"/local-catalog/categories\") -> h.listLocalCategories": {
		operation: "READ_HISTORY",
		reason:    "读取可信租户/数据范围内已有类别，application/local_catalog.go:114与mysql/local_catalog.go:169仅Find；无目录修改。",
	},
	"contract_management/internal/transport/httpapi/local_catalog.go|api.GET(\"/local-catalog/counterparties\") -> h.listLocalCounterparties": {
		operation: "READ_HISTORY",
		reason:    "读取可信租户/数据范围内已有交易对方，application/local_catalog.go:64与mysql/local_catalog.go:89仅Find；非管理员按有效目录授权过滤。",
	},
	"contract_management/internal/transport/httpapi/local_catalog.go|api.POST(\"/local-catalog/categories\") -> h.saveLocalCategory": {
		operation: "MUTATE_BUSINESS",
		reason:    "新建类别及版本、审计，application/local_catalog.go:142先checkBusinessLicense，mysql/local_catalog.go:204事务写入。",
	},
	"contract_management/internal/transport/httpapi/local_catalog.go|api.POST(\"/local-catalog/counterparties\") -> h.saveLocalCounterparty": {
		operation: "MUTATE_BUSINESS",
		reason:    "保存/新建本地交易对方与版本、审计，application/local_catalog.go:92先checkBusinessLicense，mysql/local_catalog.go:124事务写入。",
	},
	"contract_management/internal/transport/httpapi/local_catalog.go|api.PUT(\"/local-catalog/categories/:entryID\") -> h.saveLocalCategory": {
		operation: "MUTATE_BUSINESS",
		reason:    "更新/停用类别与版本、审计，稳定编码不可替换；沿用saveLocalCategory的业务许可和管理范围。",
	},
	"contract_management/internal/transport/httpapi/local_catalog.go|api.PUT(\"/local-catalog/counterparties/:entryID\") -> h.saveLocalCounterparty": {
		operation: "MUTATE_BUSINESS",
		reason:    "更新/停用已有交易对方与版本、审计，沿用专有管理权限及预期版本；同saveLocalCounterparty链路。",
	},
	"contract_management/internal/transport/httpapi/local_catalog_grants.go|api.GET(\"/local-catalog/\" + kind + \"/:entryID/grants\") -> <anonymous>": {
		operation: "READ_HISTORY",
		reason:    "仅封闭kind={counterparties,categories}循环，:16→listCatalogGrants→application/local_catalog_grants.go:70；核验当前条目管理范围后仅读授权历史含撤销版本。kind集合需指纹绑定。",
	},
	"contract_management/internal/transport/httpapi/local_catalog_grants.go|api.GET(\"/local-catalog/grant-targets\") -> h.localCatalogTargets": {
		operation: "READ_HISTORY",
		reason:    "local_catalog_grants.go:14→application/local_catalog_grants.go:57，在contract.catalog.manage与可信租户下读取平台当前有效销售/组织候选；不授予权限。",
	},
	"contract_management/internal/transport/httpapi/local_catalog_grants.go|api.POST(\"/local-catalog/\" + kind + \"/:entryID/grants\") -> <anonymous>": {
		operation: "MUTATE_BUSINESS",
		reason:    "仅封闭kind={counterparties,categories}循环，:17→saveCatalogGrant→application/local_catalog_grants.go:81，checkBusinessLicense后校验真实目标并事务修改可见性授权版本及审计（mysql/local_catalog_grants.go:79）；撤销亦不整体豁免。",
	},
	"contract_management/internal/transport/httpapi/router.go|api.PUT(\"/contract-templates/:templateID/source\") -> h.replaceTemplateSource": {
		operation: "MUTATE_BUSINESS",
		reason:    "router.go:351→templates_source.go:15管理权限与有界上传→application/templates.go:94核业务许可后替换未来草稿模板源及文件代次。",
	},
	"contract_management/internal/transport/httpapi/router.go|r.GET(\"/auth/session\") -> <anonymous>": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "router.go:174实际调用OIDCAuthenticator.SessionState→session_governance.go:167/181 lifecycle(false)，验证Cookie身份/中央绑定后仅读非敏感窗口；认证刷新、失效撤销属于身份生命周期，无合同业务写。",
	},
	"contract_management/internal/transport/httpapi/router.go|r.POST(\"/auth/activity\") -> <anonymous>": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "router.go:175实际调用OIDCAuthenticator.Activity→session_governance.go:174/181 lifecycle(true)，认证与精确Origin/CSRF后仅代理中央活动窗口；中央client.go:157调用/oauth2/session/activity，无合同审批、授权授予或业务写。",
	},
	"customer_and_opportunity/internal/bootstrap/app.go|api.GET(\"/capabilities\") -> runtimeCapabilitiesHandler(config, workerReadiness, config.PresaleWorkerHeartbeatMaxAge)": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "只投影进程配置能力，不调用业务状态推进；bootstrap/runtime_capabilities.go:70→:66 resolveRuntimeCapabilities→:38 runtimeCapabilities，仅构造内存map。",
	},
	"customer_and_opportunity/internal/bootstrap/app.go|api.GET(\"/opportunities/:id/presale-requests\") -> opportunityPresales.List": {
		operation: "READ_HISTORY",
		reason:    "bootstrap/presale_opportunity_handler.go:23先opportunity.Get校验商机可见性，Resolve现有身份后presale.ListForOpportunity分页查询售前记录，无新申请/持久业务写。",
	},
	"customer_and_opportunity/internal/bootstrap/app.go|base.GET(\"/auth/session\") -> authHandler.SessionState": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "当前Cookie认证后仅代理中央状态；crmauth/governance_http.go:25/35→governance.go:102 Lifecycle/:64 centralState→shared/sessiongovernance/client.go:131 State(GET /oauth2/session-context)。认证可能刷新令牌/记录last_seen或撤销失效会话，均限身份生命周期。",
	},
	"customer_and_opportunity/internal/bootstrap/app.go|base.GET(\"/readyz\") -> readinessHandler(config, databaseHealthy, auditOutbox)": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "就绪处理仅Ping数据库、返回能力配置并读取审计积压/挂起计数，不触发业务；bootstrap/app.go:106–115→health.go:23→audit_outbox_readiness.go:58 snapshot；探针审计属于技术写。",
	},
	"customer_and_opportunity/internal/bootstrap/app.go|base.POST(\"/auth/activity\") -> authHandler.CentralActivity": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "身份生命周期续期：同源/CSRF保护后的真实交互代理，不能复活到期会话或授予业务权限；bootstrap/app.go:171→governance_http.go:30→governance.go:102/:64→shared/sessiongovernance/client.go:157 Activity(POST /oauth2/session/activity)。",
	},
	"customer_and_opportunity/internal/modules/importprovision/routes.go|router.POST(\"/import-provision/customers\") -> handler.CreateCustomer": {
		operation: "MUTATE_BUSINESS",
		reason:    "新建客户/联系人及编号、幂等记录和来源审计；importprovision/handler.go:38→service.go:68→customer/provision.go:139 CreateProvision/:390 persistProvisionCustomer→customer/repository.go:325 Create；平台负责人目录校验不是业务写豁免。",
	},
	"customer_and_opportunity/internal/modules/importprovision/routes.go|router.POST(\"/import-provision/customers/lookup\") -> handler.LookupCustomer": {
		operation: "READ_HISTORY",
		reason:    "只查询同租户归一化精确客户候选，额外写判定审计，不建客户/商机或推进导入；handler.go:18→importprovision/service.go:49→customer/provision.go:97 LookupProvisionCandidates→:472 FindProvisionCandidates(SELECT Limit21)；语义READ_HISTORY，现行commercial.Operation仍默认MUTATE_BUSINESS，本审核不授权放宽运行门禁。",
	},
	"customer_and_opportunity/internal/modules/importprovision/routes.go|router.POST(\"/import-provision/opportunities\") -> handler.CreateOpportunity": {
		operation: "MUTATE_BUSINESS",
		reason:    "创建新商机并写幂等记录和来源审计；importprovision/handler.go:63→service.go:90→opportunity/provision.go:101 CreateProvision(客户范围、负责人核验后事务repo.Create:191/CreateCreateIdempotency:200)。",
	},
	"customer_and_opportunity/internal/modules/importprovision/routes.go|router.POST(\"/import-provision/opportunities/:id/sign\") -> handler.SignOpportunity": {
		operation: "MUTATE_BUSINESS",
		reason:    "推进既有商机到已签约/CLOSED并持久化阶段/日志和来源审计；importprovision/handler.go:89→service.go:111→opportunity/provision.go:233 SignProvision，:271远端BelongsToCustomer验真、:283 applyStage、:292 persistStage→service.go:1584事务UpdateStage/CreateStageLog；幂等回放分支不改变整体写分类。",
	},
	"customer_and_opportunity/internal/modules/opportunity/routes.go|opportunities.GET(\"/:id/win-confirmation\") -> handler.WinConfirmation": {
		operation: "READ_HISTORY",
		reason:    "先校验商机读取权限/数据范围，再按tenant+opportunity读取历史赢单证据；opportunity/win_confirmation_handler.go:11→win_confirmation.go:222 GetWinConfirmation→repo.FindByID/:86 store.Get(SELECT)，无Create/状态推进/外部写。",
	},
	"customer_and_opportunity/internal/modules/opportunity/routes.go|router.POST(\"/integrations/opportunity-attachment/scan-events\") -> handler.ApplyAttachmentScan": {
		operation: "MUTATE_BUSINESS",
		reason:    "opportunity/handler.go:205→attachment.go:569 ApplyScan，机器事件在事务内更新扫描状态/扫描时间(:599)并写审计；不是历史读取。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.GET(\"/customers\") -> h.listCustomers": {
		operation: "READ_HISTORY",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:112→customers.go:52 ListCustomers，tenant+id键集SELECT Limit100；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.GET(\"/customers/:id/accounts\") -> h.listAccounts": {
		operation: "READ_HISTORY",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:164→queries.go:19 ListAccounts，读取同租户客户IdentityLink脱敏投影Limit100；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.GET(\"/customers/:id/contacts\") -> h.listContacts": {
		operation: "READ_HISTORY",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:152→customers.go:109 ListContacts，activeCustomer及同租户客户SELECT Limit100；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.GET(\"/customers/:id/invitations\") -> h.listInvitations": {
		operation: "READ_HISTORY",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:176→queries.go:35 ListInvitations，只查同租户客户邀请状态Limit100，不签新邀请；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.GET(\"/customers/:id/project-access\") -> h.listProjectAccess": {
		operation: "READ_HISTORY",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:382→queries.go:58 ListProjectAccess，只查MANUAL授权元数据Limit100；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.GET(\"/customers/:id/projects\") -> h.listProjects": {
		operation: "READ_HISTORY",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:274→projects.go:134 ListProjects，仅同租户客户项目SELECT Limit100；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.GET(\"/operations\") -> h.listOperations": {
		operation: "READ_HISTORY",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:220→operations.go:288 ListOperations，按tenant/获准kind有界SELECT，不自动ResumeOperation；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.POST(\"/access/disable\") -> h.disableAccess": {
		operation: "MUTATE_BUSINESS",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:249→disable.go:25 DisableAccess，冻结账号/撤会话/撤邀请并持久化操作，再:80 resumeDisable调用平台RevokeRole/DisableBinding；现行principal(write=true)不豁免，本次保留较严门禁；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.POST(\"/contacts\") -> h.createContact": {
		operation: "MUTATE_BUSINESS",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:129→customers.go:80 CreateContact，加密联系方式并INSERT联系人/审计；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.POST(\"/customers\") -> h.createCustomer": {
		operation: "MUTATE_BUSINESS",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:100→customers.go:20 CreateCustomer，事务ClaimMode/INSERT客户/审计；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.POST(\"/invitations\") -> h.createInvitation": {
		operation: "MUTATE_BUSINESS",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:188→operations.go:35 CreateInvitation→:93 ResumeOperation/:181 resumeInvitation，写操作账本并调用真实平台Provision/AssignRole/BindCustomer、创建邀请；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.POST(\"/invitations/:id/revoke\") -> h.revokeInvitation": {
		operation: "MUTATE_BUSINESS",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:201→operations.go:320 RevokeInvitation，仅把PENDING邀请更新REVOKED并审计；虽属安全收缩，当前principal(write=true)及运行门禁仍按业务写，本次不放宽；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.POST(\"/operations/:number/report-file\") -> h.retryReportFile": {
		operation: "MUTATE_BUSINESS",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:365→reports.go:276 RetryReportUpload→:183 uploadReport，真实上传原PDF及继续摄取排队，非历史下载；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.POST(\"/operations/:number/retry\") -> h.retryOperation": {
		operation: "MUTATE_BUSINESS",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:237→operations.go:93 ResumeOperation，有限分派INVITE/DISABLE/REPORT，可重新授予角色/绑定、排新摄取任务和推进状态；不可整体按安全恢复豁免；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.POST(\"/projects\") -> h.saveProject": {
		operation: "MUTATE_BUSINESS",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:262→projects.go:72 SaveProject，INSERT/乐观UPDATE本地项目并upsert项目快照/审计；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.POST(\"/reports\") -> h.createReport": {
		operation: "MUTATE_BUSINESS",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:345→reports.go:87 CreateReport，创建操作/报告申请/报告来源/状态事件，再上传PDF并queueReport写入摄取任务；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go|group.PUT(\"/project-access\") -> h.projectAccess": {
		operation: "MUTATE_BUSINESS",
		reason:    "仅在固定local目录模式挂载/api/v1/local-admin；handler.go:295→projects.go:150 SetProjectAccess，可授予或撤销MANUAL精确账号项目权限并写隔离模式/审计，不能整体安全豁免；各查询/写仍经localidentity/service.go:84 principal精确管理员、权限、租户、模式复核。",
	},
	"customer_and_opportunity/internal/portalbootstrap/router.go|api.POST(\"/filings/:id/exports\") -> exportFiling(deps)": {
		operation: "EXPORT_HISTORY",
		reason:    "router.go:1579→filing/export.go:40 Export，读取归属和版本正确的既有备案；草稿读当前资料，非草稿解密并验SHA/表单版本的最新提交快照，PDF渲染与导出审计，不创建/提交备案。",
	},
	"customer_and_opportunity/internal/portalbootstrap/router.go|api.POST(\"/filings/:id/materials/:materialID/complete\") -> completeFilingMaterialUpload(deps)": {
		operation: "MUTATE_BUSINESS",
		reason:    "router.go:1791→filing/material.go:321 CompleteUpload，完成上传/网关确认并推进材料状态，不是既有文件下载。",
	},
	"customer_and_opportunity/internal/portalbootstrap/router.go|base.GET(\"/activate\") -> beginActivationLogin(deps)": {
		operation: "MUTATE_BUSINESS",
		reason:    "/activate仍在router.go:169注册，仅由匿名函数提取为:451 beginActivationLogin；调用account/service.go:225 BeginInvitationLogin远端Verify→:260 begin→repository.go:280 CreateActivation写入激活上下文，后续CompleteLogin消费邀请并激活账号；保持写分类且不缩小分母。",
	},
	"customer_and_opportunity/internal/portalbootstrap/router.go|base.GET(\"/auth/session\") -> centralLifecycle(deps, false)": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "当前门户会话及客户/角色认证后代理中央状态；portalbootstrap/session_governance_http.go:12→portal/account/session_governance.go:85 Lifecycle→:53 readCentral→shared/sessiongovernance/client.go:131 State；认证中可能轮换密文令牌/更新last_seen，均限会话。",
	},
	"customer_and_opportunity/internal/portalbootstrap/router.go|base.POST(\"/auth/activity\") -> centralLifecycle(deps, true)": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "同源/CSRF和门户认证后的真实交互仅续中央生命周期；router.go:172→session_governance_http.go:12→account/session_governance.go:85/:53→shared/sessiongovernance/client.go:157 Activity，不新增客户绑定或业务记录。",
	},
	"customer_and_opportunity/internal/temporalworker/metrics.go|mux.Handle(\"GET /metrics\") -> registry": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "presaleworker/app.go:95在Temporal启用时启动StartMetricsServer→metrics.go:153注册；:110 ServeHTTP只输出内存counter/gauge/timer，无业务持久化或外部写。",
	},
	"data_analysis/internal/bootstrap/app.go|base.GET(\"/auth/session\") -> authHandler.SessionState": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "app.go:223先SessionAuth，session_governance_http.go:82/87 lifecycle(false)按Cookie读取现存会话及readCentral状态；中央绑定/主体/租户/周期精确校验，不生成聚合、告警或财务事实。",
	},
	"data_analysis/internal/bootstrap/app.go|base.POST(\"/auth/activity\") -> authHandler.Activity": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "app.go:224先SessionAuth和RequireSameOriginWrite；handler.go:178中央分支lifecycle(true)，本地分支TouchActivity仅更新仍有效会话时间戳（service.go:564）；不延长原绝对期限、不复活已失效会话。",
	},
	"data_analysis/internal/bootstrap/app.go|router.GET(metrics.Path) -> handler.Handle": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "app.go:291的metrics.Path固定/metrics（metrics/access.go:27），RequireAccess为Bearer或TCP回环保护。metrics/handler.go:41只读审计outbox计数及进程指标；不认领/重试/创建事件，错误通过counts_error暴露。",
	},
	"data_analysis/internal/bootstrap/embed.go|api.GET(\"/embed-proxy/:token\") -> <anonymous>": {
		operation: "MUTATE_BUSINESS",
		reason:    "embed.go:32关闭分支DisabledProxy先查MutateBusiness与会话后明确拒绝，不访问旧令牌或Metabase；此精确嵌入资源仍遵循已批准到期禁用策略。",
	},
	"data_analysis/internal/bootstrap/embed.go|api.GET(\"/embed-proxy/:token/*resource\") -> <anonymous>": {
		operation: "MUTATE_BUSINESS",
		reason:    "embed.go:33关闭分支DisabledProxy先查MutateBusiness与会话后明确拒绝；不能因GET/wildcard或关闭模式整体豁免旧令牌资源。",
	},
	"data_analysis/internal/bootstrap/embed.go|api.GET(\"/embed/:dashboard\") -> <anonymous>": {
		operation: "MUTATE_BUSINESS",
		reason:    "embed.go:31关闭分支DisabledIssue先查MutateBusiness/会话/看板权限再明确拒绝；:36启用分支Issue先查MutateBusiness，embedbridge/embed.go:107创建一次性令牌。两个条件表达式同一精确key，均保持到期禁止嵌入。",
	},
	"data_analysis/internal/bootstrap/embed.go|proxy.GET(\"/:token\") -> <anonymous>": {
		operation: "MUTATE_BUSINESS",
		reason:    "embed.go:39启用分支Bridge.Proxy在独立嵌入Host及MutateBusiness检查后embedRecord(consume=true)，embed.go:284条件更新consumed_at并签发Metabase重定向；旧令牌亦重新核验许可。",
	},
	"data_analysis/internal/bootstrap/embed.go|proxy.GET(\"/:token/*resource\") -> <anonymous>": {
		operation: "MUTATE_BUSINESS",
		reason:    "embed.go:40启用分支ProxyResource先查MutateBusiness与仍有效且已消费令牌，再允许受限Metabase资源或活动桥脚本；此为已批准到期禁用的嵌入通道，不能归为历史下载。",
	},
	"data_analysis/internal/standardfacts/http.go|group.GET(\"/status\") -> h.Status": {
		operation: "READ_HISTORY",
		reason:    "http.go:28/68→store.go:95仅读当前租户来源发布指针/批次摘要，联动模式不查询标准表，不变更业务。",
	},
	"data_analysis/internal/standardfacts/http.go|group.GET(\"/template\") -> h.Template": {
		operation: "READ_HISTORY",
		reason:    "http.go:27/59在aggregation.manage与TENANT范围下返回空事实JSON模板、限额和字段说明；不创建文件/批次/事实，故非业务变更或历史文件导出。",
	},
	"data_analysis/internal/standardfacts/http.go|group.POST(\"/preview\") -> h.Preview": {
		operation: "READ_HISTORY",
		reason:    "http.go:29/96有界严格Decode与Analyze，store.go:128只调用Analyze+Status绑定当前版本；无预览批次持久化、事实写入或新聚合发布。不能按POST猜业务写。",
	},
	"data_analysis/internal/standardfacts/http.go|group.POST(\"/publish\") -> h.Publish": {
		operation: "MUTATE_BUSINESS",
		reason:    "http.go:30/126→store.go:144事务内重核MutateBusiness，锁来源版本后写发布批次/事实/快照/指针及审计；即使名为标准来源也是新财务事实发布。",
	},
	"project_management/internal/httpapi/api.go|api.GET(\"/project-creation-capabilities\") -> h.projectCreationCapabilities": {
		operation: "READ_HISTORY",
		reason:    "projectCreationCapabilities(httpapi/native_project.go:13)仅输出service.NativeEnabled模式标记，无服务命令、仓储或外部写入。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|api.GET(\"/projects/:id/expense-participants\") -> h.expenseParticipants": {
		operation: "READ_HISTORY",
		reason:    "expenseParticipants(httpapi/reimbursement.go:231)→ListExpenseParticipants(mysql/reimbursement_participants.go:60)；GetProject核对当前项目范围后分页Find显式资格，不更新名单。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|api.GET(\"/projects/:id/native-scope\") -> h.nativeScope": {
		operation: "READ_HISTORY",
		reason:    "nativeScope(httpapi/native_scope.go:15)→ReadNativeScope(application/native_scope_change.go:80)→Repository.ReadNativeScope(mysql/native_scope_change.go:78)；读取当前范围、版本和最近20提案，事务锁用于一致性，无业务写。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|api.GET(\"/projects/:id/reimbursement-spend\") -> h.expenseSpendSummary": {
		operation: "READ_HISTORY",
		reason:    "expenseSpendSummary(httpapi/reimbursement.go:272)→GetProject(scope)→ReadExpenseSpendSummary(mysql/reimbursement_reconcile.go:176)；读取既有贡献SUM和来源版本快照，不调用Apply/Rebuild或回写统计。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|api.POST(\"/native-projects\") -> h.createNativeProject": {
		operation: "MUTATE_BUSINESS",
		reason:    "createNativeProject(httpapi/native_project.go:20)→Service.CreateNativeProject(application/native_project.go:27)→Repository.CreateNativeProject(mysql/native_project.go:60)；权限/组织/冻结/目录检查后事务创建本地委托方、项目、服务范围版本和审计。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|api.POST(\"/projects/:id/native-scope-changes\") -> h.requestNativeScope": {
		operation: "MUTATE_BUSINESS",
		reason:    "requestNativeScope(httpapi/native_scope.go:27)→RequestNativeScope(application/native_scope_change.go:95)→Repository.RequestNativeScope(mysql/native_scope_change.go:114)；保存PENDING提案和交付事件/审计，未生效也是新业务状态。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|api.POST(\"/projects/:id/native-scope-changes/:changeID/decision\") -> h.decideNativeScope": {
		operation: "MUTATE_BUSINESS",
		reason:    "decideNativeScope(httpapi/native_scope.go:43)→DecideNativeScope(application/native_scope_change.go:134)→Repository.DecideNativeScope(mysql/native_scope_change.go:224)；保存批准/驳回状态及审计；批准另经applyNativeScope:304替换服务范围、创建新版本并更新项目。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|api.POST(\"/projects/:id/native-scope-changes/:changeID/withdraw\") -> h.withdrawNativeScope": {
		operation: "MUTATE_BUSINESS",
		reason:    "withdrawNativeScope(httpapi/native_scope.go:58)固定WITHDRAW→DecideNativeScope(application/native_scope_change.go:134/mysql/native_scope_change.go:224)；核对申请人后写WITHDRAWN、版本与审计，不是会话技术服务。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|api.PUT(\"/projects/:id/expense-participants\") -> h.expenseManageParticipant": {
		operation: "MUTATE_BUSINESS",
		reason:    "expenseManageParticipant(httpapi/reimbursement.go:256)→ManageExpenseParticipant(application/reimbursement.go:215)→SaveExpenseParticipant(mysql/reimbursement_participants.go:76)；锁项目、核对版本、写资格/撤销状态及事务审计。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|read.POST(\"/project-references/batch-validate\") -> h.expenseBatchValidate": {
		operation: "READ_HISTORY",
		reason:    "expenseBatchValidate(httpapi/reimbursement.go:104)→BatchValidateExpenseReferences(application/reimbursement.go:186)→resolveExpenseReference:95；重新读取目录、项目、名单、服务项和差旅事件，返回草稿绑定资格，不冻结或保存单据。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|read.POST(\"/project-references/search\") -> h.expenseSearch": {
		operation: "READ_HISTORY",
		reason:    "expenseSearch(httpapi/reimbursement.go:84)→SearchExpenseReferences(application/reimbursement.go:139)→ExpenseCandidates(mysql/reimbursement_participants.go:20)、GetProject/GetServiceItem、ReadExpenseParticipant；人员/审批目录List及差旅来源读取只返回资格，无业务写入。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|router.GET(\"/auth/session\") -> <anonymous>": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "api.go:237→OIDCAuthenticator.SessionState(platform/session_governance.go:162)→lifecycle:176→Authenticate/FindSession/centralState:30→CentralClient.State(sessiongovernance/client.go:131)；只维护认证/中央会话，不通过轮询续期或修改项目。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|router.POST(\"/auth/activity\") -> <anonymous>": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "api.go:238→OIDCAuthenticator.Activity(platform/session_governance.go:169)→lifecycle:176→centralState(interact=true)→CentralClient.Activity(sessiongovernance/client.go:157)；精确Origin/CSRF后记录真实交互，只更新会话技术状态。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|write.GET(\"/spend-snapshots/reconciliation\") -> h.expenseReconcile": {
		operation: "READ_HISTORY",
		reason:    "expenseReconcile(httpapi/reimbursement.go:174)→ReconcileExpenseSpend(mysql/reimbursement_reconcile.go:21)→readExpenseSnapshot:65/checkExpenseMaterialization:83；事务只读取实际物化和历史快照，内存比较差异，不调用重建或写修复审计。write组是凭据用途。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|write.POST(\"/spend-snapshots\") -> h.expenseSpendSnapshot": {
		operation: "MUTATE_BUSINESS",
		reason:    "expenseSpendSnapshot(httpapi/reimbursement.go:132)→ApplyExpenseSpend(mysql/reimbursement_spend.go:21)→consumeExpenseSpend:53；写inbox、不可变快照、付款效果/分摊、贡献和游标，冲突还写pm_expense_conflict，属于机器业务摄取。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/httpapi/api.go|write.POST(\"/spend-snapshots/rebuild\") -> h.expenseRebuild": {
		operation: "MUTATE_BUSINESS",
		reason:    "expenseRebuild(httpapi/reimbursement.go:204)→RebuildExpenseSpend(mysql/reimbursement_reconcile.go:148)；锁游标核验快照后replaceExpenseProjection删除/重建贡献并写pm_expense_repair，是业务投影修复。 证据路径相对所属仓internal/，mysql指infrastructure/mysql；保留现有身份、租户、数据范围与审计约束。",
	},
	"project_management/internal/temporalworker/metrics.go|mux.Handle(\"GET /metrics\") -> registry": {
		operation: "ESSENTIAL_SERVICE",
		reason:    "StartMetricsServer:158/163 registers metrics; MetricsRegistry.ServeHTTP:109 only reads in-memory counters/gauges/timers; no business persistence or external mutation.",
	},
	"reimbursement/internal/jobs/processor.go|Outbox:project_rebuild": {
		operation: "MUTATE_BUSINESS",
		reason:    "请求项目端按当时最新快照重建支出投影并记录repair；载荷不绑定批准时revision/hash，必须业务写许可。本轮修复统一ESSENTIAL门禁，先许可后远端调用。",
	},
	"reimbursement/internal/transport/httpapi/recovery.go|group.GET(\"/approval-recovery-tasks\") -> h.approvalRecoveryQuery": {
		operation: "READ_HISTORY",
		reason:    "QueryApprovalRecovery读取范围内当前待办/冻结快照，不接管或证明已经失职。",
	},
	"reimbursement/internal/transport/httpapi/recovery.go|group.GET(\"/approval-recovery-tasks/:id\") -> h.approvalRecoveryQuery": {
		operation: "READ_HISTORY",
		reason:    "读取同租户/组织内非申请人当前待办，恢复提交时独立实时核验资格。",
	},
	"reimbursement/internal/transport/httpapi/recovery.go|group.GET(\"/claims/:id/spend-capacity\") -> h.spendCapacity": {
		operation: "READ_HISTORY",
		reason:    "SpendCapacity只读取单据/付款效果并纯函数估算容量，不预留或记账。",
	},
	"reimbursement/internal/transport/httpapi/recovery.go|group.GET(\"/claims/:id/sync-status\") -> h.syncStatus": {
		operation: "READ_HISTORY",
		reason:    "只读既有财务快照、受理回执和投递状态，不重新发命令。",
	},
	"reimbursement/internal/transport/httpapi/recovery.go|group.GET(\"/operations/:kind\") -> h.operationsQuery": {
		operation: "READ_HISTORY",
		reason:    "分页读取范围内失败队列及对账；OperationsQuery只COUNT/SELECT，不领取或重投。",
	},
	"reimbursement/internal/transport/httpapi/recovery.go|group.POST(\"/claims/:id/spend-replays\") -> h.spendReplay": {
		operation: "MUTATE_BUSINESS",
		reason:    "人工恢复修改队列版本、租约与恢复/审计账本；不新增付款仍保留当前POST业务写门禁。",
	},
	"reimbursement/internal/transport/httpapi/recovery.go|group.POST(\"/operations/:kind/:id/retry\") -> h.operationsRetry": {
		operation: "MUTATE_BUSINESS",
		reason:    "RecoverIntegration可重新排解析/重建工作并写恢复幂等账本、审计；按最大副作用保留业务写门禁。",
	},
	"reimbursement/internal/transport/httpapi/router.go|group.POST(\"/approval-tasks/:id/recover\") -> h.command(\"approval.recover\")": {
		operation: "MUTATE_BUSINESS",
		reason:    "approval_recovery.go按双版本、职责分离、原人失职及目标岗位证明修改审批任务受理人及审计。",
	},
}

// applyReviewedEntries20261110 用途：应用本轮逐项核对的精确许可边界。
// 参数：r 为从真实源码生成的候选清单。
// 返回值：无；未知来源、方法、路径及处理器不继承本轮策略。
func applyReviewedEntries20261110(r *Registry) {
	for i := range r.Entries {
		entry := &r.Entries[i]
		if policy, ok := reviewedEntries20261110[entry.Source+"|"+entry.Registration]; ok {
			entry.Operation, entry.Reason = policy.operation, policy.reason
		}
	}
}

var reviewedRouteFiles20261110 = map[string][]string{
	"contract_management": {
		"contract_management/internal/temporalworker/metrics.go",
		"contract_management/internal/transport/httpapi/contract_import.go",
		"contract_management/internal/transport/httpapi/import_confirmation.go",
		"contract_management/internal/transport/httpapi/local_catalog.go",
		"contract_management/internal/transport/httpapi/local_catalog_grants.go",
	},
	"customer_and_opportunity": {
		"customer_and_opportunity/internal/temporalworker/metrics.go",
	},
	"customer_portal": {
		"customer_and_opportunity/internal/modules/portal/localidentity/handler.go",
	},
	"data_analysis": {
		"data_analysis/internal/bootstrap/embed.go",
		"data_analysis/internal/standardfacts/http.go",
	},
	"project_management": {
		"project_management/internal/temporalworker/metrics.go",
	},
	"settlement": {
		"Settlement/internal/httpapi/contract_import.go",
		"Settlement/internal/httpapi/contract_import_export.go",
		"Settlement/internal/httpapi/contract_import_query.go",
		"Settlement/internal/httpapi/native_financial.go",
	},
}

var reviewedExecutionSources20261110 = []string{
	"Settlement/internal/httpapi/contract_import.go",
	"Settlement/internal/httpapi/contract_import_export.go",
	"Settlement/internal/httpapi/contract_import_query.go",
	"Settlement/internal/httpapi/native_financial.go",
	"Settlement/internal/platform/session_governance.go",
	"Settlement/internal/platform/session_governance_http.go",
	"Settlement/internal/service/contract_import.go",
	"Settlement/internal/service/import_export.go",
	"Settlement/internal/service/import_reconciliation.go",
	"Settlement/internal/service/native_financial.go",
	"Settlement/internal/service/native_financial_query.go",
	"Settlement/internal/sessiongovernance/client.go",
	"Settlement/internal/sessiongovernance/rotation.go",
	"contract_management/internal/application/contract_import_auto_approve.go",
	"contract_management/internal/application/contract_import_cleanup.go",
	"contract_management/internal/application/contract_import_commit.go",
	"contract_management/internal/application/contract_import_confirmation.go",
	"contract_management/internal/application/contract_import_dispatcher.go",
	"contract_management/internal/application/contract_import_preview.go",
	"contract_management/internal/application/contract_import_provision.go",
	"contract_management/internal/application/contract_import_store.go",
	"contract_management/internal/application/contract_import_targets.go",
	"contract_management/internal/application/contract_import_upload_recovery.go",
	"contract_management/internal/application/contract_source.go",
	"contract_management/internal/application/contract_sources.go",
	"contract_management/internal/application/crm_provision.go",
	"contract_management/internal/application/import_document.go",
	"contract_management/internal/application/local_catalog.go",
	"contract_management/internal/application/local_catalog_grants.go",
	"contract_management/internal/infrastructure/mysql/contract_import.go",
	"contract_management/internal/infrastructure/mysql/contract_import_auto_approve.go",
	"contract_management/internal/infrastructure/mysql/contract_import_auto_approve_setting.go",
	"contract_management/internal/infrastructure/mysql/import_crm_provision.go",
	"contract_management/internal/infrastructure/mysql/import_report.go",
	"contract_management/internal/infrastructure/mysql/import_targets.go",
	"contract_management/internal/infrastructure/mysql/local_catalog.go",
	"contract_management/internal/infrastructure/mysql/local_catalog_grants.go",
	"contract_management/internal/infrastructure/mysql/local_outbound.go",
	"contract_management/internal/infrastructure/mysql/status_change_acceptance.go",
	"contract_management/internal/infrastructure/platform/local_catalog_directory.go",
	"contract_management/internal/infrastructure/platform/session_governance.go",
	"contract_management/internal/infrastructure/platform/session_governance_store.go",
	"contract_management/internal/infrastructure/platform/session_rotation.go",
	"contract_management/internal/integration/crm/import_provision.go",
	"contract_management/internal/lifecycle/waiter.go",
	"contract_management/internal/sessiongovernance/client.go",
	"contract_management/internal/sessiongovernance/rotation.go",
	"contract_management/internal/temporalworker/metrics.go",
	"contract_management/internal/transport/httpapi/contract_import.go",
	"contract_management/internal/transport/httpapi/import_confirmation.go",
	"contract_management/internal/transport/httpapi/local_catalog.go",
	"contract_management/internal/transport/httpapi/local_catalog_grants.go",
	"contract_management/internal/transport/httpapi/templates_source.go",
	"customer_and_opportunity/internal/bootstrap/audit_outbox_readiness.go",
	"customer_and_opportunity/internal/bootstrap/health.go",
	"customer_and_opportunity/internal/bootstrap/presale_opportunity_handler.go",
	"customer_and_opportunity/internal/bootstrap/runtime_capabilities.go",
	"customer_and_opportunity/internal/modules/crmauth/governance.go",
	"customer_and_opportunity/internal/modules/crmauth/governance_http.go",
	"customer_and_opportunity/internal/modules/crmauth/governance_refresh.go",
	"customer_and_opportunity/internal/modules/crmauth/governance_repository.go",
	"customer_and_opportunity/internal/modules/customer/provision.go",
	"customer_and_opportunity/internal/modules/customer/repository.go",
	"customer_and_opportunity/internal/modules/importprovision/handler.go",
	"customer_and_opportunity/internal/modules/importprovision/service.go",
	"customer_and_opportunity/internal/modules/opportunity/attachment.go",
	"customer_and_opportunity/internal/modules/opportunity/handler.go",
	"customer_and_opportunity/internal/modules/opportunity/provision.go",
	"customer_and_opportunity/internal/modules/opportunity/repository.go",
	"customer_and_opportunity/internal/modules/opportunity/service.go",
	"customer_and_opportunity/internal/modules/opportunity/win_confirmation.go",
	"customer_and_opportunity/internal/modules/opportunity/win_confirmation_handler.go",
	"customer_and_opportunity/internal/modules/portal/account/repository.go",
	"customer_and_opportunity/internal/modules/portal/account/session_governance.go",
	"customer_and_opportunity/internal/modules/portal/account/session_governance_repository.go",
	"customer_and_opportunity/internal/modules/portal/filing/material.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/authority.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/config.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/customer_ref.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/customer_reports.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/customers.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/disable.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/files_port.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/handler.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/identity_reservation.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/invite_adapter.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/metadata.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/model.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/operations.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/projects.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/queries.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/reports.go",
	"customer_and_opportunity/internal/modules/portal/localidentity/service.go",
	"customer_and_opportunity/internal/modules/portal/report/delivery_recovery.go",
	"customer_and_opportunity/internal/portalbootstrap/local_identity_routes.go",
	"customer_and_opportunity/internal/portalbootstrap/session_governance_http.go",
	"customer_and_opportunity/internal/presaleworker/app.go",
	"customer_and_opportunity/internal/shared/sessiongovernance/client.go",
	"customer_and_opportunity/internal/shared/sessiongovernance/rotation.go",
	"customer_and_opportunity/internal/shutdown/groups.go",
	"customer_and_opportunity/internal/temporalworker/metrics.go",
	"customer_and_opportunity/internal/workerlicense/gate.go",
	"data_analysis/cmd/aggregation-worker/integration_sources.go",
	"data_analysis/internal/bootstrap/embed.go",
	"data_analysis/internal/bootstrap/integration_sync.go",
	"data_analysis/internal/embedbridge/disabled.go",
	"data_analysis/internal/embedbridge/origin.go",
	"data_analysis/internal/embedbridge/session_activity.go",
	"data_analysis/internal/embedbridge/session_activity.js",
	"data_analysis/internal/metrics/access.go",
	"data_analysis/internal/metrics/handler.go",
	"data_analysis/internal/oidc/session_governance.go",
	"data_analysis/internal/oidc/session_governance_http.go",
	"data_analysis/internal/oidc/session_rotation.go",
	"data_analysis/internal/sessiongovernance/client.go",
	"data_analysis/internal/sessiongovernance/rotation.go",
	"data_analysis/internal/shutdown/group.go",
	"data_analysis/internal/standardfacts/decode.go",
	"data_analysis/internal/standardfacts/finance.go",
	"data_analysis/internal/standardfacts/http.go",
	"data_analysis/internal/standardfacts/store.go",
	"project_management/authz/permission-manifest.json",
	"project_management/internal/application/native_project.go",
	"project_management/internal/application/native_scope_change.go",
	"project_management/internal/application/reimbursement.go",
	"project_management/internal/application/reimbursement_approvers.go",
	"project_management/internal/application/reimbursement_directory_cache.go",
	"project_management/internal/application/reimbursement_reference_facts.go",
	"project_management/internal/config/config.go",
	"project_management/internal/config/reimbursement.go",
	"project_management/internal/domain/native_scope.go",
	"project_management/internal/domain/reimbursement.go",
	"project_management/internal/domain/reimbursement_approvers.go",
	"project_management/internal/domain/reimbursement_spend.go",
	"project_management/internal/domain/reimbursement_travel.go",
	"project_management/internal/httpapi/native_project.go",
	"project_management/internal/httpapi/native_scope.go",
	"project_management/internal/httpapi/reimbursement.go",
	"project_management/internal/infrastructure/mysql/native_project.go",
	"project_management/internal/infrastructure/mysql/native_scope_change.go",
	"project_management/internal/infrastructure/mysql/native_scope_guards.go",
	"project_management/internal/infrastructure/mysql/reimbursement_participants.go",
	"project_management/internal/infrastructure/mysql/reimbursement_reconcile.go",
	"project_management/internal/infrastructure/mysql/reimbursement_records.go",
	"project_management/internal/infrastructure/mysql/reimbursement_reference_facts.go",
	"project_management/internal/infrastructure/mysql/reimbursement_spend.go",
	"project_management/internal/platform/owner_directory.go",
	"project_management/internal/platform/reimbursement_token.go",
	"project_management/internal/platform/session_governance.go",
	"project_management/internal/platform/session_governance_refresh.go",
	"project_management/internal/platform/session_governance_store.go",
	"project_management/internal/sessiongovernance/client.go",
	"project_management/internal/sessiongovernance/rotation.go",
	"project_management/internal/temporalworker/metrics.go",
}

var reviewedImportTasks20261110 = []Entry{
	{Application: "contract_management", Source: "contract_management/internal/application/contract_import_dispatcher.go",
		Registration: "*Service.RunImportDispatcher", Operation: "MUTATE_BUSINESS",
		Reason: "循环显式调用CleanupImportSources/DispatchImportTargetOnce/DispatchImportOnce，包含新合同/CRM开立/确认/放行；不可整体ESSENTIAL。"},
	{Application: "contract_management", Source: "contract_management/internal/application/contract_import_dispatcher.go",
		Registration: "*Service.DispatchImportOnce", Operation: "MUTATE_BUSINESS",
		Reason: "认领并执行新导入、CRM开立和自动确认；当前先Claim再checkBusinessLicense，拒绝后会写PARTIAL/BLOCKED，分类不证明过期待执行状态保持。"},
	{Application: "contract_management", Source: "contract_management/internal/application/contract_import_targets.go",
		Registration: "*Service.DispatchImportTargetOnce", Operation: "MUTATE_BUSINESS",
		Reason: "许可核验后认领、发送项目/财务新业务与报表放行状态查询并保存回执；不是纯已有不可变审计投递。"},
	{Application: "contract_management", Source: "contract_management/internal/application/contract_import_cleanup.go",
		Registration: "*Service.CleanupImportSources", Operation: "MUTATE_BUSINESS",
		Reason: "变更EXPIRED/CLEANED/epoch并退役暂存请求、释放引用。当前无自身商业检查，不把整个循环豁免必要服务；本报告未修改运行行为。"},
}
