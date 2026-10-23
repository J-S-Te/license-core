// Package coverage 盘点源码中的注册项，用于发布门禁。
// 它不提供运行时授权；符号路由与混合职责工作进程仍须另行确认。
package coverage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Entry struct {
	ID           string `json:"id"`
	Application  string `json:"application"`
	Source       string `json:"source"`
	Registration string `json:"registration"`
	Kind         string `json:"kind"`
	Operation    string `json:"operation"`
	Reason       string `json:"reason"`
}
type Registry struct {
	Schema        int               `json:"schema"`
	Scope         string            `json:"scope"`
	SourceDigests map[string]string `json:"source_digests"`
	Entries       []Entry           `json:"entries"`
}

var files = map[string][]string{
	"contract_management":      {"contract_management/internal/transport/httpapi/router.go", "contract_management/internal/transport/httpapi/handover.go"},
	"project_management":       {"project_management/internal/httpapi/api.go"},
	"settlement":               {"Settlement/internal/httpapi/api.go"},
	"data_analysis":            {"data_analysis/internal/bootstrap/app.go"},
	"customer_portal":          {"customer_and_opportunity/internal/portalbootstrap/router.go"},
	"customer_and_opportunity": {"customer_and_opportunity/internal/bootstrap/app.go", "customer_and_opportunity/internal/modules/handover/handler.go"},
}

func expression(f *token.FileSet, n ast.Node) string {
	var b bytes.Buffer
	if err := format.Node(&b, f, n); err != nil {
		return "<invalid>"
	}
	return b.String()
}

