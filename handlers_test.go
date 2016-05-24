package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

const ContentTypeURLEncoded = "application/x-www-form-urlencoded"

func shutupLog() func() {
	log.SetOutput(ioutil.Discard)
	return func() {
		log.SetOutput(os.Stderr)
	}
}
func newRequest(t *testing.T, method, url string) *http.Request {
	if req, err := http.NewRequest(method, url, nil); err != nil {
		t.Fatalf("Could not construct request: %q.", err)
		return nil
	} else {
		return req
	}
}

func TestStatusError_Error(t *testing.T) {
	err := StatusError{123, "foo bar"}
	if err.Error() != "123 -- foo bar" {
		msg := "Error: %q did not formate as expected. Got: %q."
		t.Errorf(msg, err, err.Error())
	}
}

func TestStatusError_WriteToPage(t *testing.T) {
	err := StatusError{501, "This is a test error"}
	recorder := httptest.NewRecorder()
	err.WriteToPage(recorder)

	if c := recorder.Code; c != 501 {
		t.Errorf("Recorded unexpected error code: %v. Expected 501", c)
	}

	b := strings.TrimSpace(recorder.Body.String())
	if !strings.Contains(b, "<title>Error 501 – This is a test error</title>") {
		t.Errorf("Error message and code not in <title>:\n%v\n\n", b)
	}

	if !strings.Contains(b, "<h1>Error 501</h1>") {
		t.Errorf("Error code not found in heading <h1>:\n%v\n\n", b)
	}

	if !strings.Contains(b, "This is a test error<br><br>") {
		t.Errorf("Error message not found in content before <br>", b)
	}
}

func TestHandleServerError(t *testing.T) {
	defer shutupLog()()
	recorder := httptest.NewRecorder()
	err := errors.New("Test Server Error")
	handleServerError(err, recorder)

	if recorder.Code != 500 {
		msg := "Expected to get error code 500, got instead: %v"
		t.Errorf(msg, recorder.Code)
	}

	expectedParts := []string{
		"<h1>Error 500</h1>",
		"<p>While executing this request, an unexpected error occured. We are really sorry about that</p>",
		"<code>Test Server Error</code>",
	}

	body := recorder.Body.String()
	t.Logf("Response Body: %q", body)
	for _, msg := range expectedParts {
		if !strings.Contains(body, msg) {
			t.Errorf("Part '%q' not found in body.", msg)
		}
	}

	expected := "text/html; charset=utf-8"
	if tp := recorder.HeaderMap.Get("Content-Type"); tp != expected {
		t.Errorf("Recorded wrong Content Type: %v", tp)
	}
}

func TestRedirect(t *testing.T) {
	testcase := []struct {
		Location string
		Status   int
	}{
		{"/test.html", 301},
		{"/index.html", 302},
		{"/foo/bar/", 303},
	}

	for _, row := range testcase {
		recorder := httptest.NewRecorder()
		request, err := http.NewRequest("GET", "http://localhost/post", nil)
		if err != nil {
			t.Fatalf("Could not construct request: %v", err)
		}
		redirect := Redirect{
			Location: row.Location,
			Status:   row.Status,
			Request:  request,
		}
		redirect.Execute(recorder)

		if l := recorder.HeaderMap.Get("Location"); l != row.Location {
			t.Errorf("Location: %v, wanted: %v", l, row.Location)
		}

		if recorder.Code != row.Status {
			msg := "Status code: %v, wanted: %v"
			t.Errorf(msg, recorder.Code, row.Status)
		}
	}
}

func getTemplateWriter(t *testing.T, p Page) TemplateWriter {
	switch p.(type) {
	case TemplateWriter:
		return p.(TemplateWriter)

	case *TemplateWriter:
		return *p.(*TemplateWriter)

	default:
		t.Fatalf("Not a template writer: %v", p)
		return TemplateWriter{}
	}
}

func MarkLong(t *testing.T) {
	if testing.Short() {
		t.Skip("Skiping test -- marked as long-running")
	}
}

