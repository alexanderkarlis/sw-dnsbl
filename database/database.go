package database

import (
	"database/sql"
	"log"
	"os"

	"github.com/alexanderkarlis/sw-dnsbl/config"
	"github.com/alexanderkarlis/sw-dnsbl/graph/model"

	// in order to get around the go-lint; here's a comment
	// _ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3" //Required for sql driver registration
)

var (
	db  *sql.DB
	err error
)

// Db is the type for MySql Connection
type Db struct {
	Conn *sql.DB
}

// Exists reports whether the named file or directory exists.
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// NewDb method returns a new MySql instance
func NewDb(c *config.APIConfig) (*Db, error) {
	dbPath := "./swdnsbl.db"

	if c.DbPath != "" {
		dbPath = c.DbPath
	}
	log.Println(dbPath)
	if !c.PersistDb {
		os.Remove(dbPath)
	}

	db, err = sql.Open("sqlite3", dbPath)

	if err != nil {
		return nil, err
	}

	// read create table script sql
	// content, err := ioutil.ReadFile("../scripts/ip/init.sql")
	// if err != nil {
	// 	panic(err)
	// }

	if !exists(dbPath) {
		createTable := `
			CREATE TABLE IF NOT EXISTS ip_details (
				ip_address TEXT PRIMARY KEY NOT NULL, 
				uuid TEXT NOT NULL, 
				response_code text,
				created_at TEXT, 
				updated_at TEXT
			);
		`
		_, err = db.Exec(createTable)
		if err != nil {
			log.Println(err)
			log.Fatalf("create table sql statement FAILED: %s\n", createTable)
		}
	}

	log.Println("db connection started...")

	return &Db{db}, nil
}

// UpsertRecord func inserts if not exists, else update record
func (db *Db) UpsertRecord(r *model.Record) error {
	if err != nil {
		log.Println("this is the error:", err)
		return err
	}

	log.Printf("%+v\n", r)
	// upsertQuery := `
	// INSERT INTO ip_details(uuid, created_at, updated_at, ip_address, response_code)
	// VALUES (
	// 	?,
	// 	?,
	// 	?,
	// 	?,
	// 	?
	// ) ON DUPLICATE KEY
	// UPDATE response_code = ?,
	// 	updated_at = ?;
	// `

	upsertQuery := `
		INSERT INTO ip_details( 
			ip_address,
			uuid,
			response_code,
			created_at,
			updated_at 
		) VALUES( ?, ?, ?, ?, ? )
		ON CONFLICT(ip_address) DO UPDATE SET
			response_code = ?,
			updated_at = ?
	`

	upsertStmt, err := db.Conn.Prepare(upsertQuery)

	if err != nil {
		log.Println(err)
		log.Println("error on upsert preparation")
		return err
	}
	upsertResult, err := upsertStmt.Exec(
		r.IPAddress,
		r.UUID,
		r.ResponseCode,
		r.CreatedAt,
		r.UpdatedAt,
		r.ResponseCode,
		r.UpdatedAt,
	)
	if err != nil {
		log.Println(err)
		log.Println("error on upsert")
		return err
	}
	rowsAffected, err := upsertResult.RowsAffected()
	if err != nil {
		log.Println(err)
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
	// TODO: add custom struct tags so that for easy unmarshalling

	var r model.Record

	selectQuery := `
		SELECT
			ip_address,
			uuid,
			created_at,
			updated_at,
			response_code
		FROM ip_details
		WHERE ip_address = ?
	`
	// selectQuery := `
	// 	select * from ip_details
	// 	where ip_address = ?;
	// `
	selectStmt, err := db.Conn.Prepare(selectQuery)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer selectStmt.Close()

	// _, err = selectStmt.Exec(ip)
	// if err != nil {
	// 	log.Println("here", err)
	// 	return nil, err
	// }

	err = db.Conn.QueryRow(selectQuery, ip).Scan(
		&r.IPAddress,
		&r.UUID,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.ResponseCode,
	)

	if err != nil {
		log.Println("Query Row Error", err)
		return nil, err
	}
	return &r, nil
}

// Close the connection to Sqlite3
func (db *Db) Close() error {
	return db.Conn.Close()
}
