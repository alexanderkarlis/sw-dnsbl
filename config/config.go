package config

import (
	"log"
	"os"
	"strings"

	"strconv"
)

// APIConfig is the overtall config read from env vars
type APIConfig struct {
	AppPort, DbHost, DbPort, DbName string
	DbUser, DbPassword, LogFile     string
	WorkerPoolsize                  int
	DNSBlockList                    []string
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

	config.AppPort = os.Getenv("APP_PORT")
	config.DbName = os.Getenv("MYSQL_DATABASE_NAME")
	config.DbHost = os.Getenv("MYSQL_DATABASE_HOST")
	config.DbPort = os.Getenv("MYSQL_DATABASE_PORT")
	config.DbUser = os.Getenv("MYSQL_USER")
	config.DbPassword = os.Getenv("MYSQL_PASSWORD")
	config.DNSBlockList = dnsList
	config.LogFile = os.Getenv("LOG_FILE")
	config.WorkerPoolsize = workersize

	log.Printf("%+v\n", config)
	return &config
}
