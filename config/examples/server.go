package main

import (
	"net/http"
)

func main() {
    fs := http.FileServer(http.Dir("."))
    http.Handle("/", fs)

    println("📡 Serving on http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}
