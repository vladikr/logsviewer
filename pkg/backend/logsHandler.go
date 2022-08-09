package backend

import (
    "path/filepath"
    "archive/tar"
    "compress/gzip"
    "io"
    "log"
    "net/http"
    "os"
    "strings"
)

func handleTarGz(w http.ResponseWriter, srcFile string, targetPath string) {
	if err := unTarGz(w, srcFile, targetPath); err == nil {
	
	    // delete source file
            err = os.Remove(srcFile)
	    if err != nil {
            	log.Fatalln("failed to delete file ", srcFile, " - ", err)
                
	    }
	}
}

func unTarGz(w http.ResponseWriter, srcFile string, targetPath string) error {
    gzipStream, err := os.Open(srcFile)
    defer gzipStream.Close()

    if err != nil {
        log.Fatalln("failed to open file ", srcFile, " - ", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return err
    }

    pathToNamespacesDir := ""
    uncompressedStream, err := gzip.NewReader(gzipStream)
    if err != nil {
        log.Fatalln("failed create gzip stream  - ", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return err
    }

    tarReader := tar.NewReader(uncompressedStream)

    for true {
        header, err := tarReader.Next()

        if err == io.EOF {
            break
        }

        if err != nil {
            log.Fatalln("failed to get next file in tar  - ", err)
	    http.Error(w, err.Error(), http.StatusInternalServerError)
	    return err
        }

	// extract only the namespaces dir
	if !strings.Contains(header.Name, "namespaces/") {
		continue
	}
	
	if pathToNamespacesDir == "" {
	    // find path to the namespaces directory	
	    sp := strings.Split(header.Name, "/")
	    for _, ps := range sp {
	        if ps == "namespaces" {
	            break
	        }
	        pathToNamespacesDir = filepath.Join(pathToNamespacesDir, ps)
	    }
	}
	workingHeaderName := strings.TrimPrefix(header.Name, pathToNamespacesDir)
	newTarget := filepath.Join(targetPath, workingHeaderName)

	switch header.Typeflag {
        case tar.TypeDir:
            if err := os.MkdirAll(newTarget, 0755); err != nil {
            	log.Fatalln("failed to create dir  - ", err)
	        http.Error(w, err.Error(), http.StatusInternalServerError)
	        return err
            }
        case tar.TypeReg:
            outFile, err := os.Create(newTarget)

	    if err != nil {
            	log.Fatalln("failed create target ", newTarget, " - ", err)
	        http.Error(w, err.Error(), http.StatusInternalServerError)
	        return err
            }
            if _, err := io.Copy(outFile, tarReader); err != nil {
            	log.Fatalln("failed to copy from src to target ", newTarget, " - ", err)
	        http.Error(w, err.Error(), http.StatusInternalServerError)
	        return err
            }
	    outFile.Close()

        default:
            log.Fatalf(
                "ExtractTarGz: uknown type: %s in %s",
                header.Typeflag,
                header.Name)
	        http.Error(w, err.Error(), http.StatusInternalServerError)
	        return err
        }

    }
    log.Println("Extracted file: ", srcFile)
    return nil
}
