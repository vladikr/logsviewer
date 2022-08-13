package backend

import (
    "fmt"
    "net/http"
    "io"
    "os"
    "errors"
    "encoding/json"
    "io/ioutil"

    "logsviewer/pkg/backend/log"

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
	if err := regenerateEnrichmentData(); err != nil {
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
  mux.HandleFunc("/uploadLogs", uploadLogs)
  log.Log.Println("Routes set")
  return mux

}

