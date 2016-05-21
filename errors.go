package main

import (
	"fmt"
	"net/http"
)

type HTTPError interface {
	WriteToPage(w http.ResponseWriter) bool
	HTTPStatus() int
}

type StatusError struct {
	Status  int
	Message string
}

func (e StatusError) Error() string {
	return fmt.Sprintf("%v -- %v", e.Status, e.Message)
}

func (e StatusError) HTTPStatus() int {
	return e.Status
}

func (e StatusError) WriteToPage(w http.ResponseWriter) bool {
	tw := defaultTW.Configure(errorTmpl, w)

	// We can't use SetError since that would create an infinity loop.
	return tw.SetStatusCode(e.Status).SetTmplArgs(e).Execute()
}
