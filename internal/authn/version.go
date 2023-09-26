// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package authn

import "github.com/mdhender/wraithi/internal/semver"

var (
	version string = semver.Version{Major: 0, Minor: 1, Patch: 0}.String()
)

func Version() string {
	return version
}
