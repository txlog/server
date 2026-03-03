package util

import (
	"regexp"
	"strings"
)

// ExtractOSVEcosystems converts a Txlog OS string like "AlmaLinux 9.4" into
// a list of valid OSV ecosystem identifiers.
func ExtractOSVEcosystems(osString string) []string {
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
		// Map Red Hat, CentOS, Oracle Linux strictly to equivalent AlmaLinux for OSV accuracy since
		// OSV doesn't hold 'Red Hat Enterprise Linux' or 'CentOS' natively for the V1 query ecosystem format on v8+.
	} else if strings.Contains(osName, "red hat") || strings.Contains(osName, "rhel") || strings.Contains(osName, "centos") || strings.Contains(osName, "oracle") {
		ecosystems = append(ecosystems, "AlmaLinux:"+majorVersion)
	}

	return ecosystems
}
