// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

// Package authn implements an OAuth2 flow for authenticating users.
package authn

type Authentication struct {
	Id            string
	Name          string
	Email         string
	VerifiedEmail bool
	Avatar        string
}
