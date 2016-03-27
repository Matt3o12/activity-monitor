package main

import (
	"fmt"
	"log"
	"net/http"
)

func handleStatic(dir string) {
	path := fmt.Sprintf("/%v/", dir)
	http.Handle(path, http.StripPrefix(path, http.FileServer(http.Dir(dir))))
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/", http.FileServer(http.Dir("templates")))

	if err := http.ListenAndServe(":8092", nil); err != nil {
		log.Fatalf("Couldn't open http server: %v.\n", err)
	}
}
