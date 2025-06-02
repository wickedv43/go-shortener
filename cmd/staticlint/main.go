// Package main provides a static analysis tool for the go-shortener project.
//
// This tool is built using golang.org/x/tools/go/analysis/multichecker and includes:
//
//   - Standard analyzers from the golang.org/x/tools/go/analysis/passes package
//   - All SA-class analyzers from the staticcheck.io toolset
//   - A custom analyzer (noexit) that forbids the use of os.Exit in main.main
//
// Usage:
//
//	go run ./cmd/staticlint ./...
//
// Each analyzer logs its activity to stdout, showing:
//   - The name of the analyzer
//   - The package being analyzed or skipped
//
// Included analyzers:
//
//   - SA**** — All staticcheck safety and correctness rules (e.g., SA1000, SA4012, etc.)
//   - fieldalignment — Checks struct field ordering for memory alignment optimization
//   - nilness, shadow, unreachable, structtag — Go compiler-level analysis passes
//   - noexit — Custom analyzer that forbids os.Exit usage in main.main for safe shutdown handling
package main

import (
	"strings"

	"github.com/wickedv43/go-shortener/cmd/staticlint/noexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"

	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unreachable"

	"honnef.co/go/tools/staticcheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	for _, a := range staticcheck.Analyzers {
		if len(a.Analyzer.Name) >= 2 && a.Analyzer.Name[:2] == "SA" {
			analyzers = append(analyzers, a.Analyzer)
		}
	}

	analyzers = append(analyzers,
		nilness.Analyzer,
		shadow.Analyzer,
		unreachable.Analyzer,
		structtag.Analyzer,
		filteredFieldAlignment(),
		noexit.Analyzer,
	)

	multichecker.Main(analyzers...)
}

// filteredFieldAlignment wraps the fieldalignment analyzer to skip test and mock packages.
func filteredFieldAlignment() *analysis.Analyzer {
	a := *fieldalignment.Analyzer
	origRun := a.Run
	a.Run = func(pass *analysis.Pass) (interface{}, error) {
		pkg := pass.Pkg.Path()
		if strings.HasSuffix(pkg, ".test") || strings.Contains(pkg, "/internal/mocks") {
			return nil, nil
		}
		return origRun(pass)
	}
	return &a
}
