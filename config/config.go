package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// APIConfig is the overtall config read from env vars
type APIConfig struct {
	AppPort, DbPath, DbPort, DbName string
	DbUser, DbPassword, LogFile     string
	WorkerPoolsize                  int
	DNSBlockList                    []string
	PersistDb                       bool
}

// GetConfig function gets the configuration for the app and sets
func GetConfig() *APIConfig {
	var config APIConfig

	workerPoolsize := os.Getenv("WORKER_POOL_SIZE")
	workersize, err := strconv.Atoi(workerPoolsize)
	if err != nil {
		log.Println("Could not convert WORKER_POOL_SIZE to an `int`. Defaulting to `100`.")
		workersize = 100
	}

	dnsEnv := os.Getenv("DNS_BLOCKLIST")
	dnsList := strings.Split(dnsEnv, ",")
	if len(dnsList) == 0 {
		log.Println("Could not get dns blocklist. Defaulting to `zen.spamhaus.org`")
		dnsList = []string{"zen.spamhaus.org"}
	}

	persistDb := os.Getenv("PERSIST_DB")
	persistDbBool, err := strconv.ParseBool(persistDb)
	if err != nil {
		log.Println("Could not convert PERSIST_DB to a `bool`. Defaulting to `true`.")
		persistDbBool = true
	}

	config.AppPort = os.Getenv("APP_PORT")
	config.DbName = os.Getenv("MYSQL_DATABASE_NAME")
	config.DbPath = os.Getenv("DB_PATH")
	config.DbPort = os.Getenv("MYSQL_DATABASE_PORT")
	config.DbUser = os.Getenv("MYSQL_USER")
	config.DbPassword = os.Getenv("MYSQL_PASSWORD")
	config.LogFile = os.Getenv("LOG_FILE")
	config.PersistDb = persistDbBool
	config.DNSBlockList = dnsList
	config.WorkerPoolsize = workersize

	log.Printf("CONFIG SETTINGS: %+v\n", config)
	return &config
}