// Scan 统计源码中的注册表达式，不统计展开后的实际路由。
// 循环、条件功能与接收器分组仍须通过运行时测试确认。
func Scan(root string) (Registry, error) {
	r := Registry{Schema: 1, Scope: "source-registration-expressions; not live route denominator", SourceDigests: map[string]string{}}
	inputs := map[string][]string{}
	for app, list := range files {
		inputs[app] = append([]string{}, list...)
	}
	routes, err := filepath.Glob(filepath.Join(root, "customer_and_opportunity/internal/modules/*/routes.go"))
	if err != nil {
		return r, err
	}
	for _, p := range routes {
		rel, _ := filepath.Rel(root, p)
		inputs["customer_and_opportunity"] = append(inputs["customer_and_opportunity"], rel)
	}
	for app, list := range inputs {
		for _, rel := range list {
			raw, err := os.ReadFile(filepath.Join(root, rel))
			if err != nil {
				return r, err
			}
			sum := sha256.Sum256(raw)
			r.SourceDigests[rel] = hex.EncodeToString(sum[:])
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, rel, raw, 0)
			if err != nil {
				return r, err
			}
			ordinal := 0
			ast.Inspect(f, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if !strings.Contains("|GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS|Any|Handle|HandleFunc|", "|"+sel.Sel.Name+"|") || len(call.Args) == 0 {
					return true
				}
				// 未匹配到选择器时按 HTTP 注册处理。仅扫描已审核的路由文件，
				// 并记录无法解析的选择器，避免静默遗漏。
				ordinal++
				path := expression(fset, call.Args[0])
				handler := "<missing>"
				if len(call.Args) > 1 {
					if _, ok := call.Args[len(call.Args)-1].(*ast.FuncLit); ok {
						handler = "<anonymous>"
					} else {
						handler = expression(fset, call.Args[len(call.Args)-1])
					}
				}
				e := Entry{ID: fmt.Sprintf("%s:%s:%03d", app, rel, ordinal), Application: app, Source: rel, Registration: expression(fset, sel.X) + "." + sel.Sel.Name + "(" + path + ") -> " + handler, Kind: "http", Operation: "UNCLASSIFIED", Reason: "Requires handler semantic review and expanded live route comparison"}
				classify(&e, sel.Sel.Name, path, handler)
				semanticClassify(&e, handler)
				if app == "data_analysis" && handler == "<anonymous>" {
					key := sel.Sel.Name + " " + path
					for _, reviewed := range []string{`POST "/alerts/:id/ack"`, `POST "/alerts/:id/close"`, `POST "/admin/sources/:id/trigger"`, `DELETE "/alert-rules/:id"`} {
						if key == reviewed {
							e.Operation = "MUTATE_BUSINESS"
							e.Reason = "Explicit anonymous handler changes alert lifecycle, rule definitions or queues new aggregation"
						}
					}
				}
				if app == "customer_portal" && path == `"/activate"` {
					e.Operation = "MUTATE_BUSINESS"
					e.Reason = "GET starts invite activation/login flow; not a historical read"
				}
				r.Entries = append(r.Entries, e)
				return true
			})
			// Settlement 直接分派请求，不通过 Gin 路由注册。
			if app == "settlement" {
				for _, decl := range f.Decls {
					fn, ok := decl.(*ast.FuncDecl)
					if !ok || fn.Name.Name != "ServeHTTP" {
						continue
					}
					ast.Inspect(fn.Body, func(n ast.Node) bool {
						cl, ok := n.(*ast.CaseClause)
						if !ok {
							return true
						}
						for _, cond := range cl.List {
							ordinal++
							e := Entry{ID: fmt.Sprintf("%s:%s:%03d", app, rel, ordinal), Application: app, Source: rel, Kind: "http-dispatch", Registration: expression(fset, cond), Operation: "UNCLASSIFIED", Reason: "Dispatch predicate requires explicit semantic classification; not a method-wide policy"}
							for _, statement := range cl.Body {
								ast.Inspect(statement, func(n ast.Node) bool {
									c, ok := n.(*ast.CallExpr)
									if !ok {
										return true
									}
									s, ok := c.Fun.(*ast.SelectorExpr)
									if !ok {
										return true
									}
									if expression(fset, s.X) == "a" {
										semanticClassify(&e, "a."+s.Sel.Name)
									}
									return true
								})
							}
							for _, exact := range []string{`r.URL.Path == "/healthz"`, `r.URL.Path == "/readyz"`, `r.URL.Path == "/auth/callback"`, `r.URL.Path == "/auth/logout"`, `r.URL.Path == "/auth/local-logout" && r.Method == http.MethodPost`, `r.URL.Path == "/auth/backchannel-logout" && r.Method == http.MethodPost`, `r.URL.Path == "/logged-out"`} {
								if e.Registration == exact {
									e.Operation = "ESSENTIAL_SERVICE"
									e.Reason = "Exact health or auth/session dispatch predicate"
								}
							}
							r.Entries = append(r.Entries, e)
						}
						return true
					})
				}
			}
		}
	}
	// 向 Temporal 注册结构体并不会逐项列出其活动。
	// 该条目保留为待确认状态，不能把一次注册视为完整覆盖。
	taskFiles := map[string][]string{
		"contract_management":      {"contract_management/internal/workflows/register.go", "contract_management/internal/workflows/activities.go"},
		"customer_and_opportunity": {"customer_and_opportunity/internal/presaleworkflow/types.go"},
		"project_management":       {"project_management/internal/application/delivery.go", "project_management/internal/application/delivery_effects.go"},
		"settlement":               {"Settlement/internal/reportexport/worker.go", "Settlement/internal/outbox/worker.go", "Settlement/internal/dunning/scanner.go", "Settlement/internal/taxreconcile/worker.go"},
		"data_analysis":            {"data_analysis/internal/aggregation/worker.go", "data_analysis/internal/aggregation/aggregation.go", "data_analysis/internal/alertworker/worker.go"},
	}
	for app, list := range taskFiles {
		for _, rel := range list {
			raw, err := os.ReadFile(filepath.Join(root, rel))
			if err != nil {
				return r, err
			}
			sum := sha256.Sum256(raw)
			r.SourceDigests[rel] = hex.EncodeToString(sum[:])
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, rel, raw, 0)
			if err != nil {
				return r, err
			}
			ordinal := 0
			ast.Inspect(f, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if sel.Sel.Name != "RegisterWorkflowWithOptions" && sel.Sel.Name != "RegisterActivity" {
					return true
				}
				ordinal++
				op := "MUTATE_BUSINESS"
				reason := "Approval/status/archive/worklog workflow is new business; check each activity execution before side effects"
				if sel.Sel.Name == "RegisterActivity" {
					reason = "Registered approval/worklog activity struct contains business side effects; individual activity registration comparison remains required"
				}
				r.Entries = append(r.Entries, Entry{ID: fmt.Sprintf("%s:%s:temporal:%03d", app, rel, ordinal), Application: app, Source: rel, Kind: "temporal-registration", Registration: expression(fset, call), Operation: op, Reason: reason})
				return true
			})
			for _, decl := range f.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				name := fn.Name.Name
				if fn.Recv != nil && len(fn.Recv.List) > 0 && expression(fset, fn.Recv.List[0].Type) == "*Activities" {
					r.Entries = append(r.Entries, Entry{ID: app + ":" + rel + ":activity:" + name, Application: app, Source: rel, Kind: "activity-method", Registration: "*Activities." + name, Operation: "MUTATE_BUSINESS", Reason: "Exported registered activity delegates approval, status/archive, notification generation or worklog publication persistence; check before side effects"})
				}
				if name != "RunOnce" && name != "ScanSlaNotifications" && name != "DispatchNotificationOutbox" && name != "DispatchDeliveryEffects" {
					continue
				}
				op := "MUTATE_BUSINESS"
				reason := "Check before task claim and every business side effect; retain retryable pending work"
				if strings.Contains(rel, "reportexport/") {
					op = "EXPORT_HISTORY"
					reason = "Historical CSV/file generation and required export job bookkeeping only"
				}
				if strings.Contains(rel, "outbox/") {
					op = "MIXED_EXECUTION"
					reason = "Destination-level policies below are mandatory; no whole-worker essential exemption"
				}
				if name == "DispatchNotificationOutbox" {
					op = "ESSENTIAL_SERVICE"
					reason = "Only claims and delivers existing immutable notification outbox records; SLA generation separately gated"
				}
				if name == "DispatchDeliveryEffects" {
					reason = "Creates automation/warning delivery events and new notification intents from queued effects; business license required before claim, not existing immutable notification delivery"
				}
				receiver := ""
				if fn.Recv != nil && len(fn.Recv.List) > 0 {
					receiver = expression(fset, fn.Recv.List[0].Type)
				}
				r.Entries = append(r.Entries, Entry{ID: app + ":" + rel + ":task:" + receiver + name, Application: app, Source: rel, Kind: "task-boundary", Registration: receiver + "." + name, Operation: op, Reason: reason})
			}
		}
	}
	// 盘点所有可执行命令，包括维护命令与手动工作进程模式。
	// 不能仅因 Compose 未包含某命令就认定其安全。
	roots := map[string]string{"contract_management": "contract_management", "project_management": "project_management", "settlement": "Settlement", "data_analysis": "data_analysis", "customer_and_opportunity": "customer_and_opportunity"}
	for app, dir := range roots {
		commands, err := filepath.Glob(filepath.Join(root, dir, "cmd/*/main.go"))
		if err != nil {
			return r, err
		}
		for _, p := range commands {
			rel, _ := filepath.Rel(root, p)
			raw, err := os.ReadFile(p)
			if err != nil {
				return r, err
			}
			sum := sha256.Sum256(raw)
			r.SourceDigests[rel] = hex.EncodeToString(sum[:])
			name := filepath.Base(filepath.Dir(p))
			owner := app
			if strings.HasPrefix(name, "portal-") {
				owner = "customer_portal"
			}
			e := Entry{ID: owner + ":" + rel, Application: owner, Source: rel, Kind: "executable", Registration: name, Operation: "UNCLASSIFIED", Reason: "Executable requires task-level policy; maintenance is not automatically essential"}
			for _, reviewed := range []string{"api", "crm-server", "portal-server", "dashboard-api", "worker-rollout", "presale-worker-rollout", "authz-catalog", "catalog-sync", "migrate", "local-migrate", "production-migrate", "migration-plan", "openapi-check"} {
				if name == reviewed {
					e.Operation = "ESSENTIAL_SERVICE"
					e.Reason = "Executable startup/approved maintenance only; this does not exempt its separately inventoried business HTTP/tasks"
				}
			}
			for _, reviewed := range []string{"file-backfill", "import-templates", "seed-demo-data", "seed-portal-invite", "presale-integration-mock", "opportunity-alert-worker", "presale-alert-worker", "presale-engineer-sync-worker", "presale-worker", "portal-project-worker", "portal-report-worker", "portal-feedback-worker", "portal-invite-compensation-worker", "portal-filing-worker"} {
				if name == reviewed {
					e.Operation = "MUTATE_BUSINESS"
					e.Reason = "Explicit executable creates or synchronizes business/identity records; startup alone is not task authorization"
				}
			}
			if name == "portal-access-disable-worker" {
				e.Operation = "ESSENTIAL_SERVICE"
				e.Reason = "Only processes existing security revocation/disable tasks, not account provisioning"
			}
			if app == "data_analysis" && name == "sync-job-recover-legacy" {
				e.Operation = "ESSENTIAL_SERVICE"
				e.Reason = "Exact audited maintenance executable: default dry-run; apply requires root, local Docker stopped-worker proof, deployment lock and generation-zero legacy rows. It only restores queue eligibility; licensed worker separately gates every claim and business write"
			}
			if name == "worker" && app == "contract_management" {
				e.Operation = "MUTATE_BUSINESS"
				e.Reason = "Approval/status/archive worker activities and daily archive scheduling are business mutations"
			}
			if name == "worker" && app == "project_management" {
				e.Operation = "ESSENTIAL_SERVICE"
				e.Reason = "Current worker has no workflows/activities registered; startup metrics/poller only; cluster old tasks require separate evidence"
			}
			if name == "portal-project-export-worker" {
				e.Operation = "EXPORT_HISTORY"
				e.Reason = "Existing project snapshot PDF export; identity and data scope remain required"
			}
			if name == "aggregation-worker" || name == "alert-worker" || name == "contract-transfer-worker" || name == "presale-report-aggregate-worker" {
				e.Operation = "MUTATE_BUSINESS"
				e.Reason = "Generates or changes business projections, alerts or handoffs; check before claiming a task"
			}
			if name == "sla-notifier" || (name == "worker" && (app == "settlement" || app == "contract_management")) {
				e.Reason = "Mixed executable: gate business tasks separately; preserve existing audit delivery and history exports"
			}
			r.Entries = append(r.Entries, e)
		}
	}
	if err := reviewRemainingTasks(root, &r); err != nil {
		return r, err
	}
	if err := fingerprintExecutionSources(root, &r); err != nil {
		return r, err
	}
	sort.Slice(r.Entries, func(i, j int) bool { return r.Entries[i].ID < r.Entries[j].ID })
	return r, nil
}

