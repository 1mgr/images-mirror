package main

import (
	"encoding/json"
	"net/http"
)

func httpError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Write([]byte(msg))
}

func writeLine(w http.ResponseWriter, line string) {
	w.Write([]byte(line + "\n"))
	w.(http.Flusher).Flush()
}

func httpOk(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

type responseStatusWriter struct {
	w http.ResponseWriter
}

func (w *responseStatusWriter) Write(status string) {
	writeLine(w.w, status)
}

func makeStatusWriter(w http.ResponseWriter) StatusWriter {
	return &responseStatusWriter{w: w}
}
