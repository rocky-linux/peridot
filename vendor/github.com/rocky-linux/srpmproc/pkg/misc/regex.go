package misc

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rocky-linux/srpmproc/pkg/data"
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

// Given a git reference in tagless mode (like "refs/heads/c9s", or "refs/heads/stream-httpd-2.4-rhel-9.1.0"), determine
// if we are ok with importing that reference.  We are looking for the traditional <prefix><version><suffix> pattern, like "c9s", and also the
// modular "stream-<NAME>-<VERSION>-rhel-<VERSION> branch pattern as well
func TaglessRefOk(tag string, pd *data.ProcessData) bool {
	// First case is very easy: if we are "refs/heads/<prefix><version><suffix>" , then this is def. a branch we should import
	if strings.HasPrefix(tag, fmt.Sprintf("refs/heads/%s%d%s", pd.ImportBranchPrefix, pd.Version, pd.BranchSuffix)) {
		return true
	}

	// Less easy: if a modular branch is present (starts w/ "stream-"), we need to check if it's part of our major version, and return true if it is
	// (major version means we look for the text "rhel-X." in the branch name, like "rhel-9.1.0")
	if strings.HasPrefix(tag, "refs/heads/stream-") && strings.Contains(tag, fmt.Sprintf("rhel-%d.", pd.Version)) {
		return true
	}

	return false
}
