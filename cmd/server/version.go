package main

import "fmt"

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func init() {
	buildVersion = fillIfEmpty(buildVersion)
	buildDate = fillIfEmpty(buildDate)
	buildCommit = fillIfEmpty(buildCommit)
}

func showVersion() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n\n", buildVersion, buildDate, buildCommit)
}

func fillIfEmpty(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}
