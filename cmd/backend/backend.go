package main

import (
    "net/http"
    "flag"
    "fmt"
    "os"

    . "logsviewer/pkg/backend"
)
func main() {
    fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
    fs.SetOutput(os.Stdout)
    publicDir := fs.String("public-dir", "./frontend/build/", "directory containing static web assets.")
    mux := SetupRoutes(*publicDir)
    http.ListenAndServe(":8080", mux)
}
