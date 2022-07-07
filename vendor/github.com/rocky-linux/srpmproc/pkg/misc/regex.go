package misc

import (
	"fmt"
	"github.com/rocky-linux/srpmproc/pkg/data"
	"path/filepath"
	"regexp"
)

func GetTagImportRegex(pd *data.ProcessData) *regexp.Regexp {
	branchRegex := regexp.QuoteMeta(fmt.Sprintf("%s%d%s", pd.ImportBranchPrefix, pd.Version, pd.BranchSuffix))
	if !pd.StrictBranchMode {
		branchRegex += "(?:.+|)"
	} else {
		branchRegex += "(?:-stream-.+|)"
	}

	initialVerRegex := regexp.QuoteMeta(filepath.Base(pd.RpmLocation)) + "-"
	if pd.PackageVersion != "" {
		initialVerRegex += regexp.QuoteMeta(pd.PackageVersion) + "-"
	} else {
		initialVerRegex += ".+-"
	}
	if pd.PackageRelease != "" {
		initialVerRegex += regexp.QuoteMeta(pd.PackageRelease)
	} else {
		initialVerRegex += ".+"
	}

	regex := fmt.Sprintf("(?i)refs/tags/(imports/(%s)/(%s))", branchRegex, initialVerRegex)

	return regexp.MustCompile(regex)
}
