// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"github.com/mdhender/wraithi/internal/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewApp(cfg *config.Config, db *sql.DB) (*App, error) {
	ctx := context.Background()
	a := &App{
		context: ctx,
		db:      &DB{context: ctx, db: db},
	}
	return a, nil
}

type App struct {
	context context.Context
	db      *DB
	server  http.Server
	tls     struct {
		enabled  bool
		certFile string
		keyFile  string
	}
}

// Run will run the receiver's embedded http.Server and gracefully handle receipt of SIGTERM or SIGINT.
func (a *App) Run() error {
	// start the server in a new go routine
	go func(ctx context.Context) {
		log.Printf("[app] server started\n")
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
