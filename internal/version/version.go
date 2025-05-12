package version

import (
	"cmp"
	"fmt"
)

func ShowVersion(ver, date, commit string) {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n\n",
		cmp.Or(ver, "N/A"),
		cmp.Or(date, "N/A"),
		cmp.Or(commit, "N/A"))
}
