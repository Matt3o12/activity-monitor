package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/schema"
)

const (
	htmlContent = "text/html; charset=utf-8"
)

var (
	err500TemplateNotExecuting = []byte(`<!DOCTYPE html><html>
<head><title>Error 500</title></head>
<body><h1>Error 500</h1><p>Could not execute template.</p></body></html>`)

	err404 = HTTPError{Status: 404, Message: "Page could not be found"}

	formDecoder = schema.NewDecoder()
)

// All template related variables.
var (
	defaultTW = TemplateWriter{ServerErrorHandler: handleServerError}

	errorTmpl    = MustTemplate(NewBareboneTemplate("error.html"))
	error500Tmpl = MustTemplate(NewBareboneTemplate("error500.html"))

	indexTmpl      = MustTemplate(NewTemplate("index.html"))
	monitorAddTmpl = MustTemplate(NewTemplate("monitors/add.html"))
)

type HTTPError struct {
	Status  int
	Message string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("%v -- %v", e.Status, e.Message)
}

func (e HTTPError) WriteToPage(w http.ResponseWriter) bool {
	tw := defaultTW.Configure(errorTmpl, w)

	// We can't use SetError since that would create an infinity loop.
	return tw.SetStatusCode(e.Status).SetTmplArgs(e).Execute()
}

func IndexHandler(prefix string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == prefix {
			h(w, r)
		} else {
			defaultTW.Configure(nil, w).SetError(err404).Execute()
		}
	}
}

func handleServerError(org error, w http.ResponseWriter) {
	log.Println("An unexpected error 500 occured:", org)
	w.Header().Set("Content-Type", htmlContent)
	w.WriteHeader(500)
	if err := error500Tmpl.Execute(w, org.Error()); err != nil {
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

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	tw := defaultTW.Configure(indexTmpl, w)
	tw.SetTmplArgs(makeMonitors()).Execute()
}

func handleAddMonitor(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		writeAddMonitorTemplate(w, "")

	case "POST":
		handleAddMonitorPost(w, r)

	default:
		tw := defaultTW.Configure(nil, w)
		tw.SetError(HTTPError{405, "Method not allowed"}).Execute()
	}
}

func writeAddMonitorTemplate(w http.ResponseWriter, errMsg string) {
	data := struct {
		Values []string
		Err    string
	}{SupportedTypes, errMsg}
	defaultTW.Configure(monitorAddTmpl, w).SetTmplArgs(data).Execute()
}

func handleAddMonitorPost(w http.ResponseWriter, r *http.Request) {
	// TODO: actually save the monitor.
	monitor := new(Monitor)
	if err := decodeForm(monitor, r); err != nil {
		// TODO: better input validation (which fields were invalid).
		writeAddMonitorTemplate(w, "Form data invaild. Please check input")
		return
	}

	log.Printf("Created (mock) a new monitor: %q", monitor)
	// TODO: redirect to newly created URL.
	http.Redirect(w, r, "/", http.StatusFound)
}
