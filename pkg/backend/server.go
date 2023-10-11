package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	"logsviewer/pkg/backend/cleanup"
	"logsviewer/pkg/backend/db"
	"logsviewer/pkg/backend/env"
	"logsviewer/pkg/backend/insights"
	"logsviewer/pkg/backend/log"
	"logsviewer/pkg/backend/monitoring/metrics"
)

const (
	ENRICHMENT_DATA_FILE = "/space/result.json"
)

type app struct {
	storeDB          *db.DatabaseInstance
	insightsInstance *insights.Insights
}

func NewAppInstance() (*app, error) {
	newAppInstance := &app{}
	if err := newAppInstance.initStoreDB(); err != nil {
		return newAppInstance, err
	}
	return newAppInstance, nil
}

func (c *app) WithInsights(insightsBinaryPath string) {
	c.insightsInstance = insights.New(insightsBinaryPath)
}

func (c *app) IsInsightsEnabled() bool {
	return c.insightsInstance != nil
}

func (c *app) initStoreDB() error {
	dbInst, err := db.NewDatabaseInstance()
	if err != nil {
		msg := "failed to connect to database - "
		log.Log.Println(msg, err)
		return fmt.Errorf("%s%s", msg, err.Error())
	}
	c.storeDB = dbInst
	if err := c.storeDB.InitTables(); err != nil {
		log.Log.Println("failed to connect to database")
		if err := c.storeDB.DropTables(); err != nil {
			log.Log.Println("failed to drop tables", err)
		}
		c.storeDB.Shutdown()
		c.storeDB = nil
		msg := "failed to initalize the database"
		log.Log.Println(msg, err)
		return fmt.Errorf("%s: %s", msg, err.Error())
	}
	return nil
}

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

