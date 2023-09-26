// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package authn

import "net/http"

// Provider interface for OAuth2 clients.
type Provider interface {
	Name() string              // official name of the provider
	Code() string              // lower case name of the provider
	Debug(bool)                // enable/disable debugging for the provider
	LoginURL() (string, error) // get login url (with nonce, if needed)
	ProcessCallback(r *http.Request) (Authentication, error)
	Version() string
}
