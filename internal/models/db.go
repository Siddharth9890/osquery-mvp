package models

import (
	"database/sql"
)

type DBConn struct {
	Conn *sql.DB
}
