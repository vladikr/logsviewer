package main

import (
    "net/http"

    . "logsviewer/pkg/backend"
)
func main() {
    SetupRoutes()
    http.ListenAndServe(":8080", nil)
}
