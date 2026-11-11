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
)

type settlementReviewSource20261110 struct {
	name, digest string
	routes       []settlementReviewRoute20261110
}

type settlementReviewRoute20261110 struct {
	registration, operation, reason string
}

// 此处的有限展开只适用于精确SHA对应的四个实际分派源；任意漂移均回退为未分类。
var settlementReviewSources20261110 = []settlementReviewSource20261110{
	{"native_financial.go", "23e8b489f7410ef64edfb74845f29e5524a0e286b9201b97cb0401a024cab717", []settlementReviewRoute20261110{
		{"GET /api/v1/financial-baselines/capabilities", "READ_HISTORY", "serveNativeFinancial:19读取开关和CanCreateNativeFinancial权限能力，不创建财务依据。"},
		{"GET /api/v1/financial-customers", "READ_HISTORY", "nativeFinancialCustomers:74→service.ListNativeFinancialCustomers(native_financial_query.go:24)只查询同租户本地客户。"},
		{"POST /api/v1/financial-customers", "MUTATE_BUSINESS", "nativeFinancialCustomers:74→service.CreateNativeFinancialCustomer(native_financial.go:132)事务创建本地客户、幂等及审计。"},
		{"GET /api/v1/financial-baselines", "READ_HISTORY", "nativeFinancialBaselines:107→service.ListNativeFinancialBaselines(native_financial_query.go:89)分页读取既有本地财务依据。"},
		{"POST /api/v1/financial-baselines", "MUTATE_BUSINESS", "nativeFinancialBaselines:107→service.CreateNativeFinancialBaseline(native_financial.go:225)事务写本地依据、待确认计划、幂等和审计。"},
		{"GET /api/v1/financial-baselines/:id", "READ_HISTORY", "serveNativeFinancial:44→service.GetNativeFinancialBaseline(native_financial_query.go:68)读取同租户既有依据。"},
		{"POST /api/v1/financial-baselines/:id/void", "MUTATE_BUSINESS", "serveNativeFinancial:53→service.VoidNativeFinancialBaseline(native_financial_query.go:115)作废依据并取消待确认计划，改变业务状态。"},
	}},
	{"contract_import.go", "d17f483bbf43ecea90b25095a4a2ea2fa10c60975b19f152c2cb85bc64f406ff", []settlementReviewRoute20261110{
		{"POST /api/v1/settlement/contract-imports/:id/reconcile", "MUTATE_BUSINESS", "serveContractImports:20→service.ReconcileContractImport(import_reconciliation.go:119)创建对账记录并更新READY、信用决定及审批主体。"},
		{"POST /internal/v1/settlement/events/contract-imports", "MUTATE_BUSINESS", "serveContractImports:38→service.PrepareContractImport(contract_import.go:111)经ingestContract事务创建HOLD基线、计划与导入记录。"},
		{"POST /internal/v1/settlement/contract-imports/:id/release", "MUTATE_BUSINESS", "serveContractImports:39→service.ReleaseContractImport(import_reconciliation.go:167)放行财务基线、导入及信用投递状态。"},
	}},
	{"contract_import_query.go", "9b1c35d84f8069698c96f9722948332713563380f8fc78e1b0db46ac2e23d92e", []settlementReviewRoute20261110{
		{"GET /api/v1/settlement/contract-imports", "READ_HISTORY", "serveImportQueries:14在用户权限及全租户范围校验后分页读取已有导入状态，无恢复或放行写入。"},
		{"GET /internal/v1/settlement/contract-imports/", "READ_HISTORY", "serveImportQueries:18机器前缀含尾斜杠；id为空时走分页列表，可信机器租户内只读既有导入状态。"},
		{"GET /api/v1/settlement/contract-imports/:id", "READ_HISTORY", "serveImportQueries→importStatus:158读取既有基线、计划、责任及信用决定版本，不保存派生状态。"},
		{"GET /api/v1/settlement/contract-imports/:id/facts", "READ_HISTORY", "serveImportQueries→importFacts:108分页查询同租户已恢复事实及目标业务ID映射。"},
		{"GET /internal/v1/settlement/contract-imports/capabilities", "READ_HISTORY", "serveImportQueries:53在机器身份可信租户下只输出既有开关和协议能力，不执行PREPARE/RELEASE。"},
		{"GET /internal/v1/settlement/contract-imports/:id", "READ_HISTORY", "serveImportQueries→importStatus:158仅在可信机器租户读取既有导入、计划及责任ACK。"},
	}},
	{"contract_import_export.go", "0623ae7c9f1ac87357c884ea1b6aa94c9421378a1b5aae55af163bd795813902", []settlementReviewRoute20261110{
		{"POST /api/v1/settlement/contract-imports/:id/exports", "EXPORT_HISTORY", "serveImportExports:29→service.CreateImportExport(import_export.go:46)/buildImportExport:121只读已有恢复映射生成CSV，必要写导出任务/内容/状态，不推进财务对账或放行。"},
		{"GET /api/v1/settlement/contract-imports/:id/exports/:exportID", "READ_HISTORY", "serveImportExports:52→service.GetImportExport(import_export.go:250)读取历史导出状态及内容，过期清理仅导出技术数据。"},
		{"GET /api/v1/settlement/contract-imports/:id/exports/:exportID/download", "EXPORT_HISTORY", "serveImportExports:52/63经同租户既有导出查询后只传输已成功CSV内容，不创建财务业务。"},
	}},
}

