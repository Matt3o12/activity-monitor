package main

import (
	"log"
	"net/http"

	"gopkg.in/pg.v4"

	"github.com/julienschmidt/httprouter"
)

var db *pg.DB

func main() {
	db = InitConnection()
	defer func() {
		log.Println(db.Close())
	}()
	db.Exec(`SELECT set_config('log_statement', 'all', false);`)

	mux := httprouter.New()
	mux.ServeFiles("/static/*filepath", http.Dir("static"))

	// TODO: Add error 404 handler.
	mux.GET("/", dashboardHandler)
	mux.GET("/monitors/view/:id", viewMonitorHandler)
	mux.GET("/monitors/add/", addMonitorGetHandler)
	mux.POST("/monitors/add/", addMonitorPostHandler)

	server := &http.Server{Addr: ":8092", Handler: mux}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Couldn't open http server: %v.\n", err)
	}
}
