package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type config struct {
	addr string
}

type application struct {
	config config
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)
	})
	return r
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Millisecond,
	}
	log.Printf("Started server on %v\n", app.config.addr)
	err := srv.ListenAndServe()
	log.Printf("ListenAndServe is a blocking call and wint be executed unless it throws ant err %v", err)
	if err != nil {
		return err
	}
	return nil
}
