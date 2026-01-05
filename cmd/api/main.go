package main

import (
	"log"

	"code/internal/web"
)

const port = ":8080"

func main() {
	r := web.NewRouter()

	err := r.Run(port)
	if err != nil {
		log.Fatal(err)
	}
}
