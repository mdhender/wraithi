// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

// Package homedir implements a wrapper around Mitchell Hashimoto's
// homedir package (github.com/mitchellh/go-homedir).
package homedir

import (
	hd "github.com/mitchellh/go-homedir"
	"path/filepath"
)

// Dir returns the home directory for the current user or an error.
func Dir() (string, error) {
	home, err := hd.Dir()
	if err != nil {
		return "", err
	} else if home, err = filepath.Abs(home); err != nil {
		return "", err
	}
	return home, err
}
