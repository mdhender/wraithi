// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"context"
	"fmt"
	"github.com/mdhender/wraithi/internal/sessions"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// PlainOldHandler is a simple http.Handler function.
func PlainOldHandler(w http.ResponseWriter, r *http.Request) {
	// do something
}

// Adapters are http.Handler functions that allow you to wrap code around another http.Handler.
// This code is from Mat Ryer's article.
// https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81

// Adapter is a type that wraps a http.Handler around another http.Handler.
type Adapter func(http.Handler) http.Handler

// Adapt is a function to chain together adapters.
func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	// reverse the slice of adapters
	for i, j := 0, len(adapters)-1; i < j; i, j = i+1, j-1 {
		adapters[i], adapters[j] = adapters[j], adapters[i]
	}
	// then adapt them
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}

func AdaptNotify(logger *log.Logger) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Println("before")
			defer logger.Println("after")
			h.ServeHTTP(w, r)
		})
	}
}

// Interceptors work with http.HandlerFunc.
// You chain them by calling A(B(C()))
// Something like https://twitter.com/matryer/status/1445013230858952705?lang=en

func Intercept(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func InterceptNotify(h http.Handler, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Println("before")
		defer logger.Println("after")
		h.ServeHTTP(w, r)
	}
}

func anExample() {
	var h http.Handler
	var l *log.Logger
	InterceptNotify(Intercept(h), l)
	Adapt(h, AdaptNotify(l))
}

// userContextKey is the context key type for storing User in context.Context.
type userContextKey string

func (a *App) adminOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !a.currentUser(r).IsAdmin() {
			a.notFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func (a *App) authOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !a.currentUser(r).IsAuthenticated() {
			a.notFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func (a *App) mustAuth() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !a.currentUser(r).IsAuthenticated() {
				a.notFound(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

// Wrappers work with http.HandlerFunc.
// You chain them by calling A(B(C()))
// Something like https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702#.e4k81jxd3

func withNoop(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// withUser injects a User into the request's Context.
// It searches for a session token in the Bearer Token header first,
// and then in a session cookie. If a valid token is found, the User
// returned will have the appropriated roles added. If not, then the
// "guest" User and its roles are used.
func (a *App) withUser(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var u User
		// try to fetch the user from the request
		if token := sessions.FromRequest(r, a.cookies.name); token == "" {
			u.handle = "guest"
		} else {
			u.roles = UserRoles{authenticated: true, admin: true}
		}
		log.Printf("%s %s: user %q\n", r.Method, r.URL, u.handle)
		ctx := context.WithValue(r.Context(), userContextKey("user"), u)
		// serve with the user in the context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// withUser injects a User into the request's Context.
// It searches for a session token in the Bearer Token header first,
// and then in a session cookie. If a valid token is found, the User
// returned will have the appropriated roles added. If not, then the
// "guest" User and its roles are used.
func (a *App) withUserFunc(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		// try to fetch the user from the request
		if token := sessions.FromRequest(r, a.cookies.name); token == "" {
			u.handle = "guest"
		} else {
			u.roles = UserRoles{authenticated: true, admin: true}
		}
		ctx := context.WithValue(r.Context(), userContextKey("user"), u)
		// serve with the user in the context
		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

type User struct {
	handle string
	roles  UserRoles
}
type UserRoles struct {
	authenticated bool
	admin         bool
}

func (u User) IsAdmin() bool {
	return u.roles.authenticated && u.roles.admin
}

func (u User) IsAuthenticated() bool {
	return u.roles.authenticated
}

// currentUser returns the User from the request's Context.
// Returns the "guest" User if there is no User in the Context.
func (a *App) currentUser(r *http.Request) User {
	ctx := r.Context()
	if u, ok := ctx.Value(userContextKey("user")).(User); ok {
		return u
	}
	return User{handle: "guest"}
}

// match reports whether `path` matches the given `pieces`
// and assigns pointer pieces. A piece can be a string,
// *string, or *int64. This function matches pieces greedily
// and may assign pieces even when the path does not match.
// Note: Does not normalize paths with path.Clean.
// Note: Consecutive string path components need to be matched
// with separate strings, since this always splits on /.
// from https://benhoyt.com/writings/go-routing/ as updated by Yuri Vishnevsky.
func match(path string, pieces ...interface{}) bool {
	// Remove the initial "/" prefix
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	var head string
	for i, piece := range pieces {
		// Shift the next path component into `head`
		if i := strings.IndexByte(path, '/'); i == -1 {
			head, path = path, ""
		} else {
			head, path = path[:i], path[i+1:]
		}
		// Match pieces based on their type
		switch p := piece.(type) {
		case string:
			// Match a specific string
			if p != head {
				return false
			}
		case *bool:
			// Match any bool (true or false)
			switch head {
			case "true":
				*p = true
			case "false":
				*p = false
			default:
				return false
			}
		case *int:
			// Match any integer, including negative integers
			n, err := strconv.Atoi(head)
			if err != nil {
				return false
			}
			*p = n
		case *int64:
			// Match any 64-bit integer, including negative integers
			n, err := strconv.ParseInt(head, 10, 64)
			if err != nil {
				return false
			}
			*p = n
		case *string:
			// Match any string, including the empty string
			*p = head
		default:
			panic(fmt.Sprintf("unsupported type %T", piece))
		}
		// we're done when the path is consumed
		if path == "" {
			// we're successful only if the pieces are also fully consumed
			return i == len(pieces)-1
		}
	}
	// we're done when the pieces are consumed.
	// we're successful only if path is also fully consumed
	return path == ""
}
