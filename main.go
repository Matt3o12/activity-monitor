package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {

	mux := httprouter.New()
	mux.ServeFiles("/static/*filepath", http.Dir("static"))

	mux.GET("/", dashboardHandler)
	mux.GET("/monitors/view/:id", viewMonitorHandler)
	mux.GET("/monitors/add/", addMonitorGetHandler)
	mux.POST("/monitors/add/", addMonitorPostHandler)

	server := &http.Server{Addr: ":8092", Handler: mux}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Couldn't open http server: %v.\n", err)
	}
}
