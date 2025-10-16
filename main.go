package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World")
	})

	fmt.Println("Server running at http://localhost:9001")
	err := http.ListenAndServe(":9001", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

