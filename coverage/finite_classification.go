package coverage

import "strings"

// 豁免依据精确源码及完整注册项，不按 GET 方法或整个路径放行健康检查。
// 匿名处理器的变更通过独立的源码摘要绑定。
var reviewedIdentityRegistrations = map[string]bool{
	"contract_management/internal/transport/httpapi/router.go|r.GET(\"/auth/login\") -> <anonymous>":                                                                                                                                                true,
	"contract_management/internal/transport/httpapi/router.go|r.GET(\"/auth/callback\") -> <anonymous>":                                                                                                                                             true,
	"contract_management/internal/transport/httpapi/router.go|r.GET(\"/auth/logout\") -> <anonymous>":                                                                                                                                               true,
	"contract_management/internal/transport/httpapi/router.go|r.GET(\"/logged-out\") -> h.loggedOut":                                                                                                                                                true,
	"contract_management/internal/transport/httpapi/router.go|r.POST(\"/auth/local-logout\") -> <anonymous>":                                                                                                                                        true,
	"contract_management/internal/transport/httpapi/router.go|r.POST(\"/auth/backchannel-logout\") -> <anonymous>":                                                                                                                                  true,
	"contract_management/internal/transport/httpapi/router.go|r.GET(\"/healthz\") -> <anonymous>":                                                                                                                                                   true,
	"contract_management/internal/transport/httpapi/router.go|r.GET(\"/readyz\") -> <anonymous>":                                                                                                                                                    true,
	"contract_management/internal/transport/httpapi/router.go|api.GET(\"/auth/me\") -> h.me":                                                                                                                                                        true,
	"customer_and_opportunity/internal/bootstrap/app.go|base.GET(\"/livez\") -> livenessHandler":                                                                                                                                                    true,
	"customer_and_opportunity/internal/bootstrap/app.go|base.GET(\"/readyz\") -> readinessHandler(databaseHealthy)":                                                                                                                                 true,
	"customer_and_opportunity/internal/bootstrap/app.go|base.GET(\"/healthz\") -> <anonymous>":                                                                                                                                                      true,
	"customer_and_opportunity/internal/bootstrap/app.go|base.GET(\"/auth/login\") -> authHandler.Login":                                                                                                                                             true,
	"customer_and_opportunity/internal/bootstrap/app.go|base.GET(\"/auth/callback\") -> authHandler.Callback":                                                                                                                                       true,
	"customer_and_opportunity/internal/bootstrap/app.go|base.POST(\"/auth/logout\") -> authHandler.Logout":                                                                                                                                          true,
	"customer_and_opportunity/internal/bootstrap/app.go|base.POST(\"/auth/local-logout\") -> authHandler.LocalLogout":                                                                                                                               true,
	"customer_and_opportunity/internal/bootstrap/app.go|base.POST(\"/auth/backchannel-logout\") -> crmauth.NewBackchannelLogoutHandler(oidcClient, authRepository, config.OIDCIssuer, config.OIDCClientID, config.OIDCBackchannelLogoutTTL).Handle": true,
	"customer_and_opportunity/internal/bootstrap/app.go|api.GET(\"/auth/me\") -> <anonymous>":                                                                                                                                                       true,
	"customer_and_opportunity/internal/portalbootstrap/router.go|base.GET(\"/livez\") -> <anonymous>":                                                                                                                                               true,
	"customer_and_opportunity/internal/portalbootstrap/router.go|base.GET(\"/readyz\") -> <anonymous>":                                                                                                                                              true,
	"customer_and_opportunity/internal/portalbootstrap/router.go|base.GET(\"/healthz\") -> <anonymous>":                                                                                                                                             true,
	"customer_and_opportunity/internal/portalbootstrap/router.go|base.GET(\"/auth/login\") -> beginLogin(deps)":                                                                                                                                     true,
	"customer_and_opportunity/internal/portalbootstrap/router.go|base.GET(\"/auth/callback\") -> completeLogin(deps)":                                                                                                                               true,
	"customer_and_opportunity/internal/portalbootstrap/router.go|base.POST(\"/auth/logout\") -> logout(deps)":                                                                                                                                       true,
	"customer_and_opportunity/internal/portalbootstrap/router.go|base.POST(\"/auth/backchannel-logout\") -> deps.BackchannelLogout":                                                                                                                 true,
	"customer_and_opportunity/internal/portalbootstrap/router.go|api.GET(\"/auth/me\") -> me":                                                                                                                                                       true,
	"data_analysis/internal/bootstrap/app.go|base.GET(\"/livez\") -> <anonymous>":                                                                                                                                                                   true,
	"data_analysis/internal/bootstrap/app.go|base.GET(\"/readyz\") -> <anonymous>":                                                                                                                                                                  true,
	"data_analysis/internal/bootstrap/app.go|base.GET(\"/healthz\") -> <anonymous>":                                                                                                                                                                 true,
	"data_analysis/internal/bootstrap/app.go|base.GET(\"/auth/login\") -> authHandler.Login":                                                                                                                                                        true,
	"data_analysis/internal/bootstrap/app.go|base.GET(\"/auth/callback\") -> authHandler.Callback":                                                                                                                                                  true,
	"data_analysis/internal/bootstrap/app.go|base.GET(\"/auth/logout\") -> authHandler.Logout":                                                                                                                                                      true,
	"data_analysis/internal/bootstrap/app.go|base.POST(\"/auth/backchannel-logout\") -> authService.BackchannelLogout":                                                                                                                              true,
	"data_analysis/internal/bootstrap/app.go|api.GET(\"/auth/me\") -> authHandler.AuthMe":                                                                                                                                                           true,
	"project_management/internal/httpapi/api.go|router.GET(\"/healthz\") -> <anonymous>":                                                                                                                                                            true,
	"project_management/internal/httpapi/api.go|router.GET(\"/readyz\") -> <anonymous>":                                                                                                                                                             true,
	"project_management/internal/httpapi/api.go|router.GET(\"/auth/login\") -> <anonymous>":                                                                                                                                                         true,
	"project_management/internal/httpapi/api.go|router.GET(\"/auth/callback\") -> <anonymous>":                                                                                                                                                      true,
	"project_management/internal/httpapi/api.go|router.GET(\"/auth/logout\") -> <anonymous>":                                                                                                                                                        true,
	"project_management/internal/httpapi/api.go|router.POST(\"/auth/local-logout\") -> <anonymous>":                                                                                                                                                 true,
	"project_management/internal/httpapi/api.go|router.POST(\"/auth/backchannel-logout\") -> <anonymous>":                                                                                                                                           true,
	"project_management/internal/httpapi/api.go|router.GET(\"/logged-out\") -> loggedOut":                                                                                                                                                           true,
	"project_management/internal/httpapi/api.go|api.GET(\"/auth/me\") -> h.me":                                                                                                                                                                      true,
}

