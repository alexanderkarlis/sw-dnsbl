package dnsbl

import (
	"log"
	"sync"
	"time"

	"github.com/alexanderkarlis/godnsbl"
	"github.com/google/uuid"

	"github.com/alexanderkarlis/sw-dnsbl/config"
	"github.com/alexanderkarlis/sw-dnsbl/database"
	"github.com/alexanderkarlis/sw-dnsbl/graph/model"
)

// Consumer type
type Consumer struct {
	wg        sync.WaitGroup
	db        *database.Db
	inputChan chan int
	jobsChan  chan []string
	blDomains []string
	quitChan  chan struct{}
}

// ResultSet from godnsbl.Lookup()
type ResultSet []godnsbl.RBLResults

// NewConsumer function returns a consumer to be run for the alotted job queue.
func NewConsumer(db *database.Db, c *config.APIConfig) *Consumer {
	poolsize := c.WorkerPoolsize
	blDomains := c.DNSBlockList

	consumer := Consumer{
		wg:        sync.WaitGroup{},
		db:        db,
		inputChan: make(chan int, 1),
		jobsChan:  make(chan []string, poolsize),
		quitChan:  make(chan struct{}),
		blDomains: blDomains,
	}

	consumer.wg.Add(1)
	go consumer.worker()
	log.Printf("Started new Consumer with poolsize %d\n", poolsize)
	return &consumer
}

// Queue function
func (c *Consumer) Queue(ips []string) bool {
	select {
	case c.jobsChan <- ips:
		log.Printf("added %d ips to check against blist\n", len(ips))
		return true
	default: // buffer is full
		log.Printf("queue is full\n")
		return false
	}
}

// worker function that takes in a array of sources and IPs to check,
// and runs the godnsbl.Lookup function
func (c *Consumer) worker() {
	defer c.wg.Done()
	for {
		select {
		case <-c.quitChan:
			log.Println("Stop chan received. Exiting function")
			return
		case ips := <-c.jobsChan:
			log.Printf("in jobs chan, received %+v\n", ips)

			sources := c.blDomains
			results := make([]godnsbl.Result, len(sources))

			for _, ip := range ips {
				log.Printf("looking up %s", ip)
				for _, source := range sources {
					rbl := godnsbl.Lookup(source, ip)

					timeNow := time.Now().Unix()
					respCode := "NXDOMAIN"

					if !rbl.Results[0].Error {
						respCode = rbl.Results[0].Code
					}

					record := &model.Record{
						UUID:         uuid.New().String(),
						CreatedAt:    int(timeNow),
						UpdatedAt:    int(timeNow),
						IPAddress:    ip,
						ResponseCode: respCode,
					}

					err := c.db.UpsertRecord(record)
					if err != nil {
						log.Println("inserting record failed!", err)
					}

					// TODO: for debugging purposes
					if len(rbl.Results) == 0 {
						results = append(results, godnsbl.Result{})
					} else {
						results = append(results, rbl.Results[0])
					}
				}
			}
		}
	}
}

// ProcessIps function takes in a array of sources and IPs to check,
// and runs the godnsbl.Lookup function
func ProcessIps(sources, ips []string) *[]godnsbl.Result {
	results := make([]godnsbl.Result, len(sources))
	for _, ip := range ips {
		log.Printf("running for %s", ip)
		for i, source := range sources {
			rbl := godnsbl.Lookup(source, ip)
			if len(rbl.Results) == 0 {
				results[i] = godnsbl.Result{}
			} else {
				results[i] = rbl.Results[0]
			}
		}
	}
	return &results
}
