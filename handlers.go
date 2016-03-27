package main

import (
	"log"
	"net/http"

	"github.com/golang/go/src/text/template"
)

var err500TemplateNotLoading = []byte(`<!DOCTYPE html><html>
<head><title>Error 500</title></head>
<body><h1>Error 500</h1><p>Could not load error template</p></body></html>`)

var err500TemplateNotExecuting = []byte(`<!DOCTYPE html><html>
<head><title>Error 500</title></head>
<body><h1>Error 500</h1><p>Could not execute template.</p></body></html>`)

func handleServerError(org error, w http.ResponseWriter) {
	if org == nil {
		return
	}

	log.Println("An unexpected error 500 occured: ", org)
	w.WriteHeader(500)
	t, err := template.ParseFiles("templates/error500.html")
	if err != nil {
		_, _ = w.Write(err500TemplateNotLoading)
		log.Println("Could not load template.", err)
		return
	}

	if err := t.Execute(w, org.Error()); err != nil {
		_, _ = w.Write(err500TemplateNotExecuting)
		log.Println("Error while executing template.", err)
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("New request! :)")
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		handleServerError(err, w)
		return
	}

	err = t.Execute(w, makeMonitors())
	handleServerError(err, w)
}