var basicHandlerGroups = map[string]map[string]string{
	"contract_management/internal/transport/httpapi/router.go": {
		"READ_HISTORY":    "h.previewTemplate h.countSignedContractsByOpportunity",
		"EXPORT_HISTORY":  "h.exportContract h.downloadApprovedDOCX h.downloadApprovedPDF h.downloadStampedPDF",
		"MUTATE_BUSINESS": "h.createContract h.createExternalContract h.submitApproval h.activateContract",
	},
	"project_management/internal/httpapi/api.go": {
		"EXPORT_HISTORY":  "h.exportCapabilities",
		"MUTATE_BUSINESS": "h.activateContract h.submitFieldRecord h.completeServiceItemField",
	},
	"customer_and_opportunity/internal/portalbootstrap/router.go": {
		"EXPORT_HISTORY":  "createProjectExport(deps) createProjectExportGrant(deps) downloadProjectExport(deps) createReportGrant(deps) downloadReport(deps)",
		"READ_HISTORY":    "reconciliationSnapshot(deps)",
		"MUTATE_BUSINESS": "createReport(deps) provision(deps) syncProjectAccess(deps)",
	},
}

func basicHandlerOperation(source, handler string) string {
	for op, handlers := range basicHandlerGroups[source] {
		for _, exact := range strings.Fields(handlers) {
			if handler == exact {
				return op
			}
		}
	}
	return ""
}