func (c *app) getPods(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get Pods Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	queryDetails := db.GenericQueryDetails{}
	if podName, exist := params["name"]; exist {
		json.Unmarshal([]byte(fmt.Sprint(podName)), &queryDetails)
	}
	if podNamespace, exist := params["namespace"]; exist {
		json.Unmarshal([]byte(fmt.Sprint(podNamespace)), &queryDetails)
	}
	if podUUID, exist := params["uuid"]; exist {
		json.Unmarshal([]byte(fmt.Sprint(podUUID)), &queryDetails)
	}
	if status, exist := params["status"]; exist {
		queryDetails.Status = fmt.Sprint(status)
	}

	currentPage := 1

	page, err := strconv.Atoi(fmt.Sprint(params["page"]))
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

	data, err := c.storeDB.GetPods(currentPage, pageSize, &queryDetails)
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

func (c *app) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (c *app) getNodes(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get nodes Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	currentPage := 1

	page, err := strconv.Atoi(fmt.Sprint(params["page"]))
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

	queryDetails := db.GenericQueryDetails{}
	if status, exist := params["status"]; exist {
		queryDetails.Status = fmt.Sprint(status)
	}

	data, err := c.storeDB.GetNodes(currentPage, pageSize, &queryDetails)
	if err != nil {
		log.Log.Println("failed to get nodes!", err)
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

// getVmis returns a list of Virtual Machine objects
func (c *app) getVms(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get Vms Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	currentPage := 1

	page, err := strconv.Atoi(fmt.Sprint(params["page"]))
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

	queryDetails := db.GenericQueryDetails{}
	if status, exist := params["status"]; exist {
		queryDetails.Status = fmt.Sprint(status)
	}

	data, err := c.storeDB.GetVms(currentPage, pageSize, &queryDetails)
	if err != nil {
		log.Log.Println("failed to get VMs!", err)
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

func (c *app) getVmis(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get Vmis Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	currentPage := 1

	page, err := strconv.Atoi(fmt.Sprint(params["page"]))
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

	queryDetails := db.GenericQueryDetails{}
	if status, exist := params["status"]; exist {
		queryDetails.Status = fmt.Sprint(status)
	}

	data, err := c.storeDB.GetVmis(currentPage, pageSize, &queryDetails)
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

func (c *app) getVmiMigrations(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get Vmi migrations Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	vmiDetails := db.GenericQueryDetails{}
	if vmiName, exist := params["name"]; exist {
		log.Log.Println("vmiName: ", vmiName)
		json.Unmarshal([]byte(fmt.Sprint(vmiName)), &vmiDetails)
		log.Log.Println("vmiDetails: ", vmiDetails)
	}
	if vmiNamespace, exist := params["namespace"]; exist {
		log.Log.Println("Namespace: ", vmiNamespace)
		json.Unmarshal([]byte(fmt.Sprint(vmiNamespace)), &vmiDetails)
		log.Log.Println("vmiDetails: ", vmiDetails)
	}
	if status, exist := params["status"]; exist {
		vmiDetails.Status = fmt.Sprint(status)
	}

	currentPage := 1

	page, err := strconv.Atoi(fmt.Sprint(params["page"]))
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

	data, err := c.storeDB.GetVmiMigrations(currentPage, pageSize, &vmiDetails)
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

func (c *app) getPodPVCs(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get Pod PVCs Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	podUUID, exist := params["uuid"]
	if !exist {
		log.Log.Println("can't find uuid in query params")
		http.Error(w, "can't find uuid in query params", http.StatusInternalServerError)
		return
	}

	data, err := c.storeDB.GetPodPVCs(fmt.Sprintf("%s", podUUID))
	if err != nil {
		log.Log.Println("failed to get pvcs from database", err)
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

func (c *app) getVMIPVCs(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get VMI PVCs Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	vmiUUID, exist := params["uuid"]
	if !exist {
		log.Log.Println("can't find uuid in query params")
		http.Error(w, "can't find uuid in query params", http.StatusInternalServerError)
		return
	}

	data, err := c.storeDB.GetVMIPVCs(fmt.Sprintf("%s", vmiUUID))
	if err != nil {
		log.Log.Println("failed to get pvcs from database", err)
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

func (c *app) getPVCs(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get PVCs Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	queryDetails := db.GenericQueryDetails{}
	if pvcName, exist := params["name"]; exist {
		json.Unmarshal([]byte(fmt.Sprint(pvcName)), &queryDetails)
	}
	if pvcNamespace, exist := params["namespace"]; exist {
		json.Unmarshal([]byte(fmt.Sprint(pvcNamespace)), &queryDetails)
	}
	if pvcUUID, exist := params["uuid"]; exist {
		json.Unmarshal([]byte(fmt.Sprint(pvcUUID)), &queryDetails)
	}
	if status, exist := params["status"]; exist {
		queryDetails.Status = fmt.Sprint(status)
	}

	currentPage := 1

	page, err := strconv.Atoi(fmt.Sprint(params["page"]))
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

	data, err := c.storeDB.GetPVCs(currentPage, pageSize, &queryDetails)
	if err != nil {
		log.Log.Println("failed to get pvcs from database", err)
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

func (c *app) getSubscriptions(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get Subscriptions Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	currentPage := 1

	page, err := strconv.Atoi(fmt.Sprint(params["page"]))
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

	data, err := c.storeDB.GetSubscriptions(currentPage, pageSize)
	if err != nil {
		log.Log.Println("failed to get subscriptions!", err)
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

func (c *app) getImportedMustGathers(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get Imported MustGathers Endpoint Hit: ", r.URL.Query())
	data, err := c.storeDB.ListImportedMustGather()
	if err != nil {
		log.Log.Println("failed to get imported must gathers!", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	output := map[string]interface{}{
		"data": data,
	}
	if err1 := enc.Encode(output); err1 != nil {
		fmt.Println(err1.Error())
	}
}

func (c *app) getVMIQueryParams(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get VMI Query Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

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

	vmiUUIDStr := fmt.Sprintf("%s", vmiUUID)
	nodeNameStr := fmt.Sprintf("%s", nodeName)

	data, err := c.storeDB.GetVMIQueryParams(vmiUUIDStr, nodeNameStr)
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

func (c *app) getFullVMIHistoryQueryParams(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get VMI Full Query Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	vmiUUID, exist := params["vmiUUID"]
	if !exist {
		log.Log.Println("can't find uuid in query params")
		http.Error(w, "can't find uuid in query params", http.StatusInternalServerError)
		return
	}

	vmiUUIDStr := fmt.Sprintf("%s", vmiUUID)

	data, err := c.storeDB.GetFullVMIHistoryQueryParams(vmiUUIDStr)
	if err != nil {
		log.Log.Println("failed to fetch VMI params", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dslQuery := formatFullVMIHistoryDSLQuery(data)
	resp := map[string]string{"dslQuery": dslQuery}
	log.Log.Println("getFullVMIHistoryQueryParams encoded: ", data)
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err1 := enc.Encode(resp); err1 != nil {
		log.Log.Println("getVMIQueryParams error: ", err1)
		fmt.Println(err1.Error())
	}
	log.Log.Println("getFullVMIHistoryQueryParams encoded: ", resp)
}

func (c *app) getVMIDetails(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get VMI Details Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	vmiUUID, exist := params["uuid"]
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

	vmiUUIDStr := fmt.Sprintf("%s", vmiUUID)
	nodeNameStr := fmt.Sprintf("%s", nodeName)

	data, err := c.storeDB.GetVMIQueryParams(vmiUUIDStr, nodeNameStr)
	if err != nil {
		log.Log.Println("failed to fetch VMI params", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err1 := enc.Encode(data); err1 != nil {
		log.Log.Println("getVMIDetails error: ", err1)
		fmt.Println(err1.Error())
	}
	log.Log.Println("getVMIDetails encoded: ", data)
}

func (c *app) getMigrationQueryParams(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get Migration Query Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	migrationUUID, exist := params["uuid"]
	if !exist {
		log.Log.Println("failed to find uuid in the migrationQuery Params")
		http.Error(w, "failed to find uuid in the migrationQuery Params", http.StatusInternalServerError)
		return
	}

	data, err := c.storeDB.GetMigrationQueryParams(fmt.Sprintf("%s", migrationUUID))
	if err != nil {
		log.Log.Println("failed to fetch migration params", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dslQuery := formatVMIMigrationDSLQuery(data)
	resp := map[string]string{"dslQuery": dslQuery}
	log.Log.Println("getMigrationQueryParams encoded: ", resp)
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err1 := enc.Encode(resp); err1 != nil {
		fmt.Println(err1.Error())
	}
}

func (c *app) getObjYaml(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get Object Yaml Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}
	objType, exist := params["object"]
	if !exist {
		log.Log.Println("failed to find object type in the Query Params")
		http.Error(w, "failed to find object type in the Query Params", http.StatusInternalServerError)
		return
	}

	UUID, exist := params["uuid"]
	if !exist {
		log.Log.Println("failed to find uuid in the podQuery Params")
		http.Error(w, "failed to find uuid in the podQuery Params", http.StatusInternalServerError)
		return
	}

	var retObject interface{}
	var err error
	switch objType {
	case "pod":
		retObject, err = c.storeDB.GetPodObject(fmt.Sprintf("%s", UUID))
		if err != nil {
			log.Log.Println("failed to fetch pod params", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "vm":
		retObject, err = c.storeDB.GetVMObject(fmt.Sprintf("%s", UUID))
		if err != nil {
			log.Log.Println("failed to fetch vm params", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "vmi":
		retObject, err = c.storeDB.GetVMIObject(fmt.Sprintf("%s", UUID))
		if err != nil {
			log.Log.Println("failed to fetch vmi params", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "node":
		retObject, err = c.storeDB.GetNodeObject(fmt.Sprintf("%s", UUID))
		if err != nil {
			log.Log.Println("failed to fetch node params", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "pvc":
		retObject, err = c.storeDB.GetPVCObject(fmt.Sprintf("%s", UUID))
		if err != nil {
			log.Log.Println("failed to fetch pvc params", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// convert Pod Object to Yaml
	outYaml, errYaml := yaml.Marshal(retObject)
	if errYaml != nil {
		http.Error(w, "failed marshal obj to yaml", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"yaml": string(outYaml)}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err1 := enc.Encode(resp); err1 != nil {
		fmt.Println(err1.Error())
	}
}

func (c *app) getPodYaml(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get Pod Yaml Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	podUUID, exist := params["uuid"]
	if !exist {
		log.Log.Println("failed to find uuid in the podQuery Params")
		http.Error(w, "failed to find uuid in the podQuery Params", http.StatusInternalServerError)
		return
	}

	podObj, err := c.storeDB.GetPodObject(fmt.Sprintf("%s", podUUID))
	if err != nil {
		log.Log.Println("failed to fetch pod params", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// convert Pod Object to Yaml
	outYaml, errYaml := yaml.Marshal(podObj)
	if errYaml != nil {
		http.Error(w, "failed marshal pod to yaml", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"yaml": string(outYaml)}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err1 := enc.Encode(resp); err1 != nil {
		fmt.Println(err1.Error())
	}
}

func (c *app) getSinglePodQueryParams(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get Pod Query Endpoint Hit: ", r.URL.Query())
	params := map[string]interface{}{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	podUUID, exist := params["uuid"]
	if !exist {
		log.Log.Println("failed to find uuid in the podQuery Params")
		http.Error(w, "failed to find uuid in the podQuery Params", http.StatusInternalServerError)
		return
	}

	data, err := c.storeDB.GetPodQueryParams(fmt.Sprintf("%s", podUUID))
	if err != nil {
		log.Log.Println("failed to fetch pod params", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dslQuery := formatSinglePodDSLQuery(data)
	resp := map[string]string{"dslQuery": dslQuery}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err1 := enc.Encode(resp); err1 != nil {
		fmt.Println(err1.Error())
	}
}

func (c *app) getResourceStats(w http.ResponseWriter, r *http.Request) {
	log.Log.Println("Get Resource Stats Endpoint Hit: ", r.URL.Query())

	data, err := c.storeDB.GetResourceStats()
	if err != nil {
		log.Log.Println("failed to fetch resource stats", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("data", data)

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err1 := enc.Encode(data); err1 != nil {
		fmt.Println(err1.Error())
	}
}

func (c *app) uploadLogs(w http.ResponseWriter, r *http.Request) {
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
	d, err := os.Open("/space")
	if err == nil {
		defer d.Close()
		names, err := d.Readdirnames(-1)
		if err == nil {
			for _, name := range names {
				_ = os.RemoveAll(filepath.Join("/space", name))
			}
		}
	}

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

	log.Log.Println("Successfully Uploaded File: ", handler.Filename)
	metrics.NewMustGatherUploaded()

	_, exists, err := c.storeDB.GetImportedMustGather(handler.Filename)
	if err != nil {
		log.Log.Println("failed to fetch imported must gather", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		log.Log.Println("must gather already exists")
		http.Error(w, "Must gather already exists", http.StatusConflict)
		return
	}

	mime := handler.Header.Get("Content-Type")
	if mime == "application/gzip" || mime == "application/x-gzip" {
		if err := handleTarGz(destinationFilePath, "/space"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var insightsData string
		if c.IsInsightsEnabled() {
			log.Log.Println("executing insights", "/space")
			insightsDataRaw, insightsErr := c.insightsInstance.Exec("/space")
			if insightsErr != nil {
				log.Log.Println("failed to execute insights", insightsErr)
			}
			insightsData = string(insightsDataRaw)
			fmt.Println("insightsData", insightsData)
		} else {
			log.Log.Println("insights is disabled")
		}

		logsHandler := NewLogsHandler(c.storeDB)
		defer close(logsHandler.stopCh)
		var errList []error
		if err := logsHandler.processImportedMustGather(handler.Filename, insightsData); err != nil {
			log.Log.Println("failed to store imported must gather", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := logsHandler.processPodYAMLs(); err != nil {
			errList = append(errList, err)
		}
		if err := logsHandler.processVirtualMachineInstanceMigrationsYAMLs(); err != nil {
			errList = append(errList, err)
		}
		if err := logsHandler.processNodeYAMLs(); err != nil {
			errList = append(errList, err)
		}
		if err := logsHandler.processVirtualMachineYAMLs(); err != nil {
			errList = append(errList, err)
		}
		if err := logsHandler.processVirtualMachineInstanceYAMLs(); err != nil {
			errList = append(errList, err)
		}
		if err := logsHandler.processPersistentVolumeClaimYAMLs(); err != nil {
			errList = append(errList, err)
		}
		if err := logsHandler.processSubscriptionsYAMLs(); err != nil {
			errList = append(errList, err)
		}

		if len(errList) > 0 {
			http.Error(w, fmt.Sprintf("failed to process YAMLs: %v", errList), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"description": "Successfully Uploaded File",
		})
	}
}

func verifyFiles() {
	if _, err := os.Stat(ENRICHMENT_DATA_FILE); errors.Is(err, os.ErrNotExist) {
		m := make(map[string]string)
		content, _ := json.Marshal(m)
		ioutil.WriteFile(ENRICHMENT_DATA_FILE, content, 0644)
	}
}

func setKibanaDefaultDataView() {
	httpposturl := "http://localhost:5601/api/data_views/default"
	log.Log.Println("HTTP JSON POST URL:", httpposturl)

	var jsonData = []byte(`{"data_view_id": "cnvlogs-default"}`)
	request, err := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Add("kbn-xsrf", "true")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Log.Println("ERROR: ", err)
	}
	defer response.Body.Close()

	log.Log.Println("response Status:", response.Status)
	log.Log.Println("response Headers:", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	log.Log.Println("response Body:", string(body))
}

func createKibanaDataView() {
	httpposturl := "http://localhost:5601/api/data_views/data_view"
	log.Log.Println("HTTP JSON POST URL:", httpposturl)

	var jsonData = []byte(`{"data_view": {"title": "cnvlogs*", "timeFieldName":"@timestamp", "id":"cnvlogs-default"}}`)
	request, err := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Add("kbn-xsrf", "true")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Log.Println("ERROR: ", err)
	}
	defer response.Body.Close()

	log.Log.Println("response Status:", response.Status)
	log.Log.Println("response Headers:", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	log.Log.Println("response Body:", string(body))
}

func SetupRoutes(publicDir *string, insightsBinaryPath string) (*http.ServeMux, error) {
	verifyFiles()
	createKibanaDataView()
	setKibanaDefaultDataView()
	app, err := NewAppInstance()
	if err != nil {
		return nil, err
	}

	if insightsBinaryPath != "" {
		app.WithInsights(insightsBinaryPath)
	}

	// Register metrics in Prometheus
	metrics.SetupMetrics()
	metrics.InstanceCreated()

	mux := http.NewServeMux()
	web := http.FileServer(http.Dir(*publicDir))

	mux.Handle("/", web)
	//TODO: move to an API sub
	mux.HandleFunc("/healthz", app.healthz)
	mux.HandleFunc("/uploadLogs", app.uploadLogs)
	mux.HandleFunc("/pods", app.getPods)
	mux.HandleFunc("/nodes", app.getNodes)
	mux.HandleFunc("/vms", app.getVms)
	mux.HandleFunc("/vmis", app.getVmis)
	mux.HandleFunc("/vmims", app.getVmiMigrations)
	mux.HandleFunc("/getPVCs", app.getPVCs)
	mux.HandleFunc("/getPodPVCs", app.getPodPVCs)
	mux.HandleFunc("/getVMIPVCs", app.getVMIPVCs)
	mux.HandleFunc("/getVMIQueryParams", app.getVMIQueryParams)
	mux.HandleFunc("/getMigrationQueryParams", app.getMigrationQueryParams)
	mux.HandleFunc("/getSinglePodQueryParams", app.getSinglePodQueryParams)
	mux.HandleFunc("/getPodYaml", app.getPodYaml)
	mux.HandleFunc("/getObjYaml", app.getObjYaml)
	mux.HandleFunc("/getResourceStats", app.getResourceStats)
	mux.HandleFunc("/getSubscriptions", app.getSubscriptions)
	mux.HandleFunc("/getImportedMustGathers", app.getImportedMustGathers)
	mux.HandleFunc("/getVMIDetails", app.getVMIDetails)
	mux.HandleFunc("/getFullVMIHistoryQueryParams", app.getFullVMIHistoryQueryParams)

	mux.Handle("/metrics", promhttp.Handler())

	log.Log.Println("Routes set")
	return mux, nil
}

func Spawn(publicDir string, insightsBinaryPath string) error {
	mux, err := SetupRoutes(&publicDir, insightsBinaryPath)
	if err != nil {
		return err
	}

	go cleanup.StartCleanupJob(
		env.GetEnv("POD_NAME", ""),
		env.GetEnv("POD_NAMESPACE", ""),
	)

	http.ListenAndServe(":8080", cors.AllowAll().Handler(mux))
	return nil
}
