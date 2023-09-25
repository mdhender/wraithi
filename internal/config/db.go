// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package config

import "fmt"

// DBConfig holds mysql configuration.
type DBConfig struct {
	Host   string
	Port   int
	Name   string
	User   string
	Secret string
}

// DSN returns a connect string for a mysql database.
func (db DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8mb4&parseTime=True", db.User, db.Secret, db.Host, db.Port, db.Name)
}
