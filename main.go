package main

import (
	"database/sql"
	"log"
	_ "github.com/lib/pq"
	db "testuhpostgres/db/sqlc"
	"testuhpostgres/api"
	"testuhpostgres/rdstore"
	"os"
	"github.com/go-redis/redis/v7"
)

func init() {
	//Initializing redis
	dsn := os.Getenv("REDIS_URL")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	rdstore.Client = redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	_, err := rdstore.Client.Ping().Result()
	if err != nil {
		panic(err)
	}
}

func main() {

	// conn, err := sql.Open("postgres", `postgresql://postgres:anythingtobi@sk@@localhost:5432/postgres?sslmode=disable`)
	conn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
	  log.Fatal(err)
	}

	err = conn.Ping()

	if err != nil {
		log.Fatal(err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	// http.Handle("/metrics", promhttp.Handler())

	//Here we are telling prometheus to keep track of ordersPlaced metric
	// prometheus.MustRegister(api.RequestsToCreateAccount)

	// addr := ":" + os.Getenv("PORT")
	addr := ":7000"
	err = server.Start(addr)
	if err != nil {
		log.Fatal("Cannot start server:", err)
	}

	// testuhHandler := func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprint(w, "Hello, UrbanHive!")
	// }

	// http.HandleFunc("/", testuhHandler)
	

	// // addr := ":8080" use this port to testuh locally on your pc if you don't have the PORT env variable set

	// log.Fatal(http.ListenAndServe(addr, nil))
}