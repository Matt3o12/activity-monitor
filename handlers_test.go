package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func shutupLog() func() {
	log.SetOutput(ioutil.Discard)
	return func() {
		log.SetOutput(os.Stderr)
	}
}

func TestIndexHandler(t *testing.T) {
	data := []struct {
		prefix string
		path   string
		found  bool
	}{
		{"/", "/", true},
		{"/test/", "/test/", true},
		{"/", "/test", false},
		{"/", "/test.html", false},
		{"/foo/", "/foo/bar.html", false},
		{"/foo/bar/", "/foo/bar/", true},
		{"/foo/bar/", "/foo/bar/home", false},
		{"/foo/bar/", "/foo/bar/home/", false},
		{"/foo/bar/", "/foo/bar/home.test", false},
	}

	var (
		expectedBodyCalled    = "called"
		expectedBodyNotCalled = "404 page not found\n"
	)

	for _, row := range data {
		url := "http://server.local" + row.path
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("Could not construct request. Got: %v", err)
		}

		recorder := httptest.NewRecorder()
		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.Write([]byte("called"))
		}

		IndexHandler(row.prefix, handler)(recorder, request)

		cond := fmt.Sprintf("for %q, prefix %q", row.path, row.prefix)
		if row.found {
			if c := recorder.Code; c != http.StatusOK {
				t.Errorf("Unexpected status code: '%v' %v.", c, cond)
			} else if !handlerCalled {
				t.Errorf("Handler not called %v.", cond)
			} else if b := recorder.Body.String(); b != expectedBodyCalled {
				t.Errorf("Invalid body: '%v' %v.", b, cond)
			}
		} else {
			if c := recorder.Code; c != http.StatusNotFound {
				t.Errorf("Unexpected status code: %v %v", c, cond)
			} else if handlerCalled {
				t.Errorf("Handler was unexpectedly called %v.", cond)
			} else if b := recorder.Body.String(); b != expectedBodyNotCalled {
				t.Errorf("Body Unexpected body: %q %v", b, cond)
			}
		}
	}
}

func TestMainHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(200)
		w.Write([]byte("called"))
		return nil
	}

	request, err := http.NewRequest("GET", "http://server.local/", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	MainHandler(handler)(recorder, request)
	if c := recorder.Code; c != http.StatusOK {
		t.Errorf("Recorded unexpected code: %v (expected: 200)", c)
	}

	if b := recorder.Body.String(); b != "called" {
		t.Errorf("Recorded unkown body: %q.", b)
	}
}

func TestMainHandlerErr(t *testing.T) {
	defer shutupLog()()
	expectedError := errors.New("unittest error (expected)")
	recorder := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) error {
		return expectedError
	}

	request, err := http.NewRequest("GET", "http://server.local/", nil)
	if err != nil {
		t.Fatalf("Could not construct request: %v", err)
	}

	MainHandler(handler)(recorder, request)
	if c := recorder.Code; c != http.StatusInternalServerError {
		t.Errorf("Expected to record server error (500), got: %v", c)
	}

	if c := recorder.HeaderMap.Get("Content-Type"); c != htmlContent {
		t.Errorf("Expected to get HTMLContent, got: %v.", c)
	}

	if b := recorder.Body.String(); !strings.Contains(b, expectedError.Error()) {
		t.Errorf("Body: %q does not contain error: '%v'.", b, expectedError)
	}
}

func TestHTTPError_Error(t *testing.T) {
	err := HTTPError{123, "foo bar"}
	if err.Error() != "123 -- foo bar" {
		msg := "Error: %q did not formate as expected. Got: %q."
		t.Errorf(msg, err, err.Error())
	}
}

func TestHTTPError_WriteToPage(t *testing.T) {
	err := HTTPError{501, "This is a test error"}
	recorder := httptest.NewRecorder()
	err.WriteToPage(recorder)

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
