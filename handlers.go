package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gopkg.in/pg.v4"

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

func dashboardHandler(_ *http.Request, _ httprouter.Params) Page {
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
	if err := r.ParseForm(); err != nil {
		msg := "Form data invaild. Please check input."
		return getAddMonitorTemplate(msg)
	}

	monitor := Monitor{}
	monitor.Name = strings.TrimSpace(r.PostFormValue("name"))
	if monitor.Name == "" {
		return getAddMonitorTemplate("A name for the monitor is required.")
	}

	monitor.Type = r.PostFormValue("type") // TODO: Actually parse type
	tx, err := db.Begin()
	defer tx.Rollback()

	errHandler := TransactionErrorHandler{}
	errHandler.Err(err)
	if err == nil {
		errHandler.Err(tx.Create(&monitor))
		createdEvent := MonitorLog{
			Event:     MonitorCreatedEvent,
			Date:      time.Now(),
			MonitorId: monitor.Id,
		}

		errHandler.Err(tx.Create(&createdEvent))
		secondEvent := MonitorLog{
			Event:     MonitorStartedEvent,
			Date:      time.Now(),
			MonitorId: monitor.Id,
		}
		if r.PostFormValue("paused") == "on" {
			secondEvent.Event = MonitorPausedEvent
		}

		errHandler.Err(tx.Create(&secondEvent)).Err(tx.Commit())
	}

	if errHandler.FirstErr() == nil {
		tx.Commit()
		return Redirect{
			Location: fmt.Sprintf("/monitors/view/%d/", monitor.Id),
			Request:  r, Status: http.StatusSeeOther,
		}
	} else {
		return defaultTW.SetError(errHandler.FirstErr())
	}
}

func viewMonitorHandler(_ *http.Request, params httprouter.Params) Page {
	id, err := strconv.Atoi(params.ByName("id"))
	notFoundErr := defaultTW.SetError(StatusError{
		Status:  http.StatusNotFound,
		Message: "Monitor could not be found",
	})

	if err != nil {
		return notFoundErr
	}

	tx, err := db.Begin()
	dt := TransactionErrorHandler{}
	if err != nil {
		return defaultTW.SetError(dt.Err(err).FirstErr())
	}
	defer tx.Rollback()

	monitor := Monitor{Id: id}
	if err := tx.Select(&monitor); err == pg.ErrNoRows {
		return notFoundErr
	} else {
		dt.Err(err)
	}

	dt.Err(tx.Model(&monitor.Logs).Where("monitor_id=?", id).
		Limit(50).Order("date DESC").Select())

	if dt.FirstErr() != nil {
		return defaultTW.SetError(dt.FirstErr())
	}

	return defaultTW.SetTmplArgs(monitor).SetTemplate(monitorViewTmpl)
}
