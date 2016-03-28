package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/schema"
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

var formDecoder = schema.NewDecoder()

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
	log.Println("An unexpected error 500 occured:", org)
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

func decodeForm(i interface{}, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	return formDecoder.Decode(i, r.Form)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) error {
	t, err := template.ParseFiles("templates/index.html", "templates/layout.html")
	if err != nil {
		return err
	}

	return t.Execute(w, makeMonitors())
}

func handleAddMonitor(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return writeAddMonitorTemplate(w, "")

	case "POST":
		return handleAddMonitorPost(w, r)

	default:
		return errors.New("Method not supported")
	}
}

func writeAddMonitorTemplate(w http.ResponseWriter, errMsg string) error {
	t, err := template.ParseFiles("templates/monitors/add.html", "templates/layout.html")
	if err != nil {
		return err
	}

	data := struct {
		Values []string
		Err    string
	}{SupportedTypes, errMsg}

	return t.Execute(w, data)
}

func handleAddMonitorPost(w http.ResponseWriter, r *http.Request) error {
	// TODO: actually save the monitor.
	monitor := new(Monitor)
	if err := decodeForm(monitor, r); err != nil {
		// TODO: better input validation (which fields were invalid).
		writeAddMonitorTemplate(w, "Form data invaild. Please check input")
		return nil
	}

	log.Printf("Created (mock) a new monitor: %q", monitor)
	// TODO: redirect to newly created URL.
	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}
