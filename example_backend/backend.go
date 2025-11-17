package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9001"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("Hello from backend on port %s\n", port)
		_, _ = w.Write([]byte(msg))
	})

	log.Printf("Backend started on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
