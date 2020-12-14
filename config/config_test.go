package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
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

	c := GetConfig()
	assert.Equal(t, c.AppPort, "8080")
	assert.Equal(t, c.DbName, "sw_dnsbl")
	assert.Equal(t, c.DbPath, "./swdnsbl.db")
	assert.Equal(t, c.PersistDb, true)
	assert.Equal(t, c.DbPort, "3306")
	assert.Equal(t, c.DbUser, "mysql_admin")
	assert.Equal(t, c.DbPassword, "password")
	assert.Equal(t, c.WorkerPoolsize, 99)
	assert.Equal(t, c.DNSBlockList, []string{"zen.spamhaus.org"})
	assert.Equal(t, c.LogFile, "app.log")
}
