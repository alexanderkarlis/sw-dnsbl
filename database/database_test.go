package database

import (
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/alexanderkarlis/sw-dnsbl/graph/model"
)

const (
	testTable = "test_ip_details"
)

func TestDb(t *testing.T) {
	t.Run("new_database_init_success", func(t *testing.T) {
		assert.NotEqual(t, nil, db)
	})

	t.Run("test_query_record_ok", func(t *testing.T) {
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
