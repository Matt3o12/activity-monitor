package main

import (
	"fmt"
	"net/http"
)

// HTTPError is a gerneric error type that can be
// shown on a page.
type HTTPError interface {
	WriteToPage(w http.ResponseWriter) bool
	HTTPStatus() int
	Error() string
}

// StatusError is a simple error for things such as
// Page not found.
type StatusError struct {
	Status  int
	Message string
}

// Error returns the error message along with the status
// as a readable string.
func (e StatusError) Error() string {
	return fmt.Sprintf("%v -- %v", e.Status, e.Message)
}

// Returns the HTTP Status that should be shown on
// the page.
func (e StatusError) HTTPStatus() int {
	return e.Status
}

// WriteToPage writes the error message and the status to the page.
func (e StatusError) WriteToPage(w http.ResponseWriter) bool {
	tw := defaultTW.SetTemplate(errorTmpl)

	// We can't use SetError since that would create an infinity loop.
	return tw.SetStatusCode(e.Status).SetTmplArgs(e).Execute(w)
}

// DatabaseError is a wrapper for any pg database error
// so that it can be shown on the page.
type DatabaseError struct {
	dbError error
}

// Creates a new database error from any conventinal error.
// While it err is not enforced to be a database-related
// error, it is recomended to only use this function
// from errors that have been caused by interacting
// with the database directly.
// Also, if err is nil, a nil interface will be returned.
func NewDatabaseError(err error) HTTPError {
	if err == nil {
		return nil
	}

	return &DatabaseError{err}
}

// HTTPStatus returns 500 for server error.
func (e *DatabaseError) HTTPStatus() int {
	return 500
}

// Error returns the actual error message.
func (e *DatabaseError) Error() string {
	return fmt.Sprintf("DatabaseError: %v", e.dbError)
}

// Writes the error message to the page.
func (e *DatabaseError) WriteToPage(w http.ResponseWriter) bool {
	tw := defaultTW.SetTemplate(error500Tmpl)

	return tw.SetStatusCode(500).SetTmplArgs(e).Execute(w)
}
