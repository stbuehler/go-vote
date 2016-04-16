package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stbuehler/go-vote/backend"
	"github.com/stbuehler/go-vote/frontend"
	"github.com/stbuehler/go-vote/static"
	"net/http"
)

func main() {
	db, err := sql.Open("sqlite3", "elections.sqlite")
	if nil != err {
		panic(err)
	}
	edb, err := backend.ConnectDatabase(db)
	if nil != err {
		panic(err)
	}

	mux := http.NewServeMux()
	frontend.Frontend{edb}.BindServeMux(mux, "")
	edb.BindServeMux(mux, "")
	static.BindServeMux(mux, "")
	http.ListenAndServe(":8080", mux)
}
