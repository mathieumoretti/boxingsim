package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Serve static files from the web directory
	fs := http.FileServer(http.Dir("web/"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If the path is a file, serve it directly
		path := filepath.Join("web", r.URL.Path)

		// Check if the requested file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// If file doesn't exist, serve index.html (for SPA routing)
			http.ServeFile(w, r, "web/index.html")
			return
		}

		// Otherwise, serve the file directly
		fs.ServeHTTP(w, r)
	})

	fmt.Println("Starting web server on :8081")
	fmt.Println("Visit http://localhost:8081 to view the boxing simulator UI")

	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}