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

type mixedTaskInput struct {
	source          string
	node            ast.Node
	destinationType string
	constantNames   map[*ast.Object]string
}

// settlementDestinationInput 追踪结算 Worker 实际调用的目的地构造器。
// 参数：root 为工作区；r 为源码清单；fs 和 worker 为已解析的命令入口。
// 返回值：有限构造器扫描输入；未调用时返回 nil，依赖缺失或语法异常时返回错误。
func settlementDestinationInput(
	root string, r *Registry, fs *token.FileSet, worker *ast.File,
) (*mixedTaskInput, error) {
	imported, called := false, false
	for _, imp := range worker.Imports {
		if imp.Path.Value == `"github.com/j-s-te/settlement/internal/outbox"` &&
			(imp.Name == nil || imp.Name.Name == "outbox") {
			imported = true
		}
	}
	ast.Inspect(worker, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if selector, ok := call.Fun.(*ast.SelectorExpr); ok &&
				selector.Sel.Name == "PlanPlatformDestinations" {
				if receiver, ok := selector.X.(*ast.Ident); ok &&
					receiver.Name == "outbox" && receiver.Obj == nil {
					called = true
				}
			}
		}
		return true
	})
	if !imported || !called {
		return nil, nil
	}
	const source = "Settlement/internal/outbox/destinations.go"
	raw, err := os.ReadFile(filepath.Join(root, source))
	if err != nil {
		return nil, err
	}
	helper, err := parser.ParseFile(fs, source, raw, 0)
	if err != nil {
		return nil, err
	}
	if helper.Name.Name != "outbox" {
		return nil, fmt.Errorf("reviewed destination constructor package changed: %s", source)
	}
	sum := sha256.Sum256(raw)
	r.SourceDigests[source] = hex.EncodeToString(sum[:])
	for _, decl := range helper.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == "PlanPlatformDestinations" && fn.Recv == nil && fn.Body != nil {
			return &mixedTaskInput{source: source, node: fn.Body,
				destinationType: "Destination", constantNames: literalDestinationNames(helper)}, nil
		}
	}
	return nil, fmt.Errorf("reviewed destination constructor missing: %s", source)
}

// literalDestinationNames 仅解析目的地文件内显式声明的字符串常量。
// 参数：file 为实际构造器所在源码的语法树。
// 返回值：常量名到字面量的映射；动态表达式不进入映射，仍须阻断或单独审核。
func literalDestinationNames(file *ast.File) map[*ast.Object]string {
	names := map[*ast.Object]string{}
	for _, decl := range file.Decls {
		group, ok := decl.(*ast.GenDecl)
		if !ok || group.Tok != token.CONST {
			continue
		}
		for _, spec := range group.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok || len(values.Names) != len(values.Values) {
				continue
			}
			for i, expr := range values.Values {
				literal, ok := expr.(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					continue
				}
				if value, err := strconv.Unquote(literal.Value); err == nil {
					names[values.Names[i].Obj] = value
				}
			}
		}
	}
	return names
}
