package graph

import (
	"github.com/alexanderkarlis/sw-dnsbl/database"
	"github.com/alexanderkarlis/sw-dnsbl/dnsbl"
)

// Resolver is the dep injection of other reqs
type Resolver struct {
	Database *database.Db
	Consumer *dnsbl.Consumer
}
