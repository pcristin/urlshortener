package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// NewNoExitAnalyzer creates a new analyzer that prohibits using os.Exit in the main function of main package.
func NewNoExitAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "noexit",
		Doc:  "Checks that main function in main package doesn't directly call os.Exit",
		Run:  run,
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Check only in main package
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Look for main functions
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		funcDecl := node.(*ast.FuncDecl)

		// Only interested in the main function
		if funcDecl.Name.Name != "main" {
			return
		}

		// Visit all function calls in the main function
		ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
			// Look for a function call
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Check if it's a selector expression (e.g., os.Exit)
			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			// Get the identifier (X part of X.Y)
			ident, ok := selExpr.X.(*ast.Ident)
			if !ok {
				return true
			}

			// Check if it's 'os.Exit'
			if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
				pass.Reportf(callExpr.Pos(), "direct call to os.Exit is not allowed in main function")
			}

			return true
		})
	})

	return nil, nil
}
