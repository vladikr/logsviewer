package backend

import (
    "fmt"
    "net/http"
    "io"
    "os"
    "errors"
    "time"
    "encoding/json"
    "io/ioutil"
    "strconv"

    "logsviewer/pkg/backend/log"
    "logsviewer/pkg/backend/db"

    "github.com/gorilla/websocket"
)

const (
    ENRICHMENT_DATA_FILE = "/space/result.json"
)

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,

    // We'll need to check the origin of our connection
    // this will allow us to make requests from our React
    // development server to here.
    // For now, we'll do no checking and just allow any connection
    CheckOrigin: func(r *http.Request) bool { return true },
}

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
func reader(conn *websocket.Conn) {
    for {
    // read in a message
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Log.Println(err)
            return
        }
    // print out that message for clarity
        fmt.Println(string(p))

        if err := conn.WriteMessage(messageType, p); err != nil {
            log.Log.Println(err)
            return
        }

    }
}

// define our WebSocket endpoint
func serveWs(w http.ResponseWriter, r *http.Request) {
    fmt.Println(r.Host)

  // upgrade this connection to a WebSocket
  // connection
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Log.Println(err)
  }
  // listen indefinitely for new messages coming
  // through on our WebSocket connection
    reader(ws)
}

func getPods(w http.ResponseWriter, r *http.Request) {
    log.Log.Println("Get Pods Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
            params[k] = v[0]
    }

	currentPage := 1

    page, err   := strconv.Atoi(fmt.Sprint(params["page"]))
    if err == nil {
        if page >= 1 {
            currentPage = page
        }
    }

    pageSize := -1
    perPage, err := strconv.Atoi(fmt.Sprint(params["per_page"]))
    if err == nil {
        if perPage >= 1 {
            pageSize = perPage
        }  
    }

    dbInst, err := db.NewDatabaseInstance()
    if err != nil {
        log.Log.Println("failed to connect to database", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer dbInst.Shutdown()

	data, err := dbInst.GetPods(currentPage, pageSize)
    if err != nil {
        log.Log.Println("failed to get pods!", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
    w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)  
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    if err1 := enc.Encode(data); err1 != nil {
        fmt.Println(err1.Error())
    }    
}

func getVmis(w http.ResponseWriter, r *http.Request) {
    log.Log.Println("Get Vmis Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
            params[k] = v[0]
    }

	currentPage := 1

    page, err   := strconv.Atoi(fmt.Sprint(params["page"]))
    if err == nil {
        if page >= 1 {
            currentPage = page
        }
    }

    pageSize := -1
    perPage, err := strconv.Atoi(fmt.Sprint(params["per_page"]))
    if err == nil {
        if perPage >= 1 {
            pageSize = perPage
        }  
    }

    dbInst, err := db.NewDatabaseInstance()
    if err != nil {
        log.Log.Println("failed to connect to database", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer dbInst.Shutdown()

	data, err := dbInst.GetVmis(currentPage, pageSize)
    if err != nil {
        log.Log.Println("failed to get pods!", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
    w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)  
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    if err1 := enc.Encode(data); err1 != nil {
        fmt.Println(err1.Error())
    }    
}

func getVmiMigrations(w http.ResponseWriter, r *http.Request) {
    log.Log.Println("Get Vmi migrations Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
            params[k] = v[0]
    }

	currentPage := 1

    page, err   := strconv.Atoi(fmt.Sprint(params["page"]))
    if err == nil {
        if page >= 1 {
            currentPage = page
        }
    }

    pageSize := -1
    perPage, err := strconv.Atoi(fmt.Sprint(params["per_page"]))
    if err == nil {
        if perPage >= 1 {
            pageSize = perPage
        }  
    }

    dbInst, err := db.NewDatabaseInstance()
    if err != nil {
        log.Log.Println("failed to connect to database", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer dbInst.Shutdown()

	data, err := dbInst.GetVmiMigrations(currentPage, pageSize)
    if err != nil {
        log.Log.Println("failed to get pods!", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
    w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)  
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    if err1 := enc.Encode(data); err1 != nil {
        fmt.Println(err1.Error())
    }    
}

func formatSingleVMIDSLQuery(res db.QueryResults) string {
//2022-08-17T23:00:00.000Z
//queryTemplate := ```(filters:!(),refreshInterval:(pause:!t,value:0),time:(from:'%s',to:'%s'))&_a=(columns:!(msg,podName,component,uid,subcomponent,reason,enrichment_data.pod.uid,enrichment_data.host.name,level),filters:!(('$state':(store:appState),meta:(alias:!n,disabled:!t,key:msg,negate:!f,type:exists,value:exists),query:(exists:(field:msg))),('$state':(store:appState),meta:(alias:!n,disabled:!t,key:msg,negate:!t,params:(query:'certificate%%20with%%20common%%20name%%20!'kubevirt.io:system:client:virt-handler!'%%20retrieved.'),type:phrase),query:(match_phrase:(msg:'certificate%%20with%%20common%%20name%%20!'kubevirt.io:system:client:virt-handler!'%%20retrieved.')))),interval:auto,query:(language:kuery,query:'containerName:%%20%%22virt-controller%%22%%20or%%20containerName:%%20%%22virt-api%%22%%20or%%20podName:%%20%%22%s%%22%%20or%%20podName:%%20%%22%s%%22%%20or%%20podName:%%20%%22%s%%22%%20or%%20podName:%%20%%22%s%%22%%20or%%20%%22%s%%22%%20or%%20%%22%s%%22'),sort:!(!('@timestamp',asc)))```

    queryTemplate := `_q=(filters:!(),refreshInterval:(pause:!t,value:0),time:(from:'%s',to:now))&_a=(columns:!(msg,podName,component,uid,subcomponent,reason,enrichment_data.pod.uid,enrichment_data.host.name,level),filters:!(('$state':(store:appState),meta:(alias:!n,disabled:!t,key:msg,negate:!f,type:exists,value:exists),query:(exists:(field:msg))),('$state':(store:appState),meta:(alias:!n,disabled:!t,key:msg,negate:!t,params:(query:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.'),type:phrase),query:(match_phrase:(msg:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.')))),interval:auto,query:(language:kuery,query:'containerName: "virt-controller" or containerName: "virt-api" or podName: "%s" or podName: "%s" or "%s" or "%s"'),sort:!(!('@timestamp',asc)))`


    timeStamp := res.StartTimestamp.Format(time.RFC3339)

    vmiLogsQuery := fmt.Sprintf(queryTemplate, timeStamp, res.SourcePod, res.SourceHandler, res.SourcePodUUID, res.VMIUUID)

    return vmiLogsQuery
}


func getVMIQueryParams(w http.ResponseWriter, r *http.Request) {
    log.Log.Println("Get VMI Query Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
            params[k] = v[0]
    }

	//currentPage := 1
    
    vmiUUID, exist := params["vmiUUID"]
    if !exist {
        log.Log.Println("can't find uuid in query params")
		http.Error(w, "can't find uuid in query params", http.StatusInternalServerError)
        return
    }
    nodeName, exist := params["nodeName"]
    if !exist {
        log.Log.Println("can't find nodeName in query params")
		http.Error(w, "can't find nodeName in query params", http.StatusInternalServerError)
        return
    }

    dbInst, err := db.NewDatabaseInstance()
    if err != nil {
        log.Log.Println("failed to connect to database", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer dbInst.Shutdown()
    vmiUUIDStr := fmt.Sprintf("%s", vmiUUID)
    nodeNameStr := fmt.Sprintf("%s", nodeName)

	data, err := dbInst.GetVMIQueryParams(vmiUUIDStr, nodeNameStr)
    if err != nil {
        log.Log.Println("failed to fetch VMI params", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
    dslQuery := formatSingleVMIDSLQuery(data)
    resp := map[string]string{"dslQuery": dslQuery}
    log.Log.Println("getVMIQueryParams encoded: ", resp)
    w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)  
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    if err1 := enc.Encode(resp); err1 != nil {
        log.Log.Println("getVMIQueryParams error: ", err1)
        fmt.Println(err1.Error())
    }    
    log.Log.Println("getVMIQueryParams encoded: ", resp)
}

/*func getMigrationQueryParams(w http.ResponseWriter, r *http.Request) {
    log.Log.Println("Get Migration Query Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
            params[k] = v[0]
    }

	currentPage := 1

    migrationUUID, err   := strconv.Atoi(fmt.Sprint(params["uuid"]))
    if err != nil {
        log.Log.Println("failed to connect to database", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    dbInst, err := db.NewDatabaseInstance()
    if err != nil {
        log.Log.Println("failed to connect to database", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer dbInst.Shutdown()

	data, err := dbInst.getMigrationQueryParams(migrationUUID)
    if err != nil {
        log.Log.Println("failed to fetch migration params", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
    w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)  
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    if err1 := enc.Encode(data); err1 != nil {
        fmt.Println(err1.Error())
    }    
}*/

func uploadLogs(w http.ResponseWriter, r *http.Request) {
    fmt.Println("File Upload Endpoint Hit")
    log.Log.Println("File Upload Endpoint Hit")

    // Parse our multipart form, 10 << 20 specifies a maximum
    // upload of 10 MB files.
    r.ParseMultipartForm(10 << 20)
    // FormFile returns the first file for the given key `myFile`
    // it also returns the FileHeader so we can get the Filename,
    // the Header and the size of the file
    file, handler, err := r.FormFile("file")
    if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
        fmt.Println("Error Retrieving the File")
        fmt.Println(err)
        log.Log.Println(err)
        return
    }
    defer file.Close()
    fmt.Printf("Uploaded File: %+v\n", handler.Filename)
    log.Log.Println("Uploaded File: ", handler.Filename)
    fmt.Printf("File Size: %+v\n", handler.Size)
    fmt.Printf("MIME Header: %+v\n", handler.Header)


    // TODO: make this path configurable
    err = os.MkdirAll("/space", os.ModePerm)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    destinationFilePath := fmt.Sprintf("/space/%s", handler.Filename)
    dst, err := os.Create(destinationFilePath)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
	return
    } 

    defer dst.Close()

    // Copy the uploaded file to the filesystem
    // at the specified destination
    _, err = io.Copy(dst, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Println("Successfully Uploaded File: ", handler.Filename)
    log.Log.Println("Successfully Uploaded File: ", handler.Filename)
    w.Header().Set("Content-Type", "application/json;charset=utf-8")
    json.NewEncoder(w).Encode(map[string]interface{}{
       "success":     true,
       "description": "Successfully Uploaded File",
    })

    mime := handler.Header.Get("Content-Type")
    if mime == "application/gzip" {
        if err := handleTarGz(destinationFilePath, "/space"); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        logsHandler := NewLogsHandler()
        defer close(logsHandler.stopCh)
        if err := logsHandler.processPodYAMLs(); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        if err := logsHandler.processVirtualMachineInstanceMigrationsYAMLs(); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        if err := logsHandler.processVirtualMachineInstanceYAMLs(); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }
}

func verifyFiles() {
    if _, err := os.Stat(ENRICHMENT_DATA_FILE); errors.Is(err, os.ErrNotExist) {
        m := make(map[string]string)
        content, _ := json.Marshal(m)
    	ioutil.WriteFile(ENRICHMENT_DATA_FILE, content, 0644)
    }
}

func SetupRoutes(publicDir string) *http.ServeMux {
  verifyFiles()
  mux := http.NewServeMux()
  web := http.FileServer(http.Dir(publicDir))
    
  mux.Handle("/", web)
  //TODO: move to an API sub
  mux.HandleFunc("/uploadLogs", uploadLogs)
  mux.HandleFunc("/pods", getPods)
  mux.HandleFunc("/vmis", getVmis)
  mux.HandleFunc("/vmims", getVmiMigrations)
  mux.HandleFunc("/getVMIQueryParams", getVMIQueryParams)
  //mux.HandleFunc("/getMigrationQueryParams", getMigrationQueryParams)
  log.Log.Println("Routes set")
  return mux

}

