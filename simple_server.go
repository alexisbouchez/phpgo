package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := "8082"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Simple PHPGo Server - Port %s", port)
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Test endpoint working!")
	})

	fmt.Printf("Simple server running on http://localhost:%s\n", port)
	fmt.Println("Press Ctrl+C to stop")

	http.ListenAndServe(":"+port, nil)
}