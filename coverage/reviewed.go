package coverage

import "strings"

// 这些处理器允许列表限定于有限的源码范围，已对照商业授权入口矩阵
// 及处理器和服务实现审核；新增名称不得继承已有策略。
// 对持久化台账求和或筛选属于历史查询，
// 聚合采集、生成新告警及推进业务状态则不属于历史查询。
var semanticGroups = map[string]map[string]string{
	"contract_management/internal/transport/httpapi/router.go": {
		"READ_HISTORY":      "h.dashboard h.listSettlementCompletedContracts h.listProjectApprovedContracts h.listProjectApprovedContractReferences h.getProjectApprovedContract h.getProjectApprovedContractServiceItems h.listOpportunityIntakes h.getOpportunityIntake h.listDetectionCategories h.listContracts h.listApprovedContracts h.listSigningRecords h.getSigningRecord h.getContract h.listContractLifecycle h.previewContract h.listTemplates h.listApprovals h.listTasks h.getApproval h.previewApprovalContract h.listRules",
		"MUTATE_BUSINESS":   "h.acceptOpportunityIntake h.reviewOpportunityIntake h.updateContractDraft h.uploadStampedPDF h.saveSigningShipment h.markSigningReceived h.recordSigningReminder h.confirmSigning h.createTemplate h.updateTemplate h.replaceTemplateSource h.deleteTemplate h.changeStatus h.createRule h.updateRule h.deleteRule h.command(action)",
		"ESSENTIAL_SERVICE": "h.webHome",
	},
	"customer_and_opportunity/internal/bootstrap/app.go": {
		"ESSENTIAL_SERVICE": "auditOutboxStatusHandler(auditStore,config) runtimeCapabilitiesHandler(config,workerReadiness,config.PresaleWorkerHeartbeatMaxAge)",
		"READ_HISTORY":      "opportunityPresales.List",
	},
	"customer_and_opportunity/internal/modules/contractreference/routes.go": {"READ_HISTORY": "handler.Resolve"},
	"customer_and_opportunity/internal/modules/credit/routes.go": {
		"READ_HISTORY":    "h.Level h.History h.PaymentRecords h.Pending h.RuleSettings h.Statistics",
		"MUTATE_BUSINESS": "h.Apply h.Withdraw h.Approve h.Reject h.UpdateRuleSettings h.Payment",
	},
	"customer_and_opportunity/internal/modules/customer/routes.go": {
		"READ_HISTORY":      "handler.List handler.CheckDuplicate handler.Get handler.ListContacts handler.ListStakeholders handler.ListInformationSystems handler.OpportunityHistory handler.ProjectHistory handler.ListChangeLogs handler.ListFollowups",
		"MUTATE_BUSINESS":   "handler.Merge handler.PreviewImport handler.CommitImport handler.Create handler.ReplaceStakeholders handler.ReplaceInformationSystems handler.Update handler.Void handler.Restore handler.CreateFollowup",
		"EXPORT_HISTORY":    "handler.ImportErrors handler.CreateExport",
		"ESSENTIAL_SERVICE": "handler.DownloadImportTemplate",
	},
	"customer_and_opportunity/internal/modules/notification/routes.go": {"READ_HISTORY": "handler.ListMine handler.UnreadCount", "ESSENTIAL_SERVICE": "handler.MarkRead"},
	"customer_and_opportunity/internal/modules/opportunity/routes.go": {
		"READ_HISTORY":    "handler.ListCatalogItems handler.StageAlertRules handler.StageAlerts handler.List handler.Board handler.Get handler.GetMembers handler.ListMemberTerms handler.StageHistory handler.ListFollowups handler.ExternalStatus handler.AttachmentCapabilities handler.ListAttachments",
		"MUTATE_BUSINESS": "handler.CreateCatalogItem handler.UpdateCatalogItem handler.DeleteCatalogItem handler.UpdateStageAlertRule handler.Create handler.Update handler.ChangeOwner handler.ReplaceMembers handler.Void handler.Restore handler.ChangeStage handler.CreateFollowup handler.LaunchQuotation handler.LaunchBid handler.CompleteTerminalTodo handler.ContractTransfer handler.CreateAttachmentUpload handler.UploadAttachmentContent handler.CompleteAttachmentUpload handler.ApplyExternalStatus handler.ContractLinkCallback handler.ApplyAttachmentScan",
		"EXPORT_HISTORY":  "handler.DownloadAttachment", "ESSENTIAL_SERVICE": "handler.ReadStageAlert",
	},
	"customer_and_opportunity/internal/modules/ownerdirectory/routes.go":        {"READ_HISTORY": "handler.List"},
	"customer_and_opportunity/internal/modules/portalinvite/routes.go":          {"READ_HISTORY": "handler.Current", "MUTATE_BUSINESS": "handler.Create handler.Consume", "ESSENTIAL_SERVICE": "handler.Verify handler.Revoke handler.Disable"},
	"customer_and_opportunity/internal/modules/portalreportreference/routes.go": {"READ_HISTORY": "handler.GetDownloadIdentity"},
	"customer_and_opportunity/internal/modules/presale/routes.go": {
		"READ_HISTORY":    "handler.ExecutionDepartments handler.ListRequests handler.Board handler.FilterOptions handler.RequestDetail handler.ContactPhone handler.Timeline handler.AvailableActions handler.ApprovalHistory handler.Assignments handler.Worklogs handler.AlertRules handler.ApprovalRules handler.Alerts handler.ReportSummary handler.ReportTrend handler.ReportDistribution handler.ReportFilterOptions",
		"MUTATE_BUSINESS": "handler.CreateRequest handler.ReopenRequest handler.ApprovalAction handler.ReplaceAssignments handler.SelectExecutionDepartment handler.AddProgress handler.Complete handler.Cancel handler.AddWorklog handler.UpdateAlertRule handler.CreateApprovalRule handler.UpdateApprovalRule handler.DeleteApprovalRule handler.ApprovalCallback",
		"EXPORT_HISTORY":  "handler.RequestReportExport", "ESSENTIAL_SERVICE": "handler.ReadAlert",
	},
	"customer_and_opportunity/internal/portalbootstrap/router.go": {
		"READ_HISTORY":      "getCustomerCapabilities(deps) listReportRiskAlertsForReview(deps) listFeedbackForOperator(deps) evaluationStatistics(deps) listEvaluationNotices(deps) listProjectHistoryForCRM(deps) listProjectConversationsForManager(deps) getProjectConversationForManager(deps) listProjects(deps) getProject(deps) listActivities(deps) getCurrentProjectConversation(deps) getProjectConversation(deps) evaluationEligibility(deps) getProjectExport(deps) listReports(deps) getReport(deps) listReportNotifications(deps) reportNotificationUnreadCount(deps) listOwnedReportRiskAlerts(deps) listFeedback(deps) getFeedback(deps) listFeedbackNotifications(deps) countFeedbackNotifications(deps) getEvaluation(deps) listFilings(deps) getFiling(deps) validateFiling(deps)",
		"MUTATE_BUSINESS":   "updateCustomerCapabilities(deps) reportCallback(deps) reviewReportRiskAlert(deps) unlockFiling(deps) filingMaterialScanCallback(deps) sendProjectMessageForManager(deps) processFeedback(deps,action) createProjectConversation(deps) sendProjectMessage(deps) createFeedback(deps) addFeedbackMessage(deps) closeFeedback(deps) submitEvaluation(deps) createFiling(deps) saveFilingSection(deps) saveFilingMatrix(deps) submitFiling(deps) deleteFiling(deps) createFilingMaterialUpload(deps) uploadFilingMaterialContent(deps) completeFilingMaterialUpload(deps)",
		"ESSENTIAL_SERVICE": "disableAccount(deps) readEvaluationNotice(deps) readProjectMessagesForManager(deps) capabilities(deps) accountSecurity(deps) accountSessions(deps) revokeAccountSession(deps) acknowledgeSecurityEvent(deps) readProjectMessages(deps) readReportNotification(deps) readFeedbackNotification(deps)",
		"EXPORT_HISTORY":    "unsupported(\"PORTAL_FILING_EXPORT_NOT_CONFIGURED\")",
	},
	"data_analysis/internal/bootstrap/app.go": {
		"READ_HISTORY":    "alertHandler.List alertHandler.Summary dictHandler.Get dictHandler.History dashboardHandler.Contract dashboardHandler.Project dashboardHandler.Business(\"report\") dashboardHandler.Business(\"finance\") dashboardHandler.Contracts dashboardHandler.Trend adminHandler.ListSources adminHandler.ListAlertRules",
		"MUTATE_BUSINESS": "dictHandler.Put dictHandler.Create dictHandler.Update dictHandler.SetEnabled dictHandler.Delete adminHandler.TriggerAllSources adminHandler.PutAlertRules",
	},
	"project_management/internal/httpapi/api.go": {
		"READ_HISTORY":      "h.listContractDetectionCategories h.dashboard h.navigation h.roleCatalog h.ruleConfigurationCatalog h.listProjects h.monitorProjects h.listApprovedContracts h.listApprovedContractServiceItems h.getProject h.listDeliveryEvents h.listSlaOverdue h.listServiceItems h.getPenetrationWorkPackage h.listPenetrationReportRevisions h.listReportRevisions h.listPersonnel h.listQualifiedPersonnel h.resolvePersonnelNames h.listCapabilities h.listEquipment h.listEquipmentReservations h.listSites h.listRules h.getSplitPolicy h.listDetectionCategories h.listSplitOverrides",
		"MUTATE_BUSINESS":   "h.createProject h.adjustDecomposition h.ensurePenetrationWorkPackage h.savePenetrationDecision h.savePenetrationPlan h.advancePenetrationExecution h.registerPenetrationReportArtifact h.advancePenetrationReport h.uploadEvidence h.registerReportArtifact h.confirmServiceItems h.assignTeam h.revokeTeamAssignment h.returnToDecomposition h.assignExecutionTeam h.revokeExecutionAssignment h.planImplementation h.revokeImplementationPlan h.startPreparation h.revokePreparation h.startFieldExecution h.checkInField h.captureFieldSignature h.requestRollback h.withdrawRollback h.decideRollback h.requestReportCorrection h.decideReportCorrection h.reportDeviation h.reviewDeviation h.upsertCapability h.importCapabilities h.previewCapabilitiesImport h.syncPersonnelIdentities h.returnEquipment h.upsertEquipment h.importEquipment h.previewEquipmentImport h.deleteEquipment h.upsertSite h.deleteSite h.createRule h.updateRule h.updateConfigRule h.deleteRule h.saveSplitPolicy h.saveDetectionCategory h.importDetectionCategories h.previewDetectionCategoryImport h.deleteDetectionCategory h.saveSplitOverride h.deleteSplitOverride h.reviewSpecialMethod h.updateReportStatus",
		"ESSENTIAL_SERVICE": "h.capabilityImportTemplate h.equipmentImportTemplate h.detectionCategoryImportTemplate",
	},
	"Settlement/internal/httpapi/api.go": {
		"READ_HISTORY":      "a.dashboard a.invoicedReceivablesTop10 a.reportExport a.listPlans a.listReceivables a.listReceipts a.receiptMatches a.listAllocations a.listInvoiceRequests a.listTaxInvoices a.taxInvoiceDetail a.listDunningPolicies a.listDunningCases a.listDunningActions a.listDunningRecipients a.listNotifications a.unreadCount",
		"EXPORT_HISTORY":    "a.createInvoicedReceivablesExport a.createAgingReceivablesExport a.downloadReportExport a.downloadInvoiceDocument",
		"MUTATE_BUSINESS":   "a.contractEvent a.taxResult a.retryCreditSync a.confirmPlan a.createReceipt a.createAllocation a.reverseAllocation a.uploadInvoiceDocument a.requestInvoiceRedFlush a.reviewInvoiceRedFlush a.createInvoiceRequest a.approveInvoiceRequest a.recoverInvoiceFailure a.rejectInvoiceRequest a.manualIssueInvoice a.createDunningPolicy",
		"ESSENTIAL_SERVICE": "a.login a.me a.readNotification",
	},
}

