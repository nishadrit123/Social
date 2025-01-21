package main

import (
	"log"
	"social/internal/env"
	"social/internal/store"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}
	cfg := &config{
		addr: env.GetString("ADDR", ":8080"),
	}
	store := store.NewStorage(nil)
	app := &application{
		config: *cfg,
		store:  store,
	}
	mux := app.mount()

	log.Fatal(app.run(mux))
}
