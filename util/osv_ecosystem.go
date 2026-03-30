package util

import (
	"regexp"
	"strings"
)

// ExtractOSVEcosystems converts a Txlog OS string like "AlmaLinux 9.4" into
// a list of valid OSV ecosystem identifiers. The repo parameter is used for
// Red Hat-family systems to derive the exact CPE channel (baseos, appstream, crb).
func ExtractOSVEcosystems(osString, repo string) []string {
	var ecosystems []string

	if osString == "" {
		return ecosystems
	}

	// Extract major version
	versionRe := regexp.MustCompile(`\b(\d+)\.\d+`)
	matches := versionRe.FindStringSubmatch(osString)
	var majorVersion string
	if len(matches) > 1 {
		majorVersion = matches[1]
	} else if regexp.MustCompile(`\b(\d+)`).MatchString(osString) {
		majorVersion = regexp.MustCompile(`\b(\d+)`).FindStringSubmatch(osString)[1]
	}

	if majorVersion == "" {
		return ecosystems
	}

	osName := strings.ToLower(osString)

	if strings.Contains(osName, "almalinux") {
		ecosystems = append(ecosystems, "AlmaLinux:"+majorVersion)
	} else if strings.Contains(osName, "rocky") {
		ecosystems = append(ecosystems, "Rocky Linux:"+majorVersion)
	} else if strings.Contains(osName, "red hat") || strings.Contains(osName, "rhel") || strings.Contains(osName, "centos") || strings.Contains(osName, "oracle") {
		channels := MapRepoToRHCPEChannels(repo)
		for _, ch := range channels {
			ecosystems = append(ecosystems, "Red Hat:enterprise_linux:"+majorVersion+"::"+ch)
		}
	}

	return ecosystems
}

// MapRepoToRHCPEChannels normalizes a repository name from transaction_items.repo
// into the Red Hat CPE channel identifier(s) used by the OSV API.
//
// Examples:
//
//	"baseos"                            → ["baseos"]
//	"rhel-9-for-x86_64-baseos-rpms"    → ["baseos"]
//	"rhel-9-for-x86_64-appstream-rpms" → ["appstream"]
//	"codeready-builder-for-..."        → ["crb"]
//	"crb"                              → ["crb"]
//	""  or unrecognized                 → ["baseos", "appstream", "crb"]
func MapRepoToRHCPEChannels(repo string) []string {
	r := strings.ToLower(repo)

	switch {
	case strings.Contains(r, "baseos"):
		return []string{"baseos"}
	case strings.Contains(r, "appstream"):
		return []string{"appstream"}
	case strings.Contains(r, "codeready") || r == "crb" || strings.Contains(r, "-crb-"):
		return []string{"crb"}
	default:
		// Fallback: query all three channels
		return []string{"baseos", "appstream", "crb"}
	}
}
