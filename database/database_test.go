package database

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexanderkarlis/sw-dnsbl/config"
	"github.com/alexanderkarlis/sw-dnsbl/graph/model"
)

const (
	testTable = "test_ip_details"
)

func TestDb(t *testing.T) {
	os.Setenv("APP_PORT", "8080")
	os.Setenv("MYSQL_DATABASE_NAME", "sw_dnsbl")
	os.Setenv("MYSQL_DATABASE_HOST", "0.0.0.0")
	os.Setenv("MYSQL_DATABASE_PORT", "3306")
	os.Setenv("MYSQL_USER", "mysql_admin")
	os.Setenv("MYSQL_PASSWORD", "password")
	os.Setenv("WORKER_POOL_SIZE", "99")
	os.Setenv("DNS_BLOCKLIST", "zen.spamhaus.org")
	os.Setenv("LOG_FILE", "app.log")
	os.Setenv("PERSIST_DB", "true")
	os.Setenv("DB_PATH", "./swdnsbl.db")

	// get config and remove db if already there
	conf := config.GetConfig()
	os.Remove(conf.DbPath)

	var db *Db

	t.Run("new_database_init/close_success", func(t *testing.T) {
		db, err = NewDb(conf)
		if err != nil {
			t.Errorf("db was not started")
		}
		assert.NotEqual(t, nil, db)

		err := db.Close()
		os.Remove(conf.DbPath)
		assert.Equal(t, nil, err)

	})

	t.Run("upsert_db_and_query", func(t *testing.T) {
		db, err = NewDb(conf)
		assert.NotEqual(t, nil, db)
		defer os.Remove(conf.DbPath)

		record := model.Record{
			UUID:         uuid.New().String(),
			CreatedAt:    int(time.Now().Unix()),
			UpdatedAt:    int(time.Now().Unix()),
			ResponseCode: "NXDOMAIN",
			IPAddress:    "127.0.0.43",
		}
		err := db.UpsertRecord(&record)
		require.Equal(t, nil, err)

		time.Sleep(time.Second / 2)

		r, err := db.QueryRecord("127.0.0.43")

		assert.Equal(t, "127.0.0.43", r.IPAddress)
		assert.Equal(t, "NXDOMAIN", r.ResponseCode)

		err = db.Close()
		assert.Equal(t, nil, err)

	})
}

func TestMySqlDEPRECATED(t *testing.T) {
	t.Skip()
	t.Run("test_query_record_ok", func(t *testing.T) {
		t.Skip()

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()
		appDb := &Db{db}

		columns := []string{"uuid", "created_at", "updated_at", "ip_address", "response_code"}
		uuid := uuid.New().String()
		ca := int(time.Now().Unix())
		ua := int(time.Now().Unix())
		ip := "127.0.0.2"
		rc := "127.0.0.2"

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT (.+) FROM ip_details WHERE").WithArgs("127.0.0.2").
			WillReturnRows(sqlmock.NewRows(columns).
				AddRow(uuid, ca, ua, ip, rc))
		mock.ExpectCommit()

		r, err := appDb.QueryRecord("127.0.0.2")
		if err != nil {
			t.Errorf("error retrieving data for db transaction.")
		}
		assert.Equal(t, "127.0.0.2", r.ResponseCode)
	})

	t.Run("test_query_record_bad", func(t *testing.T) {
		t.Skip()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()
		appDb := &Db{db}

		columns := []string{"uuid", "created_at", "updated_at", "ip_address", "response_code"}
		uuid := uuid.New().String()
		ca := int(time.Now().Unix())
		ua := int(time.Now().Unix())
		ip := "127.0.0.2"
		rc := "127.0.0.2"

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT (.+) FROM ip_details WHERE").WithArgs("127.0.0.2").
			WillReturnRows(sqlmock.NewRows(columns).
				AddRow(uuid, ca, ua, ip, rc))
		mock.ExpectCommit()

		r, err := appDb.QueryRecord("127.0.0.3")

		assert.Equal(t, "", r.ResponseCode)
		log.Println(err)
	})

	t.Run("test_upsert", func(t *testing.T) {
		t.Skip()

		// t.Skip()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()
		appDb := &Db{db}

		// columns := []string{"uuid", "created_at", "updated_at", "ip_address", "response_code"}
		uuid := uuid.New().String()
		ca := int(time.Now().Unix())
		ua := int(time.Now().Unix())
		ip := "127.0.0.2"

		updatedRc := "NXDOMAIN"
		updatedUa := ua + 1

		record2 := model.Record{
			UUID:         uuid,
			CreatedAt:    ca,
			UpdatedAt:    updatedUa,
			IPAddress:    ip,
			ResponseCode: updatedRc,
		}

		mock.ExpectBegin()
		mock.ExpectPrepare(`INSERT INTO ip_details`).
			ExpectExec().
			WithArgs(
				&record2.UUID,
				&record2.CreatedAt,
				&record2.UpdatedAt,
				&record2.IPAddress,
				&record2.ResponseCode,
				&record2.ResponseCode,
				&record2.UpdatedAt,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err = appDb.UpsertRecord(&record2)

		assert.Equal(t, nil, err)
	})

}
