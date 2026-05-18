// Command kranix-mock-api runs an in-memory fake kranix-api for local integration tests.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	skipAuth := flag.Bool("skip-auth", true, "do not require Authorization header (recommended for unit tests)")
	flag.Parse()

	if v := os.Getenv("KRANIX_MOCK_ADDR"); v != "" {
		*addr = v
	}
	if os.Getenv("KRANIX_MOCK_REQUIRE_AUTH") == "1" {
		*skipAuth = false
	}

	srv := newMockServer(*skipAuth)
	log.Printf("kranix-mock-api listening on %s (skip-auth=%v)", *addr, *skipAuth)
	if err := http.ListenAndServe(*addr, srv); err != nil {
		log.Fatal(err)
	}
}