// appendSettlementReview20261110 用途：在精确源码审核绑定下展开分文件财务分派。
// 参数：root为工作区根路径，r为待追加的源码登记清单。
// 返回值：文件缺失或解析失败返回错误；任意内容漂移留下未分类条目供Check拒绝。
func appendSettlementReview20261110(root string, r *Registry) error {
	if r.SourceDigests == nil {
		r.SourceDigests = map[string]string{}
	}
	for _, source := range settlementReviewSources20261110 {
		rel := "Settlement/internal/httpapi/" + source.name
		raw, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			return err
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, rel, raw, 0)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		digest := hex.EncodeToString(sum[:])
		r.SourceDigests[rel] = digest
		if digest != source.digest {
			appendUnreviewedSettlementDispatch20261110(r, fset, file, rel)
			continue
		}
		for i, route := range source.routes {
			r.Entries = append(r.Entries, Entry{
				ID:          fmt.Sprintf("settlement:%s:review-20261110:%03d", rel, i+1),
				Application: "settlement", Source: rel, Kind: "http-dispatch-policy",
				Registration: route.registration, Operation: route.operation,
				Reason: route.reason + " 精确来源SHA绑定有限展开；保留认证、租户、权限、数据范围及审计。",
			})
		}
	}
	return nil
}

// appendUnreviewedSettlementDispatch20261110 用途：记录漂移源的实际条件和调用证据。
// 参数：r为清单，fset/file为当前真实AST，source为精确来源路径。
// 返回值：无；即使函数或条件全部删除也保留一条未分类源码漂移证据。
func appendUnreviewedSettlementDispatch20261110(
	r *Registry, fset *token.FileSet, file *ast.File, source string,
) {
	ordinal := 0
	appendUnknown := func(registration string) {
		ordinal++
		r.Entries = append(r.Entries, Entry{
			ID:          fmt.Sprintf("settlement:%s:unreviewed-20261110:%03d", source, ordinal),
			Application: "settlement", Source: source, Kind: "http-dispatch-review",
			Registration: registration, Operation: "UNCLASSIFIED",
			Reason: "分文件分派源码与已审核SHA不同，条件、方法或处理器不得继承旧策略；须重新语义审核。",
		})
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.IfStmt:
				appendUnknown(fn.Name.Name + "|if " + expression(fset, n))
			case *ast.SwitchStmt:
				appendUnknown(fn.Name.Name + "|switch " + expression(fset, n))
			}
			return true
		})
	}
	if ordinal == 0 {
		appendUnknown("reviewed dispatch functions or conditions removed")
	}
}