// semanticClassify 应用经审核的精确来源和处理器策略，不按HTTP方法放行。
// 参数：e 为待分类条目；handler 为实际处理器表达式。
// 返回值：无；未匹配的条目保留原分类。
func semanticClassify(e *Entry, handler string) {
	if review20261028Classify(e, handler) {
		return
	}
	if upgradeSemanticClassify(e, handler) {
		return
	}
	for op, list := range semanticGroups[e.Source] {
		normalized := strings.ReplaceAll(handler, " ", "")
		for _, name := range strings.Fields(list) {
			if normalized != name {
				continue
			}
			e.Operation = op
			e.Reason = "Source-scoped semantic review: " + handler + "; existing auth, tenant/data-scope and audit requirements remain enforced"
			if strings.HasPrefix(name, "dashboardHandler.Business(") {
				e.Reason += "; reads an existing tenant-scoped business_dashboard_snapshot only; does not trigger aggregation or write a new projection"
			}
			if name == "a.recoverInvoiceFailure" {
				e.Reason += "; transactional failed-invoice review returns ISSUE_FAILED to SUBMITTED and increments version; does not issue a tax command but advances business state"
			}
			if handler == "handler.RequestReportExport" || strings.HasPrefix(handler, "unsupported(") {
				e.Reason += "; currently returns explicit unavailable error, classification is not proof export works"
			}
			return
		}
	}
}
