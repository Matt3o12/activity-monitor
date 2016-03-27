package main

import (
	"html/template"
	"log"
	"net/http"
)

const (
	htmlContent = "text/html; charset=utf-8"
)

var (
	err500TemplateNotLoading = []byte(`<!DOCTYPE html><html>
<head><title>Error 500</title></head>
<body><h1>Error 500</h1><p>Could not load error template</p></body></html>`)

	err500TemplateNotExecuting = []byte(`<!DOCTYPE html><html>
<head><title>Error 500</title></head>
<body><h1>Error 500</h1><p>Could not execute template.</p></body></html>`)
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func MainHandler(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			handleServerError(err, w)
		}
	}
}

func IndexHandler(prefix string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == prefix {
			h(w, r)
		} else {
			http.NotFound(w, r)
		}
	}
}

func handleServerError(org error, w http.ResponseWriter) {
	log.Println("An unexpected error 500 occured: ", org)
	w.Header().Set("Content-Type", htmlContent)
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

func dashboardHandler(w http.ResponseWriter, r *http.Request) error {
	t, err := template.ParseFiles("templates/index.html", "templates/layout.html")
	if err != nil {
		return err
	}

	return t.Execute(w, makeMonitors())
}
