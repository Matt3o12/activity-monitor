package main

import (
	"fmt"
	"net/http"
)

type HTTPError interface {
	WriteToPage(w http.ResponseWriter) bool
	HTTPStatus() int
	Error() string
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

type DatabaseError struct {
	dbError error
}

func NewDatabaseError(err error) HTTPError {
	if err == nil {
		return nil
	}

	return &DatabaseError{err}
}

func (e *DatabaseError) HTTPStatus() int {
	return 500
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("DatabaseError: %v", e.dbError)
}

func (e *DatabaseError) WriteToPage(w http.ResponseWriter) bool {
	tw := defaultTW.Configure(error500Tmpl, w)

	return tw.SetStatusCode(500).SetTmplArgs(e).Execute()
}
