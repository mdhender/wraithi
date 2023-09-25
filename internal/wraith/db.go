// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	context context.Context
	db      *sql.DB
}
