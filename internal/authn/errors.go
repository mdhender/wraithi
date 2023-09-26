// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package authn

// Errors used by the package.
const (
	ErrDecodingResponse = constError("decoding response")
	ErrExchangeCode     = constError("exchange-code failed")
	ErrFetchUser        = constError("fetch-user failed")
	ErrInvalidNonce     = constError("invalid nonce")
	ErrMethodNotAllowed = constError("method not allowd")
	ErrNotImplemented   = constError("not implemented")
	ErrReadingResponse  = constError("reading response")
	ErrUnknownProvider  = constError("unknown provider")
)

// declarations to support constant errors
type constError string

func (ce constError) Error() string {
	return string(ce)
}
