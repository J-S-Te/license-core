package coverage

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSettlementDestinationConstructorIsScanned 核验拆分构造器仍受有限登记门禁约束。
// 参数：t 为当前测试上下文，夹具仅写入独立临时目录。
// 返回值：无；漏扫、漏指纹、未知项被豁免或已知项丢失未阻断时失败。
func TestSettlementDestinationConstructorIsScanned(t *testing.T) {
	const source = "Settlement/cmd/worker/main.go"
	const helper = "Settlement/internal/outbox/destinations.go"
	const worker = `package main
import "github.com/j-s-te/settlement/internal/outbox"
func main() {
 scanner.RunOnce(ctx); taxReconciler.RunOnce(ctx); exporter.RunOnce(ctx)
 worker.Run(ctx); commercialGate.Run(ctx, nil); creditWorker.RunOnce(ctx)
 plan, _ := outbox.PlanPlatformDestinations(input)
 destinations = append(destinations, plan.Destinations...)
 _ = outbox.Destination{Name:"TAX_INVOICE_COMMAND"}
}`
	const constructor = `package outbox
const DestinationPlatformAudit = "PLATFORM_AUDIT"
const DestinationPlatformNotification = "PLATFORM_NOTIFICATION"
func PlanPlatformDestinations(in PlatformDestinationInput) (PlatformDestinationPlan, error) {
 var plan PlatformDestinationPlan
 plan.Destinations = append(plan.Destinations, Destination{Name: DestinationPlatformAudit})
 plan.Destinations = append(plan.Destinations, Destination{Name: DestinationPlatformNotification})
 EXTRA
 return plan, nil
}
`
	for _, tc := range []struct {
		name, extra, replace, wantError, wantUnknown                     string
		omitHelper, wrongImport, omitCall, shadowedImport, omitWorkerRun bool
	}{
		{name: "literal-constants"},
		{name: "unknown-destination", extra: `_ = Destination{Name:"UNKNOWN_BUSINESS"}`, wantUnknown: "UNKNOWN_BUSINESS"},
		{name: "dynamic-destination", extra: `_ = Destination{Name:readName()}`, wantUnknown: "dynamic-destination"},
		{name: "shadowed-variable", extra: `{ DestinationPlatformAudit := readName(); _ = Destination{Name:DestinationPlatformAudit} }`, wantUnknown: "dynamic-destination"},
		{name: "shadowed-local-constant", extra: `{ const DestinationPlatformAudit = "UNKNOWN_BUSINESS"; _ = Destination{Name:DestinationPlatformAudit} }`, wantUnknown: "dynamic-destination"},
		{name: "unknown-task", extra: `newBusiness.RunOnce(ctx)`, wantUnknown: "newBusiness.RunOnce"},
		{name: "helper-worker-is-not-main-worker", extra: `worker.Run(ctx)`, wantUnknown: "worker.Run"},
		{name: "helper-gate-is-not-main-gate", extra: `commercialGate.Run(ctx, nil)`, wantUnknown: "commercialGate.Run"},
		{name: "helper-cannot-restore-main-task", extra: `worker.Run(ctx)`, omitWorkerRun: true, wantError: "reviewed mixed task removed"},
		{name: "missing-known-destination", replace: "REMOVED_AUDIT", wantError: "reviewed destination removed: PLATFORM_AUDIT"},
		{name: "unused-function-cannot-restore-destination", replace: "REMOVED_AUDIT", extra: "", wantError: "reviewed destination removed: PLATFORM_AUDIT"},
		{name: "missing-helper", omitHelper: true, wantError: "destinations.go"},
		{name: "wrong-package-import", wrongImport: true, wantError: "reviewed destination removed"},
		{name: "unused-helper", omitCall: true, wantError: "reviewed destination removed"},
		{name: "shadowed-import", shadowedImport: true, wantError: "reviewed destination removed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			mainCode := worker
			if tc.wrongImport {
				mainCode = strings.ReplaceAll(mainCode, "github.com/j-s-te/settlement/internal/outbox", "example.invalid/other/outbox")
			}
			if tc.omitCall {
				mainCode = strings.ReplaceAll(mainCode, "outbox.PlanPlatformDestinations(input)", "other.PlanPlatformDestinations(input)")
			}
			if tc.shadowedImport {
				mainCode = strings.ReplaceAll(mainCode, "plan, _ :=", "outbox := otherDestinationPlanner{}; plan, _ :=")
			}
			if tc.omitWorkerRun {
				mainCode = strings.ReplaceAll(mainCode, "worker.Run(ctx);", "")
			}
			helperCode := strings.ReplaceAll(constructor, "EXTRA", tc.extra)
			if tc.replace != "" {
				helperCode = strings.ReplaceAll(helperCode, `Destination{Name: DestinationPlatformAudit}`, `Destination{Name: "`+tc.replace+`"}`)
			}
			if tc.name == "unused-function-cannot-restore-destination" {
				helperCode += `func unused() { _ = Destination{Name: DestinationPlatformAudit} }`
			}
			for name, code := range map[string]string{source: mainCode, helper: helperCode} {
				if name == helper && tc.omitHelper {
					continue
				}
				path := filepath.Join(root, name)
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(code), 0600); err != nil {
					t.Fatal(err)
				}
			}
			r := Registry{SourceDigests: map[string]string{}}
			parent := Entry{ID: "mixed", Application: "settlement", Source: source, Operation: "MIXED_EXECUTION"}
			err := discoverMixedTasks(root, &r, parent)
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("want error %q, got %v", tc.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256([]byte(helperCode))
			if r.SourceDigests[helper] != hex.EncodeToString(sum[:]) {
				t.Fatal("actual destination constructor source is not fingerprinted")
			}
			if tc.wantUnknown == "" {
				if len(r.Entries) != 0 {
					t.Fatalf("reviewed destinations unexpectedly unresolved: %+v", r.Entries)
				}
				return
			}
			if len(r.Entries) != 1 || r.Entries[0].Registration != tc.wantUnknown ||
				r.Entries[0].Operation != "UNCLASSIFIED" || r.Entries[0].Source != helper {
				t.Fatalf("unknown constructor behavior escaped source review: %+v", r.Entries)
			}
		})
	}
}
