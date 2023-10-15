// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package sessions

import (
	"log"
	"net/http"
	"strings"
)

// FromRequest extracts a token from a request.
// It tries to find a bearer token first.
// If it can't, it searches for a cookie.
func FromRequest(r *http.Request, cookie string) string {
	token := fromBearerToken(r)
	if token == "" {
		token = fromCookie(r, cookie)
	}
	return token
}

// fromBearerToken extracts and returns a bearer token from the request.
// Returns an empty string if there is no bearer token or the token is invalid.
func fromBearerToken(r *http.Request) string {
	// first try a bearer token
	// log.Printf("[session] bearer: entered\n")
	headerAuthText := r.Header.Get("Authorization")
	if headerAuthText == "" {
		return ""
	}
	// log.Printf("[session] bearer: found authorization header\n")
	authTokens := strings.SplitN(headerAuthText, " ", 2)
	if len(authTokens) != 2 {
		return ""
	}
	// log.Printf("[session] bearer: found authorization token\n")
	authType, authToken := authTokens[0], strings.TrimSpace(authTokens[1])
	if authType != "Bearer" {
		return ""
	}
	// log.Printf("[session] bearer: found bearer token\n")
	return authToken
}

// fromCookie extracts and returns a token from a cookie in the request.
// Returns an empty string if there is no cookie or the token is invalid.
func fromCookie(r *http.Request, cookie string) string {
	// log.Printf("[session] cookie: entered\n")
	c, err := r.Cookie(cookie)
	if err != nil {
		log.Printf("[session] cookie: %+v\n", err)
		return ""
	}
	// log.Printf("[session] cookie: token\n")
	return c.Value
}
