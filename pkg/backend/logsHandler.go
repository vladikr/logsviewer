package backend

import (
    "path/filepath"
    "archive/tar"
    "compress/gzip"
    "io"
    "os"
    "strings"
    "fmt"
    "io/ioutil"
    "encoding/json"

    "logsviewer/pkg/backend/log"
    "gopkg.in/yaml.v2"
)

type Pods struct {
    Metadata Metadata `yaml:"metadata"`
    Spec     Spec     `yaml:"spec"`
    Status   Status   `yaml:"status"`
}

type Metadata struct {
    Namespace       string              `yaml:"namespace,omitempty"`
    Name            string              `yaml:"name,omitempty"`
    OwnerReferences []OwnerReference    `yaml:"ownerReferences,omitempty"`
    UID             string              `yaml:"uid,omitempty"`
}

type Spec struct {
    NodeName string
}

type Status struct {
    HostIP string
}

type OwnerReference struct {
    UID string
}

type EnrichmentData struct {
    HostName        string      `json:"host.name"`
    HostIP          string      `json:"host.ip"`
    UID             string      `json:"pod.uid"`
    OwnerReferences []string    `json:"pod.ownerReferences,omitempty"`
}


func handleTarGz(srcFile string, targetPath string) error {
    if err := unTarGz(srcFile, targetPath); err != nil {
        return err
    }    
    // delete source file
    if err := os.Remove(srcFile); err != nil {
        log.Log.Fatalln("failed to delete file ", srcFile, " - ", err)
    }
    return nil
}


func unTarGz(srcFile string, targetPath string) error {
    gzipStream, err := os.Open(srcFile)
    defer gzipStream.Close()

    if err != nil {
        log.Log.Fatalln("failed to open file ", srcFile, " - ", err)
	return err
    }

    pathToNamespacesDir := ""
    uncompressedStream, err := gzip.NewReader(gzipStream)
    if err != nil {
        log.Log.Fatalln("failed create gzip stream  - ", err)
	return err
    }

    tarReader := tar.NewReader(uncompressedStream)

    for true {
        header, err := tarReader.Next()

        if err == io.EOF {
            break
        }

        if err != nil {
            log.Log.Fatalln("failed to get next file in tar  - ", err)
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
            	log.Log.Fatalln("failed to create dir  - ", err)
	        return err
            }
        case tar.TypeReg:
            outFile, err := os.Create(newTarget)

	    if err != nil {
            	log.Log.Fatalln("failed create target ", newTarget, " - ", err)
	        return err
            }
            if _, err := io.Copy(outFile, tarReader); err != nil {
            	log.Log.Fatalln("failed to copy from src to target ", newTarget, " - ", err)
	        return err
            }
	    outFile.Close()

        default:
            log.Log.Fatalf(
                "ExtractTarGz: uknown type: %s in %s",
                header.Typeflag,
                header.Name)
	        return err
        }

    }
    log.Log.Println("Extracted file: ", srcFile)
    return nil
}

func regenerateEnrichmentData() error {
    var pod Pods
    lookupData := make(map[string]EnrichmentData)

    // read the existing enrichment data file
    jsonFile, err := os.Open(ENRICHMENT_DATA_FILE)
    if err != nil {
        return(err)
    }
    defer jsonFile.Close()

    byteValue, _ := ioutil.ReadAll(jsonFile)
    json.Unmarshal([]byte(byteValue), &lookupData)


    //TODO: make path configurable
    layouts, err := filepath.Glob("/space/namespaces/*/pods/*/*.yaml")
    if err != nil {
        return(err)
    }

    for _, filename := range layouts {
          // read pod yaml
	  yamlFile, err := ioutil.ReadFile(filename)
	  if err != nil {
	    return(err)
	  }
	  err = yaml.Unmarshal(yamlFile, &pod)
	  if err != nil {
	    return(err)
	  }

          // generate enrichment data
	  key := fmt.Sprintf("%s/%s", pod.Metadata.Namespace, pod.Metadata.Name)
	  
          enrichmentData := EnrichmentData{
		  HostName: pod.Spec.NodeName,
		  HostIP:   pod.Status.HostIP,
		  UID:      pod.Metadata.UID,
          }
          if pod.Metadata.OwnerReferences != nil {
              for _, ref := range pod.Metadata.OwnerReferences {
                  enrichmentData.OwnerReferences = append(enrichmentData.OwnerReferences, ref.UID)
              }
	  }
          lookupData[key] = enrichmentData


    }
    js1, _ := json.Marshal(lookupData)
    _ = ioutil.WriteFile(ENRICHMENT_DATA_FILE, js1, 0644)
    
    return nil
}
