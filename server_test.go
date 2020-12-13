package main

import (
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/alexanderkarlis/sw-dnsbl/config"
	"github.com/alexanderkarlis/sw-dnsbl/database"
	"github.com/alexanderkarlis/sw-dnsbl/dnsbl"
	"github.com/alexanderkarlis/sw-dnsbl/graph"
	"github.com/alexanderkarlis/sw-dnsbl/graph/generated"
	"github.com/alexanderkarlis/sw-dnsbl/graph/model"
	"github.com/alexanderkarlis/sw-dnsbl/middleware"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type createToken struct {
	CreateToken model.Token
}

type enqueue struct {
	Enqueue bool
}

type getipdetails struct {
	GetIPDetails model.Record
}

type gqlError struct {
	errors []struct {
		message string
		path    []string
	}
}

func isIPInResponseList(ip string, ips []string) bool {
	for i := 0; i < len(ips); i++ {
		if ip == ips[i] {
			return true
		}
	}
	return false
}

func TestSWDNSBLServer(t *testing.T) {
	// t.Skip()
	config := config.GetConfig()

	port := "8080"
	if config.AppPort != "" {
		port = config.AppPort
	}

	// new db
	db, err := database.NewDb(config)

	// new consumer
	consumer := dnsbl.NewConsumer(db, config)
	if err != nil {
		log.Fatalln(err)
	}
	resolver := graph.Resolver{
		Database: db,
		Consumer: consumer,
	}

	// new router and auth layer
	router := mux.NewRouter()

	router.Use(middleware.Middleware())

	srv := handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{
				Resolvers: &resolver,
			},
		),
	)

	// gqp client
	c := client.New(router)

	router.Handle("/", srv)

	go func() {
		if err = http.ListenAndServe(":"+port, router); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%+s\n", err)
		}
	}()

	// create structs to reuse for the tests below (auth, enqueue, query)
	var auth createToken
	var enqueueResp enqueue
	var ipdetails getipdetails

	// sign in mutation
	authMutation := `
		mutation {
			createToken(
				data: {
					username: "secureworks"
					password: "supersecret"
				}
			),
			{
				bearer_token
			}
		}
	`

	c.MustPost(authMutation, &auth)
	require.Equal(t, "Bearer", strings.Split(auth.CreateToken.BearerToken, " ")[0])

	// add some preliminary ip addresses
	enqueueMutation := `
	mutation {
		enqueue(
			ips: ["127.0.0.2", "127.0.0.23", "127.0.0.255"]
		),
	}
	`

	authHeader := client.AddHeader("Authorization", auth.CreateToken.BearerToken)
	err = c.Post(enqueueMutation, &enqueueResp, authHeader)

	// wait for enqueue lookups to finish
	time.Sleep(3 * time.Second)

	t.Run("authenticate_fail", func(t *testing.T) {
		var resp createToken
		// t.Skip()
		// var respErr gqlError
		badAuth := `[{"message":"invalid credentials","path":["createToken"]}]`
		badAuthMutation := `
		mutation {
			createToken(
				data: {
					username: "bad"
					password: "secret"
				}
			),
			{
				bearer_token
			}
		}
		`
		err = c.Post(badAuthMutation, &resp)
		log.Printf("%+v\n", resp)
		assert.EqualError(t, err, badAuth)
	})

	t.Run("enqueue_no_auth", func(t *testing.T) {
		enqueueMutation = `
		mutation {
			enqueue(
				ips: ["127.0.0.2", "127.0.0.23", "127.0.0.255"]
			),
		}
		`
		time.Sleep(3 * time.Second)
		err := c.Post(enqueueMutation, &enqueueResp)
		assert.EqualError(t, err, `[{"message":"missing auth token","path":["enqueue"]}]`)
	})

	t.Run("enqueue_new_and_update_success", func(t *testing.T) {
		enqueueMutation = `
		mutation {
			enqueue(
				ips: ["127.0.0.1", "127.0.0.3", "127.0.0.4", "127.0.0.9", "127.0.0.10", "127.0.0.11", "127.0.0.255", "127.0.0.23"]
			),
		}
		`
		c.Post(enqueueMutation, &enqueueResp, authHeader)
		// Sleep 5 seconds to ensure values are in the DB
		time.Sleep(5 * time.Second)
		assert.Equal(t, true, enqueueResp.Enqueue)
	})

	t.Run("query_ip_no_auth", func(t *testing.T) {
		getDetailsQuery := `
		query {
			getIPDetails(
			  ip: "127.0.0.1"
			),
			{
			  uuid
			  created_at
			  updated_at
			  response_code
			  ip_address
			}
		  }
		`
		err := c.Post(getDetailsQuery, &ipdetails)
		assert.EqualError(t, err, `[{"message":"missing auth token","path":["getIPDetails"]}]`)
	})

	t.Run("query_ip_success_127.0.0.1_NXDOMAIN", func(t *testing.T) {
		getDetailsQuery := `
		query {
			getIPDetails(
			  ip: "127.0.0.1"
			),
			{
			  uuid
			  created_at
			  updated_at
			  response_code
			  ip_address
			}
		  }
		`
		c.Post(getDetailsQuery, &ipdetails, authHeader)
		log.Println("IP DETAILS", ipdetails.GetIPDetails.ResponseCode)
		assert.Equal(t, "NXDOMAIN", ipdetails.GetIPDetails.ResponseCode)
	})

	t.Run("query_ip_success_127.0.0.255_NXDOMAIN", func(t *testing.T) {
		getDetailsQuery := `
		query {
			getIPDetails(
			  ip: "127.0.0.255"
			),
			{
			  uuid
			  created_at
			  updated_at
			  response_code
			  ip_address
			}
		  }
		`
		c.Post(getDetailsQuery, &ipdetails, authHeader)
		assert.Equal(t, "NXDOMAIN", ipdetails.GetIPDetails.ResponseCode)
	})

	t.Run("query_ip_success_127.0.0.2", func(t *testing.T) {
		getDetailsQuery := `
		query {
			getIPDetails(
			  ip: "127.0.0.2"
			),
			{
			  uuid
			  created_at
			  updated_at
			  response_code
			  ip_address
			}
		  }
		`
		c.Post(getDetailsQuery, &ipdetails, authHeader)
		assert.Equal(t, true, isIPInResponseList(ipdetails.GetIPDetails.ResponseCode, []string{"127.0.0.2", "127.0.0.10", "127.0.0.4"}))
	})

	t.Run("query_ip_success_127.0.0.3", func(t *testing.T) {
		getDetailsQuery := `
		query {
			getIPDetails(
			  ip: "127.0.0.3"
			),
			{
			  uuid
			  created_at
			  updated_at
			  response_code
			  ip_address
			}
		  }
		`
		c.Post(getDetailsQuery, &ipdetails, authHeader)
		assert.Equal(t, "127.0.0.3", ipdetails.GetIPDetails.ResponseCode)
	})

	t.Run("query_ip_success_127.0.0.4", func(t *testing.T) {
		getDetailsQuery := `
		query {
			getIPDetails(
			  ip: "127.0.0.4"
			),
			{
			  uuid
			  created_at
			  updated_at
			  response_code
			  ip_address
			}
		  }
		`
		c.Post(getDetailsQuery, &ipdetails, authHeader)
		assert.Equal(t, "127.0.0.4", ipdetails.GetIPDetails.ResponseCode)
	})

	t.Run("query_ip_success_127.0.0.9", func(t *testing.T) {
		getDetailsQuery := `
		query {
			getIPDetails(
			  ip: "127.0.0.9"
			),
			{
			  uuid
			  created_at
			  updated_at
			  response_code
			  ip_address
			}
		  }
		`
		//127.0.0.2, 127.0.0.9
		c.Post(getDetailsQuery, &ipdetails, authHeader)
		assert.Equal(t, true, isIPInResponseList(ipdetails.GetIPDetails.ResponseCode, []string{"127.0.0.2", "127.0.0.9"}))
	})

	t.Run("query_ip_success_127.0.0.10", func(t *testing.T) {
		getDetailsQuery := `
		query {
			getIPDetails(
			  ip: "127.0.0.10"
			),
			{
			  uuid
			  created_at
			  updated_at
			  response_code
			  ip_address
			}
		  }
		`
		c.Post(getDetailsQuery, &ipdetails, authHeader)
		assert.Equal(t, "127.0.0.10", ipdetails.GetIPDetails.ResponseCode)
	})

	t.Run("query_ip_success_127.0.0.11", func(t *testing.T) {
		getDetailsQuery := `
		query {
			getIPDetails(
			  ip: "127.0.0.11"
			),
			{
			  uuid
			  created_at
			  updated_at
			  response_code
			  ip_address
			}
		  }
		`
		c.Post(getDetailsQuery, &ipdetails, authHeader)
		assert.Equal(t, "127.0.0.10", ipdetails.GetIPDetails.ResponseCode)
	})

	t.Run("query_ip_empty", func(t *testing.T) {
		getDetailsQuery := `
		query {
			getIPDetails(
			  ip: "127.0.0.80"
			),
			{
			  uuid
			  created_at
			  updated_at
			  response_code
			  ip_address
			}
		  }
		`
		err := c.Post(getDetailsQuery, &ipdetails, authHeader)
		log.Println("IP DETAILS", ipdetails.GetIPDetails.ResponseCode)
		assert.EqualError(t, err, `[{"message":"sql: no rows in result set","path":["getIPDetails"]}]`)
	})
}
