package main

import (
	"flag"

	. "logsviewer/pkg/backend"
	"logsviewer/pkg/backend/log"
)

func main() {
	log.Log.Println("Starting logsviewer")

	publicDir := flag.String("public-dir", "./frontend/build/", "directory containing static web assets.")
	insightsBinaryPath := flag.String("insights-binary-path", "", "path to the insights-client binary.")
	flag.Parse()

	log.Log.Printf("public-dir: %s", *publicDir)
	log.Log.Printf("insights-binary-path: %s", *insightsBinaryPath)

	if spawnErr := Spawn(*publicDir, *insightsBinaryPath); spawnErr != nil {
		panic(spawnErr)
	}
}