func classify(e *Entry, method, path, handler string) {
	// 经操作者批准的策略：BI 嵌入可能触发上游实时计算；
	// 授权到期后仅保留普通历史查询与导出，不保留嵌入能力。
	if e.Source == "data_analysis/internal/bootstrap/app.go" && method == "GET" &&
		(e.Registration == `api.GET("/embed/:dashboard") -> <anonymous>` || e.Registration == `embedProxy.GET("/:token") -> <anonymous>` || e.Registration == `embedProxy.GET("/:token/*resource") -> <anonymous>`) {
		e.Operation = "MUTATE_BUSINESS"
		e.Reason = "User-approved expiry policy: prohibit BI embedding and proxy resources; preserve separately classified history reads and exports"
		return
	}
	// 仅放行已审核的精确路径，不按所有 GET 请求或处理器名称推断豁免。
	if reviewedIdentityRegistrations[e.Source+"|"+e.Registration] {
		e.Operation = "ESSENTIAL_SERVICE"
		e.Reason = "Exact health or identity/session lifecycle endpoint; authentication remains independent"
		return
	}
	if op := basicHandlerOperation(e.Source, handler); op != "" {
		e.Operation = op
		e.Reason = "Explicit handler reviewed in commercial-license module entry matrix; retain ordinary scope checks"
	}
	if e.Source == "customer_and_opportunity/internal/portalbootstrap/router.go" && handler == "exportFiling(deps)" {
		e.Operation = "EXPORT_HISTORY"
		e.Reason = "Reviewed existing submitted filing snapshot PDF export with version/SHA verification and export audit only; identity, data scope and runtime license remain required"
	}
	_ = method
	_ = path
}

