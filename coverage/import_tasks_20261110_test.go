package coverage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNovemberImportTaskRemovalClosed 用途：禁止已审核后台方法移除后悄悄缩小分母。
// 参数：t 为临时源码夹具测试上下文。
// 返回值：无；完整四任务缺失、误分类或方法删除后继续放行时失败。
func TestNovemberImportTaskRemovalClosed(t *testing.T) {
	root := t.TempDir()
	r := Registry{SourceDigests: map[string]string{}}
	for _, task := range reviewedImportTasks20261110 {
		raw, err := os.ReadFile(filepath.Join("../..", task.Source))
		if os.IsNotExist(err) {
			t.Skip("requires workspace application sources")
		}
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, task.Source)
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
		r.SourceDigests[task.Source] = "fixture"
	}
	if err := appendImportTasks20261110(root, &r); err != nil {
		t.Fatal(err)
	}
	if len(r.Entries) != 4 {
		t.Fatalf("missing reviewed tasks: %d", len(r.Entries))
	}
	for _, e := range r.Entries {
		if e.Operation != "MUTATE_BUSINESS" {
			t.Fatalf("business import exempted: %+v", e)
		}
	}
	path := filepath.Join(root, "contract_management/internal/application/contract_import_dispatcher.go")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(string(raw), "func (s *Service) DispatchImportOnce(", "func (s *Service) RemovedImportOnce(", 1)
	if changed == string(raw) {
		t.Fatal("fixture method anchor absent")
	}
	if err = os.WriteFile(path, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	if err = appendImportTasks20261110(root, &r); err == nil {
		t.Fatal("removed import task was accepted")
	}
}
