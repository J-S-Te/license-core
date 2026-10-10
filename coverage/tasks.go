package coverage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
)

// Finite policies review execution, not the executable's transport/liveness.
// The implementation files are fingerprinted independently of command wrappers.
func reviewRemainingTasks(root string, r *Registry) error {
	commands := map[string]string{
		"notification-delivery-worker":           "ESSENTIAL_SERVICE",
		"opportunity-owner-notification-worker":  "MUTATE_BUSINESS",
		"presale-assignment-notification-worker": "MUTATE_BUSINESS",
		"presale-progress-notification-worker":   "MUTATE_BUSINESS",
	}
	for i := range r.Entries {
		e := &r.Entries[i]
		if e.Kind == "executable" && e.Application == "customer_and_opportunity" {
			if op, ok := commands[e.Registration]; ok {
				e.Operation = op
				e.Reason = "Reviewed task implementation below: existing notification delivery is retained; creating notification projections requires business entitlement"
			}
		}
		if e.Kind == "executable" && ((e.Application == "settlement" && e.Registration == "worker") || (e.Application == "project_management" && e.Registration == "sla-notifier")) {
			e.Operation = "MIXED_EXECUTION"
			e.Reason = "Task-level policies required; this process must remain running for historical exports and existing delivery"
		}
	}
	type task struct{ app, source, name, op, reason string }
	tasks := []task{
		{"customer_and_opportunity", "customer_and_opportunity/internal/notificationdeliveryworker/worker.go", "RunOnce", "ESSENTIAL_SERVICE", "Deliver previously persisted personal notifications and completion receipts only"},
		{"customer_and_opportunity", "customer_and_opportunity/internal/ownerchangenotificationworker/worker.go", "RunOnce", "MUTATE_BUSINESS", "Materialize owner-change event into notification records"},
		{"customer_and_opportunity", "customer_and_opportunity/internal/presaleassignmentnotificationworker/worker.go", "RunOnce", "MUTATE_BUSINESS", "Materialize assignment evidence into notification records"},
		{"customer_and_opportunity", "customer_and_opportunity/internal/presaleprogressnotificationworker/worker.go", "RunOnce", "MUTATE_BUSINESS", "Materialize progress evidence into notification records"},
		{"customer_and_opportunity", "customer_and_opportunity/internal/modules/portalinvite/service.go", "Verify", "ESSENTIAL_SERVICE", "Validate existing identity invitation; only expire invalid invitations and record verification audit, never consume or provision"},
	}
	for _, t := range tasks {
		raw, err := os.ReadFile(filepath.Join(root, t.source))
		if err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		r.SourceDigests[t.source] = hex.EncodeToString(sum[:])
		f, err := parser.ParseFile(token.NewFileSet(), t.source, raw, 0)
		if err != nil {
			return err
		}
		found := false
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if ok && fn.Name.Name == t.name && fn.Recv != nil {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("reviewed task missing %s:%s", t.source, t.name)
		}
		r.Entries = append(r.Entries, Entry{ID: t.app + ":" + t.source + ":reviewed-task:" + t.name, Application: t.app, Source: t.source, Registration: t.name, Kind: "task-policy", Operation: t.op, Reason: t.reason})
	}
	// These policies are keyed to actual task declarations already inventoried.
	// Their parent files and command wrappers are independently fingerprinted.
	parents := append([]Entry(nil), r.Entries...)
	for _, e := range parents {
		if e.Operation != "MIXED_EXECUTION" {
			continue
		}
		policies := map[string]string{}
		if e.Application == "settlement" {
			policies = map[string]string{"PLATFORM_AUDIT": "ESSENTIAL_SERVICE", "PLATFORM_NOTIFICATION": "ESSENTIAL_SERVICE", "TAX_INVOICE_COMMAND": "MUTATE_BUSINESS"}
			// Only this exact license consumer synchronizer is essential. Other Run
			// receivers continue to require review, even in the same process.
			policies["commercialGate.Run"] = "ESSENTIAL_SERVICE"
			if e.Kind == "executable" {
				policies["dunning.Scanner.RunOnce"] = "MUTATE_BUSINESS"
				policies["taxreconcile.Worker.RunOnce"] = "MUTATE_BUSINESS"
				policies["reportexport.Worker.RunOnce"] = "EXPORT_HISTORY"
			}
		} else if e.Application == "project_management" {
			policies = map[string]string{"Service.ScanSlaNotifications": "MUTATE_BUSINESS", "Service.DispatchNotificationOutbox": "ESSENTIAL_SERVICE"}
		}
		for name, op := range policies {
			r.Entries = append(r.Entries, Entry{ID: e.ID + ":policy:" + name, Application: e.Application, Source: e.Source, Kind: "task-policy", Registration: name, Operation: op, Reason: "Finite reviewed execution policy; gate before task claim and side effect, never extrapolate to unknown tasks/destinations"})
		}
		if err := discoverMixedTasks(root, r, e); err != nil {
			return err
		}
	}
	return nil
}

