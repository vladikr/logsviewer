package main

import (
    //"net/http"
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
    publicDirAlt := fs.String("public-dir-alt", "./frontend-alt/build/", "directory containing static web assets.")
    if err := Spawn(*publicDir, *publicDirAlt); err != nil {
        panic(err)
    }
    
}
