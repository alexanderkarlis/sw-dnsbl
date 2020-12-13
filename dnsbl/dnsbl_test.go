package dnsbl

import (
	"os"
	"testing"

	"github.com/alexanderkarlis/sw-dnsbl/config"
	"github.com/alexanderkarlis/sw-dnsbl/database"
	"github.com/stretchr/testify/assert"
)

func TestConsumer(t *testing.T) {
	os.Setenv("MYSQL_DATABASE_NAME", "sw_dnsbl")
	os.Setenv("MYSQL_DATABASE_HOST", "0.0.0.0")
	os.Setenv("MYSQL_DATABASE_PORT", "3306")
	os.Setenv("MYSQL_USER", "mysql_admin")
	os.Setenv("MYSQL_PASSWORD", "password")
	os.Setenv("WORKER_POOL_SIZE", "99")
	os.Setenv("DNS_BLOCKLIST", "zen.spamhaus.org")
	os.Setenv("LOG_FILE", "app.log")

	c := config.GetConfig()
	db, err := database.NewDb(c)

	defer db.Close()
	if err != nil {
		t.Error(err)
	}
	t.Run("new_consumer_success", func(t *testing.T) {
		consumer := NewConsumer(db, c)
		assert.NotEqual(t, nil, consumer)
	})

	t.Run("add_to_queue_success", func(t *testing.T) {
		consumer := NewConsumer(db, c)
		addedToQueue := consumer.Queue([]string{"127.0.0.1"})
		assert.Equal(t, true, addedToQueue)
	})

	t.Run("add_to_queue_fail", func(t *testing.T) {
		os.Setenv("WORKER_POOL_SIZE", "0")
		cc := config.GetConfig()
		consumer := NewConsumer(db, cc)

		var addedToQueue bool
		for {
			addedToQueue = consumer.Queue([]string{"127.0.0.1", "127.0.0.2", "127.0.0.255"})
			if !addedToQueue {
				break
			}
		}
		assert.Equal(t, false, addedToQueue)
	})
}
