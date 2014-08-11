package goji_middleware

import (
	"net/http"
)

// A version of http.ResponseWriter that lets you see the status code writtern to the response
type StatusTrackingResponseWriter struct {
	http.ResponseWriter
	// http status code written
	Status int
}

func (w *StatusTrackingResponseWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}