// Reconcile actual dispatch declarations/calls with the finite reviewed set.
// A new RunOnce call or destination never inherits a mixed-process exemption.
func discoverMixedTasks(root string, r *Registry, parent Entry) error {
	source := parent.Source
	if parent.Application == "settlement" {
		source = "Settlement/cmd/worker/main.go"
	}
	raw, err := os.ReadFile(filepath.Join(root, source))
	if err != nil {
		return err
	}
	sum := sha256.Sum256(raw)
	r.SourceDigests[source] = hex.EncodeToString(sum[:])
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, source, raw, 0)
	if err != nil {
		return err
	}
	knownCalls := map[string]bool{"scanner.RunOnce": true, "taxReconciler.RunOnce": true, "exporter.RunOnce": true, "worker.Run": true, "commercialGate.Run": true}
	if parent.Application == "project_management" {
		knownCalls = map[string]bool{"service.ScanSlaNotifications": true, "service.DispatchNotificationOutbox": true}
	}
	knownDest := map[string]bool{"PLATFORM_AUDIT": true, "PLATFORM_NOTIFICATION": true, "TAX_INVOICE_COMMAND": true}
	seenCalls, seenDest := map[string]bool{}, map[string]bool{}
	unknown := func(name string) {
		r.Entries = append(r.Entries, Entry{ID: parent.ID + ":discovered:" + name, Application: parent.Application, Source: source, Kind: "task-policy", Registration: name, Operation: "UNCLASSIFIED", Reason: "Actual mixed dispatch not present in finite reviewed task policy"})
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && (sel.Sel.Name == "RunOnce" || sel.Sel.Name == "ScanSlaNotifications" || sel.Sel.Name == "DispatchNotificationOutbox" || sel.Sel.Name == "Run") {
				name := expression(fs, sel)
				seenCalls[name] = true
				if !knownCalls[name] {
					unknown(name)
				}
			}
		}
		if lit, ok := n.(*ast.CompositeLit); ok && expression(fs, lit.Type) == "outbox.Destination" {
			for _, element := range lit.Elts {
				kv, ok := element.(*ast.KeyValueExpr)
				if !ok || expression(fs, kv.Key) != "Name" {
					continue
				}
				name, err := strconv.Unquote(expression(fs, kv.Value))
				if err != nil {
					unknown("dynamic-destination")
					continue
				}
				seenDest[name] = true
				if !knownDest[name] {
					unknown(name)
				}
			}
		}
		return true
	})
	for name := range knownCalls {
		if !seenCalls[name] {
			return fmt.Errorf("reviewed mixed task removed: %s %s", source, name)
		}
	}
	if parent.Application == "settlement" {
		for name := range knownDest {
			if !seenDest[name] {
				return fmt.Errorf("reviewed destination removed: %s", name)
			}
		}
	}
	return nil
}
