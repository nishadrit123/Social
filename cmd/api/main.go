package main

import (
	"log"
	"social/internal/env"

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
	app := &application{
		config: *cfg,
	}
	mux := app.mount()

	log.Fatal(app.run(mux))
}
