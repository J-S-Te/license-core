package coverage

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReviewedSplitDispatchKeepsMethodPolicies 验证混合辅助分派不会把业务写入归入历史读取。
// 参数：t 为隔离AST测试上下文。
// 返回值：无；方法策略缺失或未知分支被豁免时失败。
func TestReviewedSplitDispatchKeepsMethodPolicies(t *testing.T) {
	source := "Settlement/internal/httpapi/role_finance.go"
	code := `package httpapi
func (a *API) serveRoleFinance(w W,r *R)bool{switch{case path=="settings":a.roleSettings(w,r);case path=="unreviewed":a.roleSettings(w,r)};return true}`
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, source, code, 0)
	if err != nil {
		t.Fatal(err)
	}
	entries := settlementDispatchEntries(fs, f, source, 0)
	if len(entries) != 4 {
		t.Fatalf("expected mixed parent, two exact policies and unknown predicate: %+v", entries)
	}
	ops := map[string]string{}
	for _, e := range entries {
		ops[e.Registration] = e.Operation
	}
	if ops[`path == "settings"`] != "MIXED_EXECUTION" || ops["GET /api/v1/settings"] != "READ_HISTORY" || ops["PUT /api/v1/settings"] != "MUTATE_BUSINESS" || ops[`path == "unreviewed"`] != "UNCLASSIFIED" {
		t.Fatalf("wrong finite policies: %+v", ops)
	}
	r := Registry{Entries: entries}
	if Check(r, r) == nil {
		t.Fatal("unknown split predicate passed gate")
	}
}

// TestTemporalActivityOptionsIsAnInventoryBoundary 验证带选项活动注册被盘点而非静默遗漏。
// 参数：t 为隔离AST测试上下文。
// 返回值：无；活动注册被漏扫或注册身份丢失时失败。
func TestTemporalActivityOptionsIsAnInventoryBoundary(t *testing.T) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "worker.go", `package worker
func setup(){w.RegisterActivityWithOptions(handler,activity.RegisterOptions{Name:"new-business"})}`, 0)
	if err != nil {
		t.Fatal(err)
	}
	entries := temporalRegistrationEntries(fs, f, "project_management", "other/worker.go")
	if len(entries) != 1 || entries[0].Kind != "temporal-registration" || entries[0].Registration != `w.RegisterActivityWithOptions(handler, activity.RegisterOptions{Name: "new-business"})` {
		t.Fatalf("missing activity boundary: %+v", entries)
	}
}

// TestReviewedProjectHandlersCannotEscapeSourceScope 验证新只读端点的批准仅适用于审核来源。
// 参数：t 为分类测试上下文。
// 返回值：无；未知来源继承读取权限时失败。
func TestReviewedProjectHandlersCannotEscapeSourceScope(t *testing.T) {
	for _, source := range []string{"project_management/internal/httpapi/api.go", "other/routes.go"} {
		e := Entry{Source: source, Operation: "UNCLASSIFIED"}
		semanticClassify(&e, "h.settlementProjectReference")
		want := "UNCLASSIFIED"
		if source == "project_management/internal/httpapi/api.go" {
			want = "READ_HISTORY"
		}
		if e.Operation != want {
			t.Fatalf("%s got %s", source, e.Operation)
		}
	}
}

// TestSplitDispatchRejectsChangedHandler 验证同路径替换成未审核处理器时不能继承既有方法策略。
// 参数：t 为隔离AST上下文。
// 返回值：无；未知处理器被分类为已审核操作时失败。
func TestSplitDispatchRejectsChangedHandler(t *testing.T) {
	fs := token.NewFileSet()
	source := "Settlement/internal/httpapi/role_finance.go"
	f, err := parser.ParseFile(fs, source, `package httpapi
func(a *API)serveRoleFinance(w W,r *R)bool{switch{case path=="settings":a.newBusiness(w,r)};return true}`, 0)
	if err != nil {
		t.Fatal(err)
	}
	entries := settlementDispatchEntries(fs, f, source, 0)
	if len(entries) != 1 || entries[0].Operation != "UNCLASSIFIED" {
		t.Fatalf("unknown handler inherited policy: %+v", entries)
	}
}

// TestProjectBackgroundKindsRejectNewLoopMember 验证新增循环活动不能继承既有任务许可策略。
// 参数：t 为隔离任务定义与调度文件上下文。
// 返回值：无；新增任务被归入已审核策略或移除未被发现时失败。
func TestProjectBackgroundKindsRejectNewLoopMember(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "project_management/internal/background")
	os.MkdirAll(dir, 0700)
	types := `package background
const(TaskSLA="sla";TaskEffects="effects";TaskNotifications="notifications";TaskNew="new";SLAActivityName="project.background.sla.v1";EffectsActivityName="project.background.effects.v1";NotificationsActivityName="project.background.notifications.v1";NewName="unknown")
func activityName(kind K)(string,error){switch kind{case TaskSLA:return SLAActivityName,nil;case TaskEffects:return EffectsActivityName,nil;case TaskNotifications:return NotificationsActivityName,nil;case TaskNew:return NewName,nil};return "",nil}`
	os.WriteFile(filepath.Join(dir, "types.go"), []byte(types), 0600)
	for _, members := range []string{"TaskSLA,TaskEffects,TaskNotifications", "TaskSLA,TaskEffects,TaskNotifications,TaskNew", "TaskSLA,TaskNotifications"} {
		os.WriteFile(filepath.Join(dir, "schedules.go"), []byte("package background\nvar taskKinds=[]K{"+members+"}"), 0600)
		r := Registry{SourceDigests: map[string]string{}}
		err := reviewProjectBackgroundKinds(root, &r, Entry{ID: "mixed", Application: "project_management"})
		if !strings.Contains(members, "TaskEffects") {
			if err == nil {
				t.Fatal("removed reviewed activity not rejected")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(members, "TaskNew") {
			if Check(r, r) == nil {
				t.Fatal("unknown activity inherited policy")
			}
		} else if err := Check(r, r); err != nil {
			t.Fatal(err)
		}
	}
}
