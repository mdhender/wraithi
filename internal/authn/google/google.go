// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package google

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/mdhender/wraithi/internal/authn"
	"github.com/mdhender/wraithi/internal/nonces"
	"github.com/mdhender/wraithi/internal/semver"
	"golang.org/x/oauth2"
	oa "golang.org/x/oauth2/google"
	"io"
	"log"
	"net/http"
)

type Provider struct {
	debug   bool
	nf      *nonces.Factory
	oauth2  *oauth2.Config
	version string
}

// Google returns an OAuth2 client that uses Google as a provider.
// The id and secret parameters are the Client ID and Secret.
// The cb parameter is the callback URL to send to the provider.
// The nf parameter is the nonce factory used when requesting and validating tokens.
func Google(id, secret, cb string, nf *nonces.Factory) *Provider {
	return &Provider{
		nf: nf,
		oauth2: &oauth2.Config{
			RedirectURL:  cb,
			ClientID:     id,
			ClientSecret: secret,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: oa.Endpoint,
		},
		version: semver.Version{Major: 0, Minor: 1, Patch: 0}.String(),
	}
}

// Name implements the authn.Provider interface.
func (p *Provider) Name() string {
	return "Google"
}

// Code implements the authn.Provider interface.
func (p *Provider) Code() string {
	return "google"
}

// Debug implements the Provider interface.
func (p *Provider) Debug(flag bool) {
	p.debug = flag
}

// LoginURL implements the authn.Provider interface.
func (p *Provider) LoginURL() (string, error) {
	// create a nonce
	nonce, err := p.nf.Create()
	if err != nil {
		return "", err
	}

	// lookup the provider's login page and add our nonce to it.
	url := p.oauth2.AuthCodeURL(nonce)
	if p.debug {
		log.Printf("[google] login: %q\n", url)
	}

	return url, nil
}

// ProcessCallback implements the authn.Provider interface.
func (p *Provider) ProcessCallback(r *http.Request) (authn.Authentication, error) {
	if r.Method != "GET" {
		return authn.Authentication{}, authn.ErrMethodNotAllowed
	}

	// extract nonce and code from the request
	nonce, code := r.FormValue("state"), r.FormValue("code")
	if !p.nf.Lookup(nonce) {
		return authn.Authentication{}, authn.ErrInvalidNonce
	}

	// exchange the code for a token
	token, err := p.oauth2.Exchange(context.TODO(), code)
	if err != nil {
		return authn.Authentication{}, errors.Join(authn.ErrExchangeCode, err)
	}

	// use that token to request user information
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return authn.Authentication{}, errors.Join(authn.ErrFetchUser, err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return authn.Authentication{}, errors.Join(authn.ErrReadingResponse, err)
	}
	if p.debug {
		log.Printf("[google] getUserInfo: contents %s\n", string(contents))
	}

	// extract the user info from the provider
	var authorization struct {
		Id            string `json:"id"`
		Name          string `json:"name,omitempty"`
		Email         string `json:"email,omitempty"`
		VerifiedEmail bool   `json:"verified_email,omitempty"`
		Picture       string `json:"picture,omitempty"`
	}
	if err := json.Unmarshal(contents, &authorization); err != nil {
		return authn.Authentication{}, errors.Join(authn.ErrDecodingResponse, err)
	}
	if p.debug {
		log.Printf("[google] getUserInfo: user +%v\n", authorization)
	}

	return authn.Authentication{
		Id:            authorization.Id,
		Name:          authorization.Name,
		Email:         authorization.Email,
		VerifiedEmail: authorization.VerifiedEmail,
		Avatar:        authorization.Picture,
	}, nil
}

// Version implements the authz.Provider interface.
func (p *Provider) Version() string {
	return p.version
}
