// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

// Package semver implements semantic versioning.
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is the major/minor/patch of the application.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Build      string
}

// String implements the Stringer interface.
// The version string is formatted per https://semver.org/ rules
func (v Version) String() string {
	hasPreRelease, hasBuild := v.PreRelease != "", v.Build != ""
	if hasPreRelease && hasBuild {
		return fmt.Sprintf("%d.%d.%d-%s+%s", v.Major, v.Minor, v.Patch, v.PreRelease, v.Build)
	} else if hasPreRelease && !hasBuild {
		return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.PreRelease)
	} else if !hasPreRelease && hasBuild {
		return fmt.Sprintf("%d.%d.%d+%s", v.Major, v.Minor, v.Patch, v.Build)
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Less compares the versions per https://semver.org/#spec-item-11.
// Example: 1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta < 1.0.0-beta.2 < 1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0.
func (v Version) Less(v2 Version) bool {
	// compare major first
	if v.Major < v2.Major {
		return true
	} else if v.Major > v2.Major {
		return false
	}
	// major is equal, so compare minor
	if v.Minor < v2.Minor {
		return true
	} else if v.Minor > v2.Minor {
		return false
	}
	// major and minor are equal, so compare patch
	if v.Patch < v2.Patch {
		return true
	} else if v.Patch > v2.Patch {
		return false
	}

	// major, minor, patch are equal, so compare pre-release.
	fields1 := strings.Split(v.PreRelease, ".")
	fields2 := strings.Split(v2.PreRelease, ".")
	for i := 0; i < len(fields1) && i < len(fields2); i++ {
		n1, err1 := strconv.Atoi(fields1[i])
		n2, err2 := strconv.Atoi(fields2[i])
		if err1 == nil && err2 == nil { // both fields are int
			if n1 < n2 {
				return true
			} else if n1 > n2 {
				return false
			}
			// n1 and n2 are equal, so compare the next field
		} else if err1 == nil { // only field1 is an int
			return true
		} else if err2 == nil { // only field2 is an int
			return false
		} else if fields1[i] < fields2[i] { // compare as text
			return true
		} else if fields1[i] > fields2[i] { // compare as text
			return false
		}
	}
	return len(fields1) >= len(fields2)
}
