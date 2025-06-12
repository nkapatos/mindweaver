package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/nkapatos/mindweaver/internal/templates/pages"
)

func main() {
	component := pages.Home()

	http.Handle("/", templ.Handler(component))

	fs := http.FileServer(http.Dir("dist"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
