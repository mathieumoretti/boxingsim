package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status": "OK"}`)
	})

	fmt.Println("Boxing API Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}