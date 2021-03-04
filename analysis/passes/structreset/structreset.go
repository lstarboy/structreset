package structreset

import (
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"strings"
)

const Doc = `check for the completeness of Reset function 

The struct reset checker ensures the fields of struct are always have the reset code frame.`

var Analyzer = &analysis.Analyzer{
	Name:     "structreset",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var refCountTypeName = map[string]uint8 {"refcount": 1}

func scanAsAssignStmt(assStmt *ast.AssignStmt, m *map[string]bool, isObjectSelector func(s *ast.SelectorExpr) bool)  {
	selectorExpr, isSelector := assStmt.Lhs[0].(*ast.SelectorExpr)
	switch x := assStmt.Lhs[0].(type) {
	case *ast.SelectorExpr:
		if ok, name := isObjectSelectorRecursive(x, isObjectSelector); ok && name != "" {
			(*m)[name] = true
		}
	case *ast.IndexExpr:
		s, ok := x.X.(*ast.SelectorExpr)
		if ok {
			if ok, name := isObjectSelectorRecursive(s, isObjectSelector); ok && name != "" {
				(*m)[name] = true
			}
		}
	}

	if isSelector && isObjectSelector(selectorExpr) && selectorExpr.Sel.Name != "" {
		(*m)[selectorExpr.Sel.Name] = true
	}
}
func scanAsCallStmt(exprStmt *ast.ExprStmt, m *map[string]bool, getContextFuncBody getContextFunc, isObjectSelector func(s *ast.SelectorExpr) bool)  {
	//i := callExpr.Fun.(*ast.Ident)

	switch x := exprStmt.X.(type) {
	case *ast.CallExpr:
		scanCallExpr(x, m, getContextFuncBody, isObjectSelector)
	}

}

func isObjectSelectorRecursive(s *ast.SelectorExpr, f func(s *ast.SelectorExpr) bool) (bool, string) {
	switch x := s.X.(type) {
	case *ast.SelectorExpr:
		return isObjectSelectorRecursive(x, f)
	case *ast.Ident:
		return f(s), s.Sel.Name
	default:
		return false, ""
	}
}

func scanCallExpr(c *ast.CallExpr, m *map[string]bool, getContextFuncBody getContextFunc, isObjectSelector func(s *ast.SelectorExpr) bool) {
	switch y := c.Fun.(type) {
	case *ast.Ident:
		// 暂时只支持对函数调用的「入参」做识别。
		// 如，delete(p.d, key) 函数调用，遍历参数列表
		// TODO 如果是传递对象，再由函数内部对属性做处理，暂时未能识别
		//      如：resetBCD(p) 函数内部再处理 p.b, p.c, p.d 的情况
		for _, arg := range c.Args {
			switch x := arg.(type) {
			case *ast.SelectorExpr:
				if ok, name := isObjectSelectorRecursive(x, isObjectSelector); ok && name != "" {
					(*m)[name] = true
				}
			case *ast.CallExpr:
				scanCallExpr(x, m, getContextFuncBody, isObjectSelector)
			}
		}
	case *ast.SelectorExpr:
		if isObjectSelector(y) {
			// p.clearXYZ() 方法调用，递归函数体处理
			decl := getContextFuncBody(y.Sel.Name, true)
			if decl != nil {
				scanFuncResetFields(decl.Body, m, getContextFuncBody, isObjectSelector)
			}
		} else {
			// p.f.Reset() 属性方法调用
			if ok, name := isObjectSelectorRecursive(y, isObjectSelector); ok && name != "" {
				(*m)[name] = true
			}
		}
	}
}

type getContextFunc func(functionName string, method bool) *ast.FuncDecl
func scanFuncResetFields(body *ast.BlockStmt, m *map[string]bool, getContextFuncBody getContextFunc, isObjectSelector func(s *ast.SelectorExpr) bool) {
	for _, stmt := range body.List {
		switch x := stmt.(type) {
		case *ast.AssignStmt:
			// 赋值清空 p.a = 0, p.a = "", p.a = nil, p.a=p.a[:0], p.a[i]=false
			scanAsAssignStmt(x, m, isObjectSelector)
		case *ast.ExprStmt:
			// 属性方法调用 p.b.Reset() p.b.Release() p.b.xxx()
			scanAsCallStmt(x, m, getContextFuncBody, isObjectSelector)

			// 方法调用，递归判断
			// 1. p.clearXYZ()
			// 2. delete(p.m, xx)
			// 3. clearXYZ(p)

		case *ast.IfStmt:
			if x.Body != nil {
				scanFuncResetFields(x.Body, m, getContextFuncBody, isObjectSelector)
			}
			if x.Else != nil {
				scanFuncResetFields(x.Else.(*ast.BlockStmt), m, getContextFuncBody, isObjectSelector)
			}
		case *ast.ForStmt:
			if x.Body != nil {
				scanFuncResetFields(x.Body, m, getContextFuncBody, isObjectSelector)
			}
		case *ast.RangeStmt:
			if x.Body != nil {
				scanFuncResetFields(x.Body, m, getContextFuncBody, isObjectSelector)
			}
		case *ast.SwitchStmt:
			if x.Body != nil {
				scanFuncResetFields(x.Body, m, getContextFuncBody, isObjectSelector)
			}
		default:
			// TODO 语句未识别
		}
	}
}

func compareFuncStmtAndStruct(pass *analysis.Pass, structName string, f []*ast.Field, m map[string]bool) {
	for _, field := range f {
		if field.Names == nil {
			continue
		}

		fieldName := field.Names[0].Name
		if !m[fieldName] {
			pass.Report(analysis.Diagnostic{
				Pos:     field.Pos(),
				Message: fmt.Sprintf("struct[%s].%s 缺少Reset语句", structName, field.Names[0].Name),
			})
		}
	}
}

func isRefCountType(t string) bool {
	exists := refCountTypeName[strings.ToLower(t)]
	return exists == 1
}

func isContainRefCount(fields []*ast.Field) (bool, token.Pos) {
	for _, field := range fields {
		switch x := field.Type.(type) {
		case *ast.Ident:
			if isRefCountType(x.Name) {
				return true, x.NamePos
			}
		case *ast.SelectorExpr:
			if isRefCountType(x.Sel.Name) {
				return true, x.Sel.NamePos
			}
		}
	}

	return false, 0
}

func getStructNameFromRecv(t ast.Expr) string {
	switch x := t.(type) {
	case *ast.StarExpr:
		return x.X.(*ast.Ident).Name
	case *ast.Ident:
		return x.Name
	}
	return ""
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.StructType)(nil),
		(*ast.FuncDecl)(nil),
	}
	structFields := make(map[string][]*ast.Field)
	structMethods := make(map[string]map[string]*ast.FuncDecl)
	normalFunctions := make(map[string]*ast.FuncDecl)
	resetFunctionStruct := make(map[string]*ast.FuncDecl)

	for ident, _ := range pass.TypesInfo.Defs {
		if ident.Obj != nil {
			if ident.Obj.Decl != nil {
				ts, ok := ident.Obj.Decl.(*ast.TypeSpec)
				if ok && ts.Type != nil {
					st, ok1 := ts.Type.(*ast.StructType)
					if ok1 {
						structFields[ident.Name] = st.Fields.List
					}
				}
			}
		}
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		function, ok2 := n.(*ast.FuncDecl)
		if ok2 {
			funcName := function.Name.Name
			if function.Recv == nil || len(function.Recv.List) < 1 {
				normalFunctions[funcName] = function
				return
			}

			//structName := function.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
			structName := getStructNameFromRecv(function.Recv.List[0].Type)
			if structName != "" {
				if structMethods[structName] == nil {
					structMethods[structName] = make(map[string]*ast.FuncDecl)
				}
				structMethods[structName][funcName] = function
			}

			if structFields[structName] != nil && funcName == "Reset" {
				resetFunctionStruct[structName] = function
				return
			}
		}
	})

	readyToCheckStruct := make(map[string]bool)
	// 检查包含 refCount 结构体是否包含Reset方法
	for s, fields := range structFields {
		if ok, pos := isContainRefCount(fields); ok  {
			readyToCheckStruct[s] = true
			if resetFunctionStruct[s] == nil {
				pass.Report(analysis.Diagnostic{
					Pos:     pos,
					Message: fmt.Sprintf("struct[%s] 未找到响相应的Reset方法", s),
				})
			}
		}
	}

	var f getContextFunc

	for structName, decl := range resetFunctionStruct {
		if !readyToCheckStruct[structName] {
			continue
		}
		f = func(funcName string, method bool) *ast.FuncDecl {
			if method {
				m := structMethods[structName]
				if m != nil {
					return m[funcName]
				}
			} else {
				return normalFunctions[funcName]
			}
			return nil
		}

		isObjectSelector := func(s *ast.SelectorExpr) bool {
			i, ok := s.X.(*ast.Ident)
			if !ok {
				return false
			}
			return i.Name == decl.Recv.List[0].Names[0].Name && s.Sel.Name != ""
		}

		m := make(map[string]bool)
		scanFuncResetFields(decl.Body, &m, f, isObjectSelector)

		compareFuncStmtAndStruct(pass, structName, structFields[structName], m)
	}

	return nil, nil
}
