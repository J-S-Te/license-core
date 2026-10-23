package coverage

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

// 新增的有限条目已对照交接、审计及源码台账审核。
// 辅助注册文件也须扫描；仅计算调用方摘要会遗漏路由。
var upgradeSemanticGroups = map[string]map[string]string{
	"contract_management/internal/transport/httpapi/router.go": {
		"ESSENTIAL_SERVICE": "h.listFailedAuditEvents h.retryFailedAuditEvent",
		"MUTATE_BUSINESS":   "h.recoverApproval",
	},
	"contract_management/internal/transport/httpapi/handover.go": {
		"READ_HISTORY":    "h.handoverResponsibilities h.handoverReceipt",
		"MUTATE_BUSINESS": "h.handoverFreeze h.handoverTransfer h.handoverComplete",
	},
	"customer_and_opportunity/internal/modules/handover/handler.go": {
		"READ_HISTORY":    "handler.responsibilities handler.receipt",
		"MUTATE_BUSINESS": "handler.freeze handler.transfer handler.complete",
	},
	"customer_and_opportunity/internal/modules/opportunity/routes.go": {
		"READ_HISTORY":    "handler.ListFailedContractTransfers",
		"MUTATE_BUSINESS": "handler.RetryFailedContractTransfer",
	},
	"customer_and_opportunity/internal/modules/presale/routes.go": {
		"MUTATE_BUSINESS": "handler.ReplaceExecutorAssignments handler.TakeoverExecutingTeamLead handler.RejectExecutionAssignment",
	},
	"project_management/internal/httpapi/api.go": {
		"READ_HISTORY":      "h.handoverResponsibilities h.handoverReceipt",
		"MUTATE_BUSINESS":   "h.handoverFreeze h.handoverTransfer h.handoverComplete",
		"ESSENTIAL_SERVICE": "h.listFailedAuditEvents h.retryFailedAuditEvent",
	},
	"Settlement/internal/httpapi/api.go": {
		"READ_HISTORY":    "a.handoverResponsibilities a.handoverReceipt a.listOutboxEvents a.agingSummary a.dunningOverview a.dunningCoverage",
		"MUTATE_BUSINESS": "a.handoverFreeze a.handoverTransfer a.handoverComplete a.retryOutboxEvent a.completeDunningAction",
	},
}

func upgradeSemanticClassify(e *Entry, handler string) bool {
	for op, handlers := range upgradeSemanticGroups[e.Source] {
		for _, exact := range strings.Fields(handlers) {
			if handler == exact {
				e.Operation = op
				e.Reason = "Reviewed exact source/handler: " + handler + "; handover changes, approval recovery, credit/tax retries and new notifications remain business mutations; only existing reads and immutable audit delivery have narrower policies"
				return true
			}
		}
	}
	return false
}

// 包装层摘要无法保护导入的门禁、范围校验或调度逻辑。
// 这些条目仅作为审核输入，不新增必要操作的豁免范围。
var executionSources = []string{
	"contract_management/internal/transport/httpapi/license.go",
	"contract_management/internal/transport/httpapi/audit_buffer.go",
	"contract_management/internal/application/handover.go",
	"contract_management/internal/infrastructure/mysql/handover.go",
	"contract_management/internal/auditoutbox/gorm_store.go",
	"contract_management/internal/auditoutbox/dispatcher.go",
	"customer_and_opportunity/internal/commercial/gate.go",
	"customer_and_opportunity/internal/commercial/routes.go",
	"customer_and_opportunity/internal/modules/handover/service.go",
	"customer_and_opportunity/internal/modules/handover/repository.go",
	"customer_and_opportunity/internal/modules/handover/repository_transfers.go",
	"customer_and_opportunity/internal/shared/requestaudit/dispatcher.go",
	"project_management/internal/httpapi/commercial_license.go",
	"project_management/internal/httpapi/handover_integration.go",
	"project_management/internal/handover/service.go",
	"project_management/internal/auditoutbox/gorm_store.go",
	"project_management/internal/auditoutbox/dispatcher.go",
	"data_analysis/internal/commercial/gate.go",
	"data_analysis/internal/modules/dashboard/application/business.go",
	"data_analysis/internal/modules/dashboard/infrastructure/business.go",
	"data_analysis/internal/modules/dashboard/interfaces/http/business.go",
	"data_analysis/internal/aggregation/business.go",
	"data_analysis/internal/aggregation/queue_lease.go",
	"data_analysis/internal/auditoutbox/dispatcher.go",
	"data_analysis/internal/auditoutbox/gorm_store.go",
	"data_analysis/internal/alertworker/gorm_store.go",
	"data_analysis/internal/alertworker/worker.go",
	"data_analysis/internal/modules/alerts/application/service.go",
	"data_analysis/internal/modules/alerts/infrastructure/gorm_repository.go",
	"data_analysis/internal/modules/alerts/interfaces/http/handler.go",
	"Settlement/internal/httpapi/commercial_license.go",
	"Settlement/internal/httpapi/handover.go",
	"Settlement/internal/handover/service.go",
	"Settlement/internal/handover/repository.go",
	"Settlement/internal/service/operations.go",
	"Settlement/internal/service/credit_sync.go",
	"Settlement/internal/service/credit_sync_worker.go",
	"Settlement/internal/creditclient/client.go",
}

func fingerprintExecutionSources(root string, r *Registry) error {
	for _, source := range executionSources {
		raw, err := os.ReadFile(filepath.Join(root, source))
		if err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		r.SourceDigests[source] = hex.EncodeToString(sum[:])
	}
	return nil
}
