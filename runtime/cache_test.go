package runtime

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	core "github.com/J-S-Te/license-core"
)

// warmCachedRead 预热读路径，并断言第二次评估没有再次执行验签。
// 其余用例都在预热后执行安全、时钟与并发断言，
// 用于确认读路径缓存只跳过重复的 strict + VerifySnapshot，
// 不会绕过文件元数据校验，也不会改变任何可见判定语义。
func (f *fixture) warmCachedRead(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now); err != nil {
		t.Fatal(err)
	}
	verified := f.r.verifications.Load()
	if verified == 0 {
		t.Fatal("first evaluation performed no verification")
	}
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now); err != nil {
		t.Fatal(err)
	}
	if again := f.r.verifications.Load(); again != verified {
		t.Fatalf("preheated read re-verified state: verifications %d -> %d", verified, again)
	}
}

// 预热后收紧状态文件权限，缓存不得让读路径跳过 0600 校验。
func TestCachedReadStillRejectsInsecureStatePermissions(t *testing.T) {
	f := newFixture(t)
	f.apply(f.snapshot())
	f.warmCachedRead(t)
	if err := os.Chmod(f.path, 0644); err != nil {
		t.Fatal(err)
	}
	if err := f.r.Evaluate(context.Background(), core.MUTATE_BUSINESS, f.now); !errors.Is(err, ErrState) {
		t.Fatalf("cached read accepted insecure state permissions: %v", err)
	}
}

// 预热后把状态文件替换为符号链接，缓存不得让读路径跳过 O_NOFOLLOW 校验。
func TestCachedReadStillRejectsSymlinkedState(t *testing.T) {
	f := newFixture(t)
	f.apply(f.snapshot())
	f.warmCachedRead(t)
	target := filepath.Join(filepath.Dir(f.path), "target.json")
	if err := os.Rename(f.path, target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, f.path); err != nil {
		t.Fatal(err)
	}
	if err := f.r.Evaluate(context.Background(), core.MUTATE_BUSINESS, f.now); !errors.Is(err, ErrState) {
		t.Fatalf("cached read accepted symlinked state: %v", err)
	}
}

// 预热后出现超过 300s 的时钟回拨，仍须返回 ErrClock 并把闩锁写入磁盘；
// 进程重启后闩锁必须继续生效。
func TestCachedReadStillLatchesClockRollback(t *testing.T) {
	f := newFixture(t)
	f.apply(f.snapshot())
	f.warmCachedRead(t)
	ctx := context.Background()
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(-301*time.Second)); !errors.Is(err, ErrClock) {
		t.Fatalf("cached read accepted clock rollback: %v", err)
	}
	f.r = f.restart()
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now); !errors.Is(err, ErrClock) {
		t.Fatalf("restart cleared the durable clock latch: %v", err)
	}
}

// 预热后推进评估时间，缓存不得阻止高水位落盘，ReadState 必须读到新值。
func TestCachedReadPersistsClockHighWater(t *testing.T) {
	f := newFixture(t)
	f.apply(f.snapshot())
	f.warmCachedRead(t)
	ctx := context.Background()
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	state, err := f.r.ReadState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.ClockHighWater != f.now.Unix()+1 {
		t.Fatalf("clock high water = %d, want %d", state.ClockHighWater, f.now.Unix()+1)
	}
}

// 同一运行时上并发评估与落盘，缓存结构必须无数据竞争，
// 且高水位仍取所有评估时间的最大值。
func TestConcurrentEvaluateAndApplySnapshotWithCache(t *testing.T) {
	f := newFixture(t)
	f.apply(f.snapshot())
	f.warmCachedRead(t)
	raw := f.sign(f.snapshot())
	ctx := context.Background()
	const workers = 30
	var wg sync.WaitGroup
	failures := make(chan error, workers*2)
	for i := 0; i < workers; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			failures <- f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(time.Duration(i)*time.Second))
		}(i)
		go func() {
			defer wg.Done()
			failures <- f.r.ApplySnapshot(ctx, raw, f.now)
		}()
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatalf("concurrent evaluate/apply: %v", err)
		}
	}
	state, err := f.r.ReadState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.ClockHighWater != f.now.Unix()+workers-1 {
		t.Fatalf("clock high water = %d, want %d", state.ClockHighWater, f.now.Unix()+workers-1)
	}
}

// entryCount 仅用于测试观察缓存条目数量。
func (c *verificationCache) entryCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

// 任何 save() 落盘后，进程内缓存的验签结果必须失效。
func TestSaveInvalidatesLoadCache(t *testing.T) {
	f := newFixture(t)
	f.apply(f.snapshot())
	f.warmCachedRead(t)
	if got := f.r.cache.entryCount(); got == 0 {
		t.Fatal("preheated read did not cache the verification result")
	}
	if err := f.r.Evaluate(context.Background(), core.MUTATE_BUSINESS, f.now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if got := f.r.cache.entryCount(); got != 0 {
		t.Fatalf("save left %d cached verification results", got)
	}
}

// 预热后由另一运行时写入更高修订，原运行时不得沿用旧字节的验签结果。
func TestCachedReadSeesExternalStateChange(t *testing.T) {
	f := newFixture(t)
	f.apply(f.snapshot())
	f.warmCachedRead(t)
	next := f.snapshot()
	next.Revision = 2
	next.SnapshotIssuedAt = f.now.Unix() + 1
	other := f.restart()
	if err := other.ApplySnapshot(context.Background(), f.sign(next), f.now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	state, err := f.r.ReadState(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if state.ControlRevision != 2 {
		t.Fatalf("control revision = %d, want 2", state.ControlRevision)
	}
}
