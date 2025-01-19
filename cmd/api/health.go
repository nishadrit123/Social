package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	s := "All OK 200"
	w.Write([]byte(s))
}
