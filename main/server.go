package main

import (
	"db-hw/config"
	"db-hw/internal/db"
	"db-hw/internal/router"
	"github.com/jackc/pgx"
	"log"
	"net/http"
)

func main() {
	var err error
	db.DB, err = pgx.NewConnPool(config.DBConfig)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = db.InitDB(db.DB)

	db.DB, err = pgx.NewConnPool(config.DBConfig)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = db.InitDB(db.DB)

	if err != nil {
		log.Fatal(err)
		return
	}
	r := router.GetRouter()

	log.Println("Serving at HTTP port 5000")
	err = http.ListenAndServe(":5000", r)
	if err != nil {
		log.Fatal(err)
		return
	}
}
