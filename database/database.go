package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/alexanderkarlis/sw-dnsbl/config"
	"github.com/alexanderkarlis/sw-dnsbl/graph/model"

	// in order to get around the go-lint; here's a comment
	_ "github.com/go-sql-driver/mysql"
)

var (
	db  *sql.DB
	err error
)

// Db is the type for MySql Connection
type Db struct {
	Conn *sql.DB
}

// NewDb method returns a new MySql instance
func NewDb(c *config.APIConfig) (*Db, error) {
	dbHost := c.DbHost
	if c.DbHost == "" {
		dbHost = "0.0.0.0"
	}
	log.Println(dbHost)
	connString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", c.DbUser, c.DbPassword, dbHost, c.DbName)
	db, err = sql.Open("mysql", connString)
	if err != nil {
		return nil, err
	}
	log.Println("starting db connection...")

	return &Db{db}, nil
}

// UpsertRecord func inserts if not exists, else update record
func (db *Db) UpsertRecord(r *model.Record) error {
	tx, err := db.Conn.Begin()
	defer tx.Commit()

	if err != nil {
		log.Println("this is the error:", err)
		return err
	}

	log.Printf("%+v\n", r)
	upsertQuery := `
	INSERT INTO ip_details(uuid, created_at, updated_at, ip_address, response_code)
	VALUES (
		?,
		?,
		?,
		?,
		?
	) ON DUPLICATE KEY
	UPDATE response_code = ?,
		updated_at = ?;
	`
	// upsertStmt, err := db.Conn.Prepare(upsertQuery)
	upsertStmt, err := tx.Prepare(upsertQuery)
	if err != nil {
		log.Println("error on upsert preparation")
		return err
	}
	upsertResult, err := upsertStmt.Exec(r.UUID, r.CreatedAt, r.UpdatedAt, r.IPAddress, r.ResponseCode, r.ResponseCode, r.UpdatedAt)
	if err != nil {
		log.Println("error on upsert")
		return err
	}
	rowsAffected, err := upsertResult.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 2 {
		log.Printf("row data updated for ip address -> %s\n", r.IPAddress)
	} else {
		log.Printf("added new row: %s\n", r.IPAddress)
	}
	return nil
}

// QueryRecord func searches for a record
func (db *Db) QueryRecord(ip string) (*model.Record, error) {
	// NOTE: can not figure out how to add custom struct tags so that
	// db records can be more easily unmarshalled
	r := &model.Record{}

	tx, err := db.Conn.Begin()
	defer tx.Commit()
	if err != nil {
		return r, err
	}

	selectQuery := `
		SELECT * 
		FROM ip_details
		WHERE ip_address = ?
	`
	err = tx.QueryRow(selectQuery, ip).Scan(
		&r.UUID,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.IPAddress,
		&r.ResponseCode,
	)
	if err != nil {
		log.Println(err)
		return r, err
	}
	return r, nil
}

// Close the connection to MySql
func (db *Db) Close() {
	db.Conn.Close()
}
