package main

import (
	"log"
	"os"

	"code/cmd/api/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
