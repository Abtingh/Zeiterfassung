package main

import (
	"log"
	"net/http"

	"github.com/Abtingh/Zeiterfassung/internal/routes"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./public/public-static")))

	// ثبت مسیرها از پکیج handlers
	routes.Register(mux)

	log.Println("Starting Server on Port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
