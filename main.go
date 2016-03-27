package main

import (
	"log"
	"net/http"
)

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", dashboardHandler)

	if err := http.ListenAndServe(":8092", nil); err != nil {
		log.Fatalf("Couldn't open http server: %v.\n", err)
	}
}
