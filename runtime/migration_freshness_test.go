package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	core "github.com/J-S-Te/license-core"
)

// 迁移兼容旁路必须限制快照签发年龄：删除状态文件后回放旧迁移快照，
// 不得继续获得不受授权到期约束的放行。
func TestMigrationCompatibilitySnapshotFreshness(t *testing.T) {
	for _, tc := range []struct {
		name    string
		age     time.Duration
		allowed bool
	}{
		{"one-day-old", 24 * time.Hour, true},
		{"exactly-max-age", migrationSnapshotMaxAge, true},
		{"one-second-past-max-age", migrationSnapshotMaxAge + time.Second, false},
		{"thirty-one-days-old", 31 * 24 * time.Hour, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			s := f.snapshot()
			s.EnforcementState = Pending
			s.MigrationEligible = true
			s.CurrentLicenseJWS = ""
			s.PendingLicenseJWS = ""
			s.HighestLicenseVersion = 0
			s.SnapshotIssuedAt = f.now.Add(-tc.age).Unix()
			f.apply(s)
			err := f.r.Evaluate(context.Background(), core.MUTATE_BUSINESS, f.now)
			if tc.allowed {
				if err != nil {
					t.Fatalf("recent migration snapshot denied: %v", err)
				}
				return
			}
			if !errors.Is(err, core.ErrDenied) {
				t.Fatalf("stale migration snapshot allowed: %v", err)
			}
			// 失去兼容旁路后按常规无授权路径判定：需授权的读操作同样拒绝，
			// 仅 ESSENTIAL_SERVICE 保持可用。
			for _, operation := range []core.Operation{core.READ_HISTORY, core.EXPORT_HISTORY} {
				if err := f.r.Evaluate(context.Background(), operation, f.now); !errors.Is(err, core.ErrDenied) {
					t.Fatalf("stale migration snapshot allowed %s: %v", operation, err)
				}
			}
			if err := f.r.Evaluate(context.Background(), core.ESSENTIAL_SERVICE, f.now); err != nil {
				t.Fatalf("essential service blocked after migration snapshot expiry: %v", err)
			}
		})
	}
}
