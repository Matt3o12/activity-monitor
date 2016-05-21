package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/schema"
	"github.com/julienschmidt/httprouter"
)

const (
	htmlContent = "text/html; charset=utf-8"
)

var (
	err500TemplateNotExecuting = []byte(`<!DOCTYPE html><html>
<head><title>Error 500</title></head>
<body><h1>Error 500</h1><p>Could not execute template.</p></body></html>`)

	err404 = StatusError{Status: 404, Message: "Page could not be found"}

	formDecoder = schema.NewDecoder()
)

// All template related variables.
var (
	defaultTW = TemplateWriter{ServerErrorHandler: handleServerError}

	errorTmpl    = MustTemplate(NewBareboneTemplate("error.html"))
	error500Tmpl = MustTemplate(NewBareboneTemplate("error500.html"))

	indexTmpl       = MustTemplate(NewTemplate("index.html"))
	monitorViewTmpl = MustTemplate(NewTemplate("monitors/view.html"))
	monitorAddTmpl  = MustTemplate(NewTemplate("monitors/add.html"))
)

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

func dashboardHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tw := defaultTW.Configure(indexTmpl, w)
	tw.SetTmplArgs(makeMonitors()).Execute()
}

func writeAddMonitorTemplate(w http.ResponseWriter, errMsg string) {
	data := struct {
		Values []string
		Err    string
	}{SupportedTypes, errMsg}
	defaultTW.Configure(monitorAddTmpl, w).SetTmplArgs(data).Execute()
}

func addMonitorGetHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	writeAddMonitorTemplate(w, "")
}

func addMonitorPostHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

func viewMonitorHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// TODO: fetch from real database.
	monitors := makeMonitors()
	tw := defaultTW.Configure(monitorViewTmpl, w)
	name := params.ByName("id")
	monitorNotFoundErr := StatusError{
		Status: 404, Message: "Monitor could not be found",
	}

	fmt.Println(name, len(monitors))
	if id, err := strconv.Atoi(name); err != nil || id > len(monitors) {
		fmt.Println(id)
		tw = tw.SetError(monitorNotFoundErr)
	} else {
		fmt.Println(id, monitors[id])
		tw = tw.SetTmplArgs(monitors[id])
	}
	tw.Execute()
}
