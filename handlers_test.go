package main

import (
	"errors"
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