// Check 在源码过期、条目未知或覆盖范围未确认时拒绝放行。
func Check(current, saved Registry) error {
	a, _ := json.Marshal(current.SourceDigests)
	b, _ := json.Marshal(saved.SourceDigests)
	if !bytes.Equal(a, b) {
		return fmt.Errorf("source drift: regenerate inventory and review changed handlers")
	}
	if len(current.Entries) != len(saved.Entries) {
		return fmt.Errorf("registration denominator changed")
	}
	known := map[string]bool{}
	for _, e := range saved.Entries {
		if known[e.ID] {
			return fmt.Errorf("duplicate entry %s", e.ID)
		}
		known[e.ID] = true
	}
	unresolved := 0
	for i, e := range saved.Entries {
		if e.ID != current.Entries[i].ID || e.Registration != current.Entries[i].Registration {
			return fmt.Errorf("registration drift %s", e.ID)
		}
		if (e.Kind == "task-policy" || e.Operation == "MIXED_EXECUTION") && e.Operation != current.Entries[i].Operation {
			return fmt.Errorf("task policy differs from reviewed execution %s", e.ID)
		}
		switch e.Operation {
		case "READ_HISTORY", "EXPORT_HISTORY", "MUTATE_BUSINESS", "ESSENTIAL_SERVICE":
			if strings.TrimSpace(e.Reason) == "" {
				return fmt.Errorf("missing rationale %s", e.ID)
			}
		case "MIXED_EXECUTION":
			found := false
			for _, child := range saved.Entries {
				if strings.HasPrefix(child.ID, e.ID+":policy:") && child.Kind == "task-policy" {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("mixed entry has no task policy %s", e.ID)
			}
		default:
			unresolved++
		}
	}
	if unresolved > 0 {
		return fmt.Errorf("coverage incomplete: %d of %d source expressions/executables unclassified; live route/activity denominator still required", unresolved, len(saved.Entries))
	}
	return nil
}
