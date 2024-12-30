package main

import (
	"clearway-test-task/internal/net/http"
	"clearway-test-task/internal/storage/authS"
	db2 "clearway-test-task/internal/storage/db"
)

func main() {
	db, err := db2.NewDb("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	auth := authS.NewAuthStorage(db)
	svr := http.NewHttpServer("0.0.0.0", "8080", auth, db)
	_ = svr.Listen()
}
