package main

import (
	"github.com/vilasle/metrics/internal/staticlint"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {

	checks := rules()
	checks = append(checks, staticlint.ErrDirectlyExitFromMain)

	multichecker.Main(checks...)
}
