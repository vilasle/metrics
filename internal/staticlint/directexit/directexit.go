package directexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// DirectExitAnalyzer is a staticlint analyzer that checks for directly call os.Exit in main package
var DirectExitAnalyzer = &analysis.Analyzer{
	Name: "directlyExitFromMain",
	Doc:  "check directly call os.Exit in main package",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		_ = file
		ast.Inspect(file, func(n ast.Node) bool {

			decl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if decl.Name.Name != "main" {
				return true
			}

			ast.Inspect(decl, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				if selector.X.(*ast.Ident).Name == "os" && selector.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "directly call os.Exit in main package")
				}
				return true
			})

			return true
		})
	}
	return nil, nil
}
