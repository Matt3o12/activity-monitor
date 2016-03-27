package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", IndexHandler("/", MainHandler(dashboardHandler)))

	server := &http.Server{Addr: ":8092", Handler: mux}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Couldn't open http server: %v.\n", err)
	}
}
