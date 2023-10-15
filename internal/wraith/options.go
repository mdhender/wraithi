// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"context"
	"database/sql"
	"path/filepath"
)

type Option func(a *App) error

func WithAssetsLogging(flag bool) Option {
	return func(a *App) error {
		a.flags.log.assets = flag
		return nil
	}
}

func WithAssetsRoot(path string) Option {
	root, err := filepath.Abs(path)
	return func(a *App) error {
		a.root = root
		return err
	}
}

func WithDB(db *sql.DB) Option {
	return func(a *App) error {
		a.db = &DB{
			context: context.TODO(),
			db:      db,
		}
		return nil
	}
}

func WithServeSPA(flag bool) Option {
	return func(a *App) error {
		a.flags.spa = flag
		return nil
	}
}

func WithTemplates(path string) Option {
	templates, err := filepath.Abs(path)
	return func(a *App) error {
		a.templates.path = templates
		return err
	}
}
