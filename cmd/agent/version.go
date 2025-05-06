package main

import "fmt"

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func showVersion() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n\n", buildVersion, buildDate, buildCommit)
}

func init() {
	buildVersion = fillIfEmpty(buildVersion)
	buildDate = fillIfEmpty(buildDate)
	buildCommit = fillIfEmpty(buildCommit)
}

func fillIfEmpty(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}
