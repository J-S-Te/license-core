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

// reviewProjectBackgroundKinds 把循环注册的实际活动名称与有限业务策略绑定。
// 参数：root 为工作区；r 为清单；parent 为真实编排或执行边界。
// 返回值：类型/注册不完整、重复或已审核任务被移除时返回错误；新增任务保留未分类。
func reviewProjectBackgroundKinds(root string, r *Registry, parent Entry) error {
	files := map[string]*ast.File{}
	for _, name := range []string{"types.go", "schedules.go"} {
		source := "project_management/internal/background/" + name
		raw, err := os.ReadFile(filepath.Join(root, source))
		if err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		r.SourceDigests[source] = hex.EncodeToString(sum[:])
		file, err := parser.ParseFile(token.NewFileSet(), source, raw, 0)
		if err != nil {
			return err
		}
		files[name] = file
	}
	constants, names := projectActivityNames(files["types.go"])
	members := []string{}
	ast.Inspect(files["schedules.go"], func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok || len(spec.Names) != 1 || spec.Names[0].Name != "taskKinds" || len(spec.Values) != 1 {
			return true
		}
		list, ok := spec.Values[0].(*ast.CompositeLit)
		if !ok {
			return false
		}
		for _, element := range list.Elts {
			if id, ok := element.(*ast.Ident); ok {
				members = append(members, id.Name)
			} else {
				members = append(members, "<dynamic>")
			}
		}
		return false
	})
	return appendProjectActivityPolicies(r, parent, members, constants, names)
}

// projectActivityNames 解析活动名称函数实际引用的常量，不按名称前缀推断授权。
// 参数：file 为已审核的后台类型源码树。
// 返回值：字符串常量及任务标识到实际名称常量的映射。
func projectActivityNames(file *ast.File) (map[string]string, map[string]string) {
	constants, names := map[string]string{}, map[string]string{}
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok || len(spec.Names) != 1 || len(spec.Values) != 1 {
			return true
		}
		if literal, ok := spec.Values[0].(*ast.BasicLit); ok && literal.Kind == token.STRING {
			value, err := strconv.Unquote(literal.Value)
			if err == nil {
				constants[spec.Names[0].Name] = value
			}
		}
		return true
	})
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "activityName" {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			clause, ok := node.(*ast.CaseClause)
			if !ok || len(clause.Body) != 1 {
				return true
			}
			ret, ok := clause.Body[0].(*ast.ReturnStmt)
			if !ok || len(ret.Results) < 1 {
				return false
			}
			name, ok := ret.Results[0].(*ast.Ident)
			if !ok {
				return false
			}
			for _, condition := range clause.List {
				if id, ok := condition.(*ast.Ident); ok {
					names[id.Name] = name.Name
				}
			}
			return false
		})
	}
	return constants, names
}

// appendProjectActivityPolicies 核对三个已审核活动并显式保留未知循环成员。
// 参数：r 为清单；parent 为混合父项；members 为实际循环；constants、names 为源码解析结果。
// 返回值：已审核成员缺失或重复时返回错误。
func appendProjectActivityPolicies(r *Registry, parent Entry, members []string, constants, names map[string]string) error {
	expected := map[string]struct{ name, op string }{"TaskSLA": {"project.background.sla.v1", "MUTATE_BUSINESS"}, "TaskEffects": {"project.background.effects.v1", "MUTATE_BUSINESS"}, "TaskNotifications": {"project.background.notifications.v1", "ESSENTIAL_SERVICE"}}
	seen := map[string]bool{}
	for _, member := range members {
		if seen[member] {
			return fmt.Errorf("duplicate background activity %s", member)
		}
		seen[member] = true
		actual := constants[names[member]]
		policy, known := expected[member]
		op := "UNCLASSIFIED"
		reason := "New or dynamic activity requires source review and expanded runtime comparison"
		if known && actual == policy.name {
			op = policy.op
			reason = "Reviewed exact loop activity: SLA/effects generate business; notifications deliver persisted immutable payload. Runtime currently applies stricter business checks to notifications"
		} else {
			delete(seen, member)
		}
		r.Entries = append(r.Entries, Entry{ID: parent.ID + ":policy:activity:" + member, Application: "project_management", Source: "project_management/internal/background/schedules.go", Kind: "task-policy", Registration: actual + " (" + member + ")", Operation: op, Reason: reason})
	}
	for member := range expected {
		if !seen[member] {
			return fmt.Errorf("reviewed background activity removed or renamed: %s", member)
		}
	}
	return nil
}
