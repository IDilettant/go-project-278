package main

import (
	"log"
	"os"

	"code/cmd/api/app"
)

// main duplicates cmd/api entrypoint so `go run .` works in CI containers.
func main() {
	if err := app.Run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
