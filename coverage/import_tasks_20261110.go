package coverage

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
)

// appendImportTasks20261110 用途：绑定已审核的导入后台方法，不豁免整个进程。
// 参数：root 为工作区，r 为已包含执行源摘要的候选清单。
// 返回值：已审核方法或来源消失时返回错误。
func appendImportTasks20261110(root string, r *Registry) error {
	for _, task := range reviewedImportTasks20261110 {
		if r.SourceDigests[task.Source] == "" {
			return fmt.Errorf("reviewed import task source missing: %s", task.Source)
		}
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, filepath.Join(root, task.Source), nil, 0)
		if err != nil {
			return err
		}
		found := false
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Body == nil {
				continue
			}
			if expression(set, fn.Recv.List[0].Type)+"."+fn.Name.Name == task.Registration {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("reviewed import task removed: %s", task.Registration)
		}
		task.ID = task.Application + ":" + task.Source + ":task:" + task.Registration
		task.Kind = "task-boundary"
		r.Entries = append(r.Entries, task)
	}
	return nil
}