func TestDashboardHandler(t *testing.T) {
	defer InitTestConnection(t)()
	tw := getTemplateWriter(t, dashboardHandler(nil, nil))
	if tw.Err != nil {
		t.Errorf("dashboardHandler returned an error: %v", tw.Err)
	}

	monitors := tw.TmplArgs.([]struct {
		Id    int
		Name  string
		Type  string
		Event EventType
	})

	expected := []struct {
		Id    int
		Name  string
		Type  string
		Event EventType
	}{
		{1, "TCP/UDP Socket", "socket", MonitorUpEvent},
		{2, "HTTP(s) Server", "http", MonitorUpEvent},
		{3, "Main Server", "ping", MonitorUpEvent},
		{4, "Down server", "ping", MonitorDownEvent},
	}

	if len(expected) != len(monitors) {
		msg := "Wanted %v monitors; got: %v"
		t.Errorf(msg, len(expected), len(monitors))
	}

	for i, r := range expected {
		m := monitors[i]
		if r.Id != m.Id || r.Name != m.Name || r.Type !=
			m.Type || r.Event != m.Event {
			t.Errorf("%#v != %#v", r, m)
		}
	}
}

func assertAddMonitorTemplate(t *testing.T, err string, o interface{}) {

	data := o.(struct {
		Values []string
		Err    string
	})
	if data.Err != err {
		t.Errorf("Template shows a validation error: %v", data.Err)
	}

	diff := false
	for i, r := range SupportedTypes {
		if r != data.Values[i] {
			diff = true
		}
	}

	if len(data.Values) != len(SupportedTypes) || diff {
		t.Errorf("%#v != %#v", SupportedTypes, data.Values)
	}
}

func TestAddMonitorGetHandler(t *testing.T) {
	tw := getTemplateWriter(t, addMonitorGetHandler(nil, nil))

	if tw.Err != nil {
		t.Errorf("addMonitorGetHandler returned an error: %v", tw.Err)
	}

	assertAddMonitorTemplate(t, "", tw.TmplArgs)
}

func assertAddMonitorErrMsg(t *testing.T, tw TemplateWriter, expected string) {
	err := tw.TmplArgs.(struct {
		Values []string
		Err    string
	}).Err
	if err != expected {
		msg := "Wanted template error message to be: %q, got: %q"
		t.Errorf(msg, expected, err)
	}
}

func TestAddMonitorPostHandlerErrorNoName(t *testing.T) {
	// Setup the database just in case the form is "valid"
	// and the server tries to write to it.
	defer InitTestConnection(t)()

	data := url.Values{}
	data.Set("name", "")
	data.Set("type", "ping")
	data.Set("paused", "on")

	body := strings.NewReader(data.Encode())
	r, err := http.NewRequest("POST", "", body)
	if err != nil {
		t.Fatalf("Could not construct request: %v", err)
	}
	r.Header.Set("Content-Type", ContentTypeURLEncoded)

	tw := getTemplateWriter(t, addMonitorPostHandler(r, nil))
	assertAddMonitorErrMsg(t, tw, "A name for the monitor is required.")
}

func TestAddMonitorPostHandlerErrorInvalidForm(t *testing.T) {
	defer InitTestConnection(t)()
	r, err := http.NewRequest("POST", "", strings.NewReader("%"))
	r.Header.Set("Content-Type", ContentTypeURLEncoded)
	if err != nil {
		t.Fatalf("Could not construct request: %v", r)
	}
	tw := getTemplateWriter(t, addMonitorPostHandler(r, nil))
	assertAddMonitorErrMsg(t, tw, "Form data invaild. Please check input.")
}

func assertMonitor(t *testing.T, m Monitor, id int, n, mType string) {
	if m.Name != n {
		t.Errorf("Wanted montior %v to be: %q, got: %q", id, n, m.Name)
	}

	if m.Id != id {
		t.Errorf("Wanted Monitor %q's id to be: %v, got: %v", n, id, m.Id)
	}
	if m.Type != mType {
		t.Errorf("Wanted Monitor %q's type to be: %q, got: %q", n, mType, m.Type)
	}
}

