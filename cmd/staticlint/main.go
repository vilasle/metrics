package main

import (
	"github.com/vilasle/metrics/internal/staticlint/directexit"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {

	checks := rules()
	checks = append(checks, directexit.DirectExitAnalyzer)

	multichecker.Main(checks...)
}
