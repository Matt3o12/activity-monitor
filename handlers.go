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

// UptimeCheckerHandle is the basic handle for this webpage. Every
// handle should be a type of it.
type UptimeCheckerHandler func(r *http.Request, p httprouter.Params) Page

// MainMiddleware should be the first middleware. It calls the Page's
// Execution function and call further middlewares in the future.
func MainMiddleware(h UptimeCheckerHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		page := h(r, p)
		if page != nil {
			page.Execute(w)
		}
	}
}

// A Page is returned by every view (i.e. handler)
// and writes the HTML to client or sends a redirect.
type Page interface {
	Execute(http.ResponseWriter) bool
}

// Redirect is a page returned when a redirect should occur
// (instead of a template for example).
// Redirect should know the requests since it needs to determine
// the absolute URL.
type Redirect struct {
	Location string
	Status   int
	Request  *http.Request
}

// Execute redirects the user to the given location.
// Always returns true.
func (r Redirect) Execute(w http.ResponseWriter) bool {
	http.Redirect(w, r.Request, r.Location, r.Status)
	return true
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

func dashboardHandler(r *http.Request, _ httprouter.Params) Page {
	tw := defaultTW.SetTemplate(indexTmpl)
	monitors := []struct {
		Id    int
		Name  string
		Type  string
		Event EventType
	}{}

	err := NewDatabaseError(db.Model(&Monitor{}).Alias("m").
		Column("m.name", "m.type", "m.id", "l1.event").
		Join("JOIN monitor_logs l1 ON (m.id = l1.monitor_id)").
		Join("LEFT OUTER JOIN monitor_logs l2 ON (m.id = l2.monitor_id " +
		"AND l1.date <= l2.date AND l1.id < l2.id)").
		Where("l2.id IS NULL").
		Order("m.id ASC").
		Limit(50).
		Select(&monitors))

	return tw.SetTmplArgs(monitors).SetError(err)
}

func getAddMonitorTemplate(errMsg string) Page {
	data := struct {
		Values []string
		Err    string
	}{SupportedTypes, errMsg}

	return defaultTW.SetTemplate(monitorAddTmpl).SetTmplArgs(data)
}

func addMonitorGetHandler(_ *http.Request, _ httprouter.Params) Page {
	return getAddMonitorTemplate("")
}

func addMonitorPostHandler(r *http.Request, _ httprouter.Params) Page {
	// TODO: actually save the monitor.
	monitor := new(Monitor)
	if err := decodeForm(monitor, r); err != nil {
		// TODO: better input validat (which fields were invalid).
		return getAddMonitorTemplate("Form data invaild. Please check input")
	}

	log.Printf("Created (mock) a new monitor: %q", monitor)
	// TODO: redirect to newly created URL.
	return Redirect{Location: "/", Request: r, Status: http.StatusSeeOther}
}

func viewMonitorHandler(r *http.Request, params httprouter.Params) Page {
	// TODO: fetch from real database.
	monitors := makeMonitors()
	tw := defaultTW.SetTemplate(monitorViewTmpl)
	name := params.ByName("id")
	monitorNotFoundErr := StatusError{
		Status: 404, Message: "Monitor could not be found",
	}

	if id, err := strconv.Atoi(name); err != nil || id > len(monitors) {
		fmt.Println(id)
		return tw.SetError(monitorNotFoundErr)
	} else {
		return tw.SetTmplArgs(monitors[id])
	}

}