func TestAddMonitorPostHandler(t *testing.T) {
	type testcase struct {
		name   string
		mType  string
		paused bool
	}

	paused2str := func(paused bool) string {
		if paused {
			return "on"
		}

		return ""
	}

	const mId = 5
	test := func(row testcase) {
		defer InitTestConnection(t)()
		form := url.Values{}
		form.Set("name", row.name)
		form.Set("type", row.mType)
		form.Set("paused", paused2str(row.paused))

		r, err := http.NewRequest("POST", "", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", ContentTypeURLEncoded)
		if err != nil {
			t.Errorf("Could not construct request: %v", err)
			return
		}

		page := addMonitorPostHandler(r, nil)
		redirect, ok := page.(Redirect)
		if !ok {
			t.Errorf("Wanted handler to return Redirect, got: %T", page)
			return
		}

		location := fmt.Sprintf("/monitors/view/%v/", mId)
		if redirect.Location != location {
			msg := "Wanted redirect to point to %q, got: %q."
			t.Errorf(msg, location, redirect.Location)
		}

		if s := redirect.Status; s != http.StatusSeeOther {
			t.Errorf("Wanted redirect status to be SeeOther, got: %v", s)
		}

		monitor := Monitor{Id: mId}
		err = db.Select(&monitor)
		if err != nil {
			t.Errorf("Could not fetch monitor from database: %v", err)
			return
		}

		assertMonitor(t, monitor, mId, row.name, row.mType)
		logs := []MonitorLog{}
		err = db.Model(&logs).Where("monitor_id = ?", mId).Select()
		if err != nil {
			t.Errorf("Could not fetch Monitor %q's logs: %v", row.name, err)
			return
		}

		if len(logs) != 2 {
			msg := "Montior %q's logs are too long/short. Wanted 2, got: %v"
			t.Errorf(msg, row.name, len(logs))
			return
		}

		if logs[0].Event != MonitorCreatedEvent {
			msg := "Wanted %q's first event to be Monitor " +
				"Created Event, got: %v"
			t.Errorf(msg, row.name, logs[0].Event)
		}

		second := MonitorStartedEvent
		if row.paused {
			second = MonitorPausedEvent
		}

		if logs[1].Event != second {
			msg := "Wanted %q's second event to be %v, got: %v"
			t.Errorf(msg, row.name, second, logs[1].Event)
		}
	}

	cases := []testcase{
		{"foo", "hello", true},
		{"bar", "world", true},
		{"started", "ping", false},
	}

	for _, row := range cases {
		test(row)
	}
}

func TestViewMonitorHandler(t *testing.T) {
	defer InitTestConnection(t)()
	type testcase struct {
		id    int
		name  string
		mType string
	}

	cases := []testcase{
		{1, "TCP/UDP Socket", "socket"},
		{2, "HTTP(s) Server", "http"},
		{3, "Main Server", "ping"},
		{4, "Down server", "ping"},
	}

	for _, row := range cases {
		params := httprouter.Params{{Key: "id", Value: strconv.Itoa(row.id)}}
		tw := getTemplateWriter(t, viewMonitorHandler(nil, params))
		if tw.Err != nil {
			t.Errorf("TemplateWriter contains an error: %v", tw.Err)
		} else {
			monitor := tw.TmplArgs.(Monitor)
			assertMonitor(t, monitor, row.id, row.name, row.mType)
		}
	}
}

func TestViewMonitorHandlerNotFound(t *testing.T) {
	defer InitTestConnection(t)()
	testcase := []string{"5", "abc", "foo", "Bar", "100", "-1", "6"}
	for _, id := range testcase {
		param := httprouter.Params{httprouter.Param{Key: "id", Value: id}}
		tw := getTemplateWriter(t, viewMonitorHandler(nil, param))
		if tw.Err == nil {
			msg := "Expceted TemplateWriter to contain " +
				"an error for id: %q. Got nil"
			t.Errorf(msg, id)
			continue
		}

		err, ok := tw.Err.(StatusError)
		if !ok {
			msg := "Expected TemplateWriter to contain a " +
				"StatusError for id: %q, got: %v"
			t.Errorf(msg, id, tw.Err)
			continue
		}

		if err.HTTPStatus() != http.StatusNotFound {
			msg := "Expected status code to be Not Found for id: %q, got: %v"
			t.Errorf(msg, id, err.HTTPStatus)
		}

		expected := "Monitor could not be found"
		if err.Message != expected {
			msg := "Expected message to say: %q for id: %q, got: %q"
			t.Errorf(msg, expected, id, err.Message)
		}
	}
}
