package coverage

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// appendSplitActivitySources 补齐已注册活动结构体在同包内的拆分实现。
// 参数：root 为工作区；files 为既有有限任务来源，按应用补充实际活动文件。
// 返回值：源码语法或路径枚举失败时返回错误，不静默遗漏活动。
func appendSplitActivitySources(root string, files map[string][]string) error {
	packages := map[string]string{
		"contract_management":      "contract_management/internal/workflows",
		"customer_and_opportunity": "customer_and_opportunity/internal/presaleworkflow",
		"project_management":       "project_management/internal/background",
	}
	for app, directory := range packages {
		paths, err := filepath.Glob(filepath.Join(root, directory, "*.go"))
		if err != nil {
			return err
		}
		for _, path := range paths {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			fs := token.NewFileSet()
			file, err := parser.ParseFile(fs, path, nil, 0)
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			present := false
			for _, source := range files[app] {
				present = present || source == relative
			}
			if !present && len(activityMethodEntries(fs, file, app, relative)) > 0 {
				files[app] = append(files[app], relative)
			}
		}
	}
	return nil
}

// activityMethodEntries 只枚举 Temporal 可注册的导出活动方法，并应用有限来源策略。
// 参数：fs 和 file 为实际语法树；app 和 source 为应用身份及真实源码路径。
// 返回值：方法登记；新增或迁移至未审核来源的方法保持未分类。
func activityMethodEntries(fs *token.FileSet, file *ast.File, app, source string) []Entry {
	entries := []Entry{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !fn.Name.IsExported() || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		receiver := expression(fs, fn.Recv.List[0].Type)
		if receiver != "*Activities" && receiver != "Activities" {
			continue
		}
		name := fn.Name.Name
		op, reason := "UNCLASSIFIED", "New exported activity requires implementation and source review"
		for operation, names := range reviewedActivityMethods[source] {
			for _, reviewed := range strings.Fields(names) {
				if name == reviewed {
					op, reason = operation, "Reviewed exact activity source and method; business side effects remain separately licensed"
				}
			}
		}
		entries = append(entries, Entry{ID: app + ":" + source + ":activity:" + name,
			Application: app, Source: source, Kind: "activity-method",
			Registration: "*Activities." + name, Operation: op, Reason: reason})
	}
	return entries
}

var reviewedActivityMethods = map[string]map[string]string{
	"contract_management/internal/workflows/activities.go": {
		"MUTATE_BUSINESS": "ArchiveExpired CompleteApproval CreateNotification RecordCommand StartApproval",
	},
	"contract_management/internal/workflows/number_audit.go": {
		"ESSENTIAL_SERVICE": "AuditContractNumbers",
	},
	"contract_management/internal/workflows/import_confirmation.go": {
		"MUTATE_BUSINESS": "AdvanceImportConfirmation",
	},
	"customer_and_opportunity/internal/presaleworkflow/types.go": {
		"MUTATE_BUSINESS": "ApprovalAction PublishWorklog StartApproval",
	},
	"project_management/internal/background/activities.go": {
		"MIXED_EXECUTION": "RunOne",
	},
}
