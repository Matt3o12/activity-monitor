package main

import (
	"log"
	"net/http"

	"gopkg.in/pg.v4"

	"github.com/julienschmidt/httprouter"
)

var db *pg.DB

var Debug = true // TODO: Make dynamic.

func main() {
	db = InitConnection()
	defer func() {
		log.Println(db.Close())
	}()
	db.Exec(`SELECT set_config('log_statement', 'all', false);`)

	mux := httprouter.New()
	mux.ServeFiles("/static/*filepath", http.Dir("static"))

	get := func(path string, h UptimeCheckerHandler) {
		mux.GET(path, MainMiddleware(h))
	}

	post := func(path string, h UptimeCheckerHandler) {
		mux.POST(path, MainMiddleware(h))
	}

	// TODO: Add error 404 handler.
	get("/", dashboardHandler)
	get("/monitors/view/:id/", viewMonitorHandler)
	get("/monitors/add/", addMonitorGetHandler)
	post("/monitors/add/", addMonitorPostHandler)

	// Since exportLogsHandler has to do bulk writes,
	// we need to expose the writer directly.
	mux.GET("/monitors/logs/:id/export", exportLogsHandler)

	server := &http.Server{Addr: ":8092", Handler: mux}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Couldn't open http server: %v.\n", err)
	}
}
