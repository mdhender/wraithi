// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mdhender/wraithi/internal/authn"
	"github.com/mdhender/wraithi/internal/authn/google"
	"github.com/mdhender/wraithi/internal/config"
	"github.com/mdhender/wraithi/internal/nonces"
	"github.com/mdhender/wraithi/internal/semver"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func NewApp(cfg *config.Config, db *sql.DB) (*App, error) {
	if sb, err := os.Stat(cfg.App.Data); err != nil {
		return nil, fmt.Errorf("data: %w", err)
	} else if !sb.IsDir() {
		return nil, fmt.Errorf("data: not a directory")
	}
	if sb, err := os.Stat(cfg.App.Public); err != nil {
		return nil, fmt.Errorf("public: %w", err)
	} else if !sb.IsDir() {
		return nil, fmt.Errorf("public: not a directory")
	}
	if sb, err := os.Stat(cfg.App.Templates); err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	} else if !sb.IsDir() {
		return nil, fmt.Errorf("templates: not a directory")
	}

	ctx := context.Background()
	a := &App{
		context: ctx,
		data:    cfg.App.Data,
		db:      &DB{context: ctx, db: db},
		public:  cfg.App.Public,
		server: http.Server{
			Addr:           net.JoinHostPort(cfg.Server.Host, cfg.Server.Port),
			IdleTimeout:    cfg.Server.Timeout.Idle,
			MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
			ReadTimeout:    cfg.Server.Timeout.Read,
			WriteTimeout:   cfg.Server.Timeout.Write,
		},
		templates: cfg.App.Templates,
		version:   semver.Version{Major: 0, Minor: 1, Patch: 0}.String(),
	}

	nonceTTL := 5 * time.Minute
	for _, id := range strings.Split(cfg.Auth.Providers, ",") {
		switch strings.TrimSpace(id) {
		case "Facebook":
			return nil, fmt.Errorf("provider: %q: %w", id, authn.ErrUnknownProvider)
		case "GitHub":
			return nil, fmt.Errorf("provider: %q: %w", id, authn.ErrUnknownProvider)
		case "Google":
			clientId := os.Getenv("WRAITH_GOOGLE_CLIENT_ID")
			clientSecret := os.Getenv("WRAITH_GOOGLE_CLIENT_SECRET")
			nf := nonces.NewFactory(nonceTTL)
			a.authn = append(a.authn, google.Google(clientId, clientSecret, "http://localhost:8080/auth/callback/google", nf))
		default:
			return nil, fmt.Errorf("provider: %q: %w", id, authn.ErrUnknownProvider)
		}
	}

	a.server.Handler = a.Routes()

	return a, nil
}

type App struct {
	server  http.Server
	cookies struct {
		httpOnly bool
		secure   bool
	}
	tls struct {
		enabled  bool
		certFile string
		keyFile  string
	}
	authn           []authn.Provider
	context         context.Context
	db              *DB
	root            string
	data            string // path to data files
	public          string // path to public assets
	templates       string // path to templates
	timestampFormat string
	version         string
}

// Run will run the receiver's embedded http.Server and gracefully handle receipt of SIGTERM or SIGINT.
func (a *App) Run() error {
	// start the server in a new go routine
	go func(ctx context.Context) {
		log.Printf("[app] listening on %s\n", a.server.Addr)
		if err := a.Serve(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
		log.Printf("[app] server stopped gracefully\n")
	}(a.context)

	// create channels to catch signals
	stopCh, closeCh := a.SignalChannels()
	defer func() {
		closeCh()
	}()
	log.Println("[app] stopCh: notified: ", <-stopCh)

	return nil
}

// Serve is a wrapper around ListenAndServe and ListenAndServeTLS.
func (a *App) Serve(ctx context.Context) (err error) {
	a.context = ctx
	if a.tls.enabled {
		// ListenAndServeTLS jams in an untested TLS config.
		// Should probably validate against notes from
		// https://github.com/denji/golang-tls
		// and https://eli.thegreenplace.net/2021/go-https-servers-with-tls/.
		// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
		a.server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS13,
		}
		a.server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
		err = a.server.ListenAndServeTLS(a.tls.certFile, a.tls.keyFile)
	} else {
		err = a.server.ListenAndServe()
	}
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// SignalChannels creates channels to catch signals.
func (a *App) SignalChannels() (chan os.Signal, func()) {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	return stopCh, func() {
		close(stopCh)
	}
}

func (a *App) Shutdown(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		panic(err)
	}
	log.Printf("[app] app caught shutdown signal\n")
}
