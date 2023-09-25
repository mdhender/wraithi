// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

// Package config implements a configuration for Wraith game engine and web server.
package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/peterbourgon/ff/v3"
)

// Config defines configuration information for the application.
type Config struct {
	Debug bool
	App   struct {
		Data            string // path to data files
		Public          string // path to public assets
		Root            string
		Templates       string // path to templates
		TimestampFormat string
	}
	Cookies struct {
		HttpOnly bool
		Secure   bool
	}
	DB       DBConfig
	FileName string
	Server   struct {
		Scheme  string
		Host    string
		Port    string
		Timeout struct {
			Idle  time.Duration
			Read  time.Duration
			Write time.Duration
		}
		Key  string
		Salt string
	}
}

// Default returns a default configuration.
// These are the values without loading the environment, configuration file, or command line.
func Default() *Config {
	var cfg Config
	cfg.App.Root = "."
	cfg.App.Data = filepath.Join(cfg.App.Root, "testdata")
	cfg.App.Public = filepath.Join(cfg.App.Root, "web")
	cfg.App.Templates = filepath.Join(cfg.App.Root, "templates")
	cfg.App.TimestampFormat = "2006-01-02T15:04:05.99999999Z"
	cfg.DB.Port = 3306
	cfg.Server.Host = "localhost"
	cfg.Server.Key = "hush.abba.hu$h"
	cfg.Server.Port = "3000"
	cfg.Server.Salt = "nacl-clan"
	cfg.Server.Scheme = "http"
	cfg.Server.Timeout.Idle = 10 * time.Second
	cfg.Server.Timeout.Read = 5 * time.Second
	cfg.Server.Timeout.Write = 10 * time.Second
	return &cfg
}

// Load updates the values in a Config in this order:
//  1. It will load a configuration file if one is given on the
//     command line via the `-config` flag. If provided, the file
//     must contain a valid JSON object.
//  2. Environment variables, using the prefix `GOBBS`
//  3. Command line flags
func (cfg *Config) Load() error {
	fs := flag.NewFlagSet("config", flag.ExitOnError)

	fs.BoolVar(&cfg.Cookies.HttpOnly, "cookies-http-only", cfg.Cookies.HttpOnly, "set HttpOnly flag on cookies")
	fs.BoolVar(&cfg.Cookies.Secure, "cookies-secure", cfg.Cookies.Secure, "set Secure flag on cookies")
	fs.BoolVar(&cfg.Debug, "debug", cfg.Debug, "log debug information (optional)")
	fs.DurationVar(&cfg.Server.Timeout.Idle, "idle-timeout", cfg.Server.Timeout.Idle, "http idle timeout")
	fs.DurationVar(&cfg.Server.Timeout.Read, "read-timeout", cfg.Server.Timeout.Read, "http read timeout")
	fs.DurationVar(&cfg.Server.Timeout.Write, "write-timeout", cfg.Server.Timeout.Write, "http write timeout")
	fs.IntVar(&cfg.DB.Port, "db-port", cfg.DB.Port, "port of mysql database")
	fs.StringVar(&cfg.App.Data, "data-path", cfg.App.Data, "path containing data files")
	fs.StringVar(&cfg.App.Public, "public", cfg.App.Public, "path to serve web assets from")
	fs.StringVar(&cfg.App.Root, "root", cfg.App.Root, "path to treat as root for relative file references")
	fs.StringVar(&cfg.App.Templates, "templates", cfg.App.Templates, "path to template files")
	fs.StringVar(&cfg.DB.Host, "db-host", cfg.DB.Host, "host of mysql database")
	fs.StringVar(&cfg.DB.Name, "db-name", cfg.DB.Name, "name of mysql database")
	fs.StringVar(&cfg.DB.Secret, "db-secret", cfg.DB.Secret, "secret for mysql database")
	fs.StringVar(&cfg.DB.User, "db-user", cfg.DB.User, "user in mysql database")
	fs.StringVar(&cfg.FileName, "config", cfg.FileName, "config file (optional)")
	fs.StringVar(&cfg.Server.Host, "host", cfg.Server.Host, "host name (or IP) to listen on")
	fs.StringVar(&cfg.Server.Key, "key", cfg.Server.Key, "set key for signing tokens")
	fs.StringVar(&cfg.Server.Port, "port", cfg.Server.Port, "port to listen on")
	fs.StringVar(&cfg.Server.Salt, "salt", cfg.Server.Salt, "set salt for hashing passwords")
	fs.StringVar(&cfg.Server.Scheme, "scheme", cfg.Server.Scheme, "http scheme, either 'http' or 'https'")

	err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("WRAITH"), ff.WithConfigFileFlag("config"), ff.WithConfigFileParser(ff.JSONParser))
	if err != nil {
		return err
	}

	if cfg.App.Root, err = filepath.Abs(cfg.App.Root); err != nil {
		return fmt.Errorf("root: %w", err)
	}
	if cfg.DB.Port == 0 {
		cfg.DB.Port = 3306
	}
	if cfg.App.Data, err = filepath.Abs(cfg.App.Data); err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if cfg.App.Public, err = filepath.Abs(cfg.App.Public); err != nil {
		return fmt.Errorf("public: %w", err)
	}
	if cfg.App.Templates, err = filepath.Abs(cfg.App.Templates); err != nil {
		return fmt.Errorf("templates: %w", err)
	}

	return nil
}
