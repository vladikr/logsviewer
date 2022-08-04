package backend

import (
    "fmt"
    "log"
    "net/http"
    "io"
    "os"
    //"time"
    //"path/filepath"

    "github.com/gorilla/websocket"
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
            log.Println(err)
            return
        }
    // print out that message for clarity
        fmt.Println(string(p))

        if err := conn.WriteMessage(messageType, p); err != nil {
            log.Println(err)
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
        log.Println(err)
  }
  // listen indefinitely for new messages coming
  // through on our WebSocket connection
    reader(ws)
}

func uploadLogs(w http.ResponseWriter, r *http.Request) {
    fmt.Println("File Upload Endpoint Hit")

    // Parse our multipart form, 10 << 20 specifies a maximum
    // upload of 10 MB files.
    r.ParseMultipartForm(10 << 20)
    // FormFile returns the first file for the given key `myFile`
    // it also returns the FileHeader so we can get the Filename,
    // the Header and the size of the file
    file, handler, err := r.FormFile("file")
    if err != nil {
        fmt.Println("Error Retrieving the File")
        fmt.Println(err)
        return
    }
    defer file.Close()
    fmt.Printf("Uploaded File: %+v\n", handler.Filename)
    fmt.Printf("File Size: %+v\n", handler.Size)
    fmt.Printf("MIME Header: %+v\n", handler.Header)


    err = os.MkdirAll("/tmp/uploads", os.ModePerm)
    if err != nil {
        fmt.Println(err)
	return
    }

    dst, err := os.Create(fmt.Sprintf("/tmp/uploads/%s", handler.Filename))
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

    fmt.Fprintf(w, "Successfully Uploaded File\n")
}




func SetupRoutes() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Simple Server")
  })

  http.HandleFunc("/uploadLogs", uploadLogs)
  // mape our `/ws` endpoint to the `serveWs` function
    http.HandleFunc("/ws", serveWs)
}

