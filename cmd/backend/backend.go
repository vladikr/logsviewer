package main

import (
    "net/http"
    "flag"
    "os"

    . "logsviewer/pkg/backend"
    "logsviewer/pkg/backend/log"
)
func main() {
    fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
    fs.SetOutput(os.Stdout)
    log.Log.Println("Starting logsviewer")
    publicDir := fs.String("public-dir", "./frontend/build/", "directory containing static web assets.")
    mux := SetupRoutes(*publicDir)
    http.ListenAndServe(":8080", mux)
}
