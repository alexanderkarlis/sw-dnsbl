package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/alexanderkarlis/sw-dnsbl/config"
	"github.com/alexanderkarlis/sw-dnsbl/database"
	"github.com/alexanderkarlis/sw-dnsbl/dnsbl"
	"github.com/alexanderkarlis/sw-dnsbl/graph"
	"github.com/alexanderkarlis/sw-dnsbl/graph/generated"
	"github.com/alexanderkarlis/sw-dnsbl/middleware"
	"github.com/gorilla/mux"
)

const defaultPort = "8080"

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	(*w).Header().Set("Content-Type", "application/json")
}

// Alive endpoint check for helm
func Alive(w http.ResponseWriter, req *http.Request) {
	enableCors(&w)
	fmt.Fprintf(w, "ok")
}

// Ready endpoint check for helm
func Ready(w http.ResponseWriter, req *http.Request) {
	enableCors(&w)
	fmt.Fprintf(w, "ok")
}

func serve(ctx context.Context) (err error) {
	config := config.GetConfig()

	port := defaultPort
	if config.AppPort != "" {
		port = config.AppPort
	}

	// new db
	db, err := database.NewDb(config)
	if err != nil {
		log.Fatalln(err)
	}

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

	srv := handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{
				Resolvers: &resolver,
			},
		),
	)

	router.Use(middleware.Middleware())
	router.Handle("/", playground.Handler("GraphQL playground", "/graphql"))
	router.Handle("/graphql", srv)

	// helm charts had these in the config??
	router.HandleFunc("/alive", Alive)
	router.HandleFunc("/ready", Ready)

	go func() {
		if err = http.ListenAndServe(":"+port, router); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%+s\n", err)
		}
	}()

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	<-ctx.Done()

	log.Printf("Server stopped. Shutting down.")

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	return
}

func main() {
	logfile := os.Getenv("LOG_FILE")
	f, err := os.OpenFile("./logs/"+logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	log.SetOutput(f)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ossig := <-c
		log.Printf("received a system call: %+v", ossig)
		cancel()
	}()

	if err := serve(ctx); err != nil {
		log.Printf("failed to serve:+%v\n", err)
	}
}
