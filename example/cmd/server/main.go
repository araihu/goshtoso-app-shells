package main

import (
	"log"
	"net/http"

	"github.com/araihu/goshtoso-app-shells/example/internal/server"
)

func main() {
	const address = ":8092"
	log.Printf("catalog shell example listening on http://localhost%s", address)
	log.Fatal(http.ListenAndServe(address, server.New()))
}
