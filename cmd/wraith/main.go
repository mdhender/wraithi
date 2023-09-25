// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

// Package main implements a web server for Wraith.
package main

import (
	"database/sql"
	"github.com/mdhender/wraithi/internal/config"
	"github.com/mdhender/wraithi/internal/dot"
	"github.com/mdhender/wraithi/internal/wraith"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)

	defer func(started time.Time) {
		log.Printf("[main] elapsed time %v\n", time.Now().Sub(started))
	}(time.Now())
	log.Println("[main] entered")

	if err := dot.Load("WRAITH", false, false); err != nil {
		log.Fatalf("main: %+v\n", err)
	}

	cfg, err := config.Default()
	if err != nil {
		log.Fatal(err)
	} else if err = cfg.Load(); err != nil {
		log.Fatal(err)
	}

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg *config.Config) error {
	// set up database connection
	db, err := sql.Open("mysql", cfg.DB.DSN())
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		if err := db.Close(); err != nil {
			log.Printf("[main] db.close: %v\n", err)
		}
		log.Printf("[main] closed db\n")
	}(db)
	log.Printf("[main] connected to database\n")

	app, err := wraith.NewApp(cfg, db)
	if err != nil {
		return err
	}

	return app.Run()
}
