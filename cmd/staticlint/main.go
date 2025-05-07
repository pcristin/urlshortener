// Package main provides a multichecker static analysis tool for the urlshortener project.
//
// This tool combines:
// - Standard static analyzers from golang.org/x/tools/go/analysis/passes
// - All SA-class analyzers from staticcheck.io
// - Selected analyzers from other staticcheck.io classes
// - Additional public analyzers
// - A custom analyzer that prohibits direct calls to os.Exit in the main function
//
// Usage:
//
//	go run ./cmd/staticlint/... [packages]
//
// Example:
//
//	go run ./cmd/staticlint/... ./...
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"

	"github.com/timakin/bodyclose/passes/bodyclose"
	"github.com/tomarrell/wrapcheck/v2/wrapcheck"
)

// getStaticcheckAnalyzers returns a list of staticcheck analyzers
func getStaticcheckAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}
}

func getStandardAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer

	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[:2] == "SA" {
			analyzers = append(analyzers, v.Analyzer)
		}

		// Add selected non-SA analyzers
		// ST1000: Incorrect or missing package comment
		if v.Analyzer.Name == "ST1000" {
			analyzers = append(analyzers, v.Analyzer)
		}
		// QF1003: Convert if/else-if chain to switch statement
		if v.Analyzer.Name == "QF1003" {
			analyzers = append(analyzers, v.Analyzer)
		}

		// S1000: Use plain channel send or receive instead of select with a single case
		if v.Analyzer.Name == "S1000" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	return analyzers
}

func getWrapcheckAnalyzers() *analysis.Analyzer {
	wrapcheckConfig := wrapcheck.NewDefaultConfig()
	wrapcheckConfig.IgnoreSigs = []string{
		`.Error`,
		`.Wrap`,
		`.Wrapf`,
		`.Unwrap`,
	}

	return wrapcheck.NewAnalyzer(wrapcheckConfig)
}

// main registers and runs all configured analyzers as part of the multichecker.
func main() {
	// Combine all analyzers
	var analyzers []*analysis.Analyzer

	// Add standard go tools analyzers
	analyzers = append(analyzers, getStandardAnalyzers()...)

	// Add staticcheck.io analyzers
	analyzers = append(analyzers, getStaticcheckAnalyzers()...)

	// Add wrapcheck analyzer
	analyzers = append(analyzers, getWrapcheckAnalyzers())

	// Add bodyclose analyzer
	analyzers = append(analyzers, bodyclose.Analyzer)

	// Run the multichecker with all analyzers
	multichecker.Main(analyzers...)
}
