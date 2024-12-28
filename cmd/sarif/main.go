package main

import (
	"fmt"
	"github.com/b4bay/aspm/internal/server/sarif"
)

func LocationHash(ploc sarif.PhysicalLocation) string {
	hash := ""
	if ploc.ArtifactLocation.Uri != "" {
		hash = ploc.ArtifactLocation.Uri
		if ploc.Region.StartLine != 0 {
			hash += fmt.Sprintf("(%d", ploc.Region.StartLine)
			if ploc.Region.StartColumn != 0 {
				hash += fmt.Sprintf(":%d", ploc.Region.StartColumn)
			}
			if ploc.Region.EndLine != 0 {
				hash += fmt.Sprintf("-%d", ploc.Region.EndLine)
			}
			if ploc.Region.EndColumn != 0 {
				hash += fmt.Sprintf(":%d", ploc.Region.EndColumn)
			}
			hash += ")"
		}
	}

	return hash
}

func main() {
	report, _ := sarif.FromBase64(sarif.MockGosecReport)
	for _, run := range report.Runs {
		fmt.Printf("\n== %s ==\n\n", run.Tool.Driver.Name)
		for _, result := range run.Results {
			hash := LocationHash(result.Locations[0].PhysicalLocation)
			fmt.Printf("%s %s %s %s\n", result.Level, result.RuleId, hash, result.Message.Text)
		}
	}

	fmt.Println()
	report, _ = sarif.FromBase64(sarif.MockGovulncheckReport)
	for _, run := range report.Runs {
		fmt.Printf("== %s ==\n\n", run.Tool.Driver.Name)
		for _, result := range run.Results {
			hash := LocationHash(result.Locations[0].PhysicalLocation)
			fmt.Printf("%s %s %s %s\n", result.Level, result.RuleId, hash, result.Message.Text)
		}
	}
	//	fmt.Printf("%v\n", report)
}
