// Package noexit implements a custom static analysis tool that reports
// the usage of os.Exit inside the main.main function.
//
// Purpose:
// Disallow abrupt application termination in Go programs by using os.Exit directly
// in the main function. This enforces the use of graceful shutdown logic, logging,
// or proper error handling instead of forced termination.
//
// Example (incorrect):
//
//	func main() {
//	    os.Exit(1) - not allowed
//	}
//
// Integration:
// This analyzer is designed to be used within a multichecker setup (e.g. in cmd/staticlint)
// and supports package filtering to avoid false positives in generated test or mock code.
//
// It skips:
//   - Packages with `.test` suffix (test binary packages)
//   - Packages whose import path contains `_mock` (typically mocks)
package noexit

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer defines the staticcheck-compatible analyzer instance for noexit.
var Analyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "reports usage of os.Exit in main.main",
	Run:  run,
}

// run implements the logic for the noexit analyzer.
// It scans all files in the current package, and if the package is named "main",
// it traverses the body of main.main function to detect calls to os.Exit.
func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Skip test and mock packages
	if strings.HasSuffix(pass.Pkg.Path(), ".test") || strings.Contains(pass.Pkg.Path(), "_mock") {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != "main" {
				continue
			}

			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				if pkgIdent.Name == "os" && sel.Sel.Name == "Exit" {
					pos := pass.Fset.Position(call.Pos())

					fmt.Printf("Detected os.Exit in main.main\n")
					fmt.Printf("Package: %s\n", pass.Pkg.Path())
					fmt.Printf("File:    %s\n", pos.Filename)
					fmt.Printf("Line:    %d, Column: %d\n", pos.Line, pos.Column)

					pass.Reportf(call.Pos(), "os.Exit usage is forbidden in main.main")
				}
				return true
			})
		}
	}

	fmt.Println("noexit analyzed:", pass.Pkg.Path())
	return nil, nil
}
