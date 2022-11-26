package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
    "strconv"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

    "logsviewer/pkg/backend/log"

	_ "github.com/go-sql-driver/mysql"
)

//var db *sql.DB

type (
	//StringInterfaceMap map[string]interface{}
	Pod struct {
		Key       string `json:"keyid"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		UUID      string `json:"uuid"`
        Phase     string `json:"phase"`
        ActiveContainers int `json:"activeContainers"`
        TotalContainers  int `json:"totalContainers"`
        NodeName         string `json:"nodeName"`
        CreationTime     metav1.Time `json:"creationTime"`
		Content json.RawMessage `json:"content"`
        CreatedBy        string `json:"createdBy"`
	}

	VirtualMachineInstance struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		UUID      string `json:"uuid"`
        Reason     string `json:"reason"`
        Phase     string `json:"phase"`
        NodeName         string `json:"nodeName"`
        CreationTime     metav1.Time `json:"creationTime"`
        //PodName   string `json:"podName"`
        //HandlerPod  string `json:"handlerName"`
        Status kubevirtv1.VirtualMachineInstanceStatus `json:"status,omitempty"`
		Content json.RawMessage `json:"content"`
	}

	VirtualMachineInstanceMigration struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		UUID      string `json:"uuid"`
        Phase     string `json:"phase"`
        VMIName         string `json:"vmiName"`
        // The target pod that the VMI is moving to
        TargetPod string `json:"targetPod,omitempty"`
        CreationTime metav1.Time `json:"creationTime"`
        EndTimestamp metav1.Time `json:"endTimestamp,omitempty"`
        SourceNode string `json:"sourceNode,omitempty"`
        // The target node that the VMI is moving to
        TargetNode string `json:"targetNode,omitempty"`
        // Indicates the migration completed
        Completed bool `json:"completed,omitempty"`
        // Indicates that the migration failed
        Failed bool `json:"failed,omitempty"`
        //PodName   string `json:"podName"`
        //HandlerPod  string `json:"handlerName"`
		Content json.RawMessage `json:"content"`
	}
)

func (d *databaseInstance) StorePod(pod *Pod) error {
	// TimeString - given a time, return the MySQL standard string representation
	madeAt := pod.CreationTime.Format("2006-01-02 15:04:05.999999")
	ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctx, insertPodQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(
        ctx,
        pod.Key,
        pod.Kind,
        pod.Name,
        pod.Namespace,
        pod.UUID,
        pod.Phase,
        pod.ActiveContainers,
        pod.TotalContainers,
        pod.NodeName,
        madeAt,
        pod.Content,
        pod.CreatedBy)
	if err != nil {
		return err
	}

	return nil
} 

func (d *databaseInstance) StoreVmi(vmi *VirtualMachineInstance) error {
	// TimeString - given a time, return the MySQL standard string representation
	madeAt := vmi.CreationTime.Format("2006-01-02 15:04:05.999999")
	ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctx, insertVmiQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(
        ctx,
        vmi.Name,
        vmi.Namespace,
        vmi.UUID,
        vmi.Reason,
        vmi.Phase,
        vmi.NodeName,
        madeAt,
        vmi.Content)
	if err != nil {
		return err
	}

    if migrationState := vmi.Status.MigrationState; migrationState != nil {
        if existngVmim, err := d.getSingleMigrationByUUID(string(migrationState.MigrationUID)); err == nil {
            log.Log.Println("no error from SingleMigrationByUUID for uuid: ", string(migrationState.MigrationUID))
            emptyContent := json.RawMessage(`{}`)
            newVmim := VirtualMachineInstanceMigration{
                Name: existngVmim.Name,
                Namespace: existngVmim.Namespace,
                UUID: string(migrationState.MigrationUID),
                Phase: string(existngVmim.Phase),
                VMIName: string(existngVmim.VMIName),
                TargetPod: migrationState.TargetPod,
                CreationTime: *migrationState.StartTimestamp,
                EndTimestamp: *migrationState.EndTimestamp,
                SourceNode: migrationState.SourceNode,
                TargetNode: migrationState.TargetNode,
                Completed: migrationState.Completed,
                Failed: migrationState.Failed,
                Content: emptyContent}
            
            log.Log.Println("SingleMigrationByUUID going to store: ", newVmim)
            if err := d.StoreVmiMigration(&newVmim); err != nil {
                log.Log.Println("SingleMigrationByUUID store ERROR: ", err, " for uuid: ", newVmim.UUID)
                
            }
        }
    }
	return nil
} 

func (d *databaseInstance) StoreVmiMigration(vmim *VirtualMachineInstanceMigration) error {
	// TimeString - given a time, return the MySQL standard string representation
	madeAt := vmim.CreationTime.Format("2006-01-02 15:04:05.999999")
	endedAt := vmim.EndTimestamp.Format("2006-01-02 15:04:05.999999")
	ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctx, insertVmiMigrationQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(
        ctx,
        vmim.Name,
        vmim.Namespace,
        vmim.UUID,
        vmim.Phase,
        vmim.VMIName,
        vmim.TargetPod,
        madeAt,
        endedAt,
        vmim.SourceNode,
        vmim.TargetNode,
        vmim.Completed,
        vmim.Failed,
        vmim.Content)
	if err != nil {
		return err
	}

	return nil
} 

var (
	insertPodQuery       = `INSERT INTO pods(keyid, kind, name, namespace, uuid, phase, activeContainers, totalContainers, nodeName, creationTime, content, createdBy) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE keyid=VALUES(keyid);`
	insertVmiQuery       = `INSERT INTO vmis(name, namespace, uuid, reason, phase, nodeName, creationTime, content) values (?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE uuid=VALUES(uuid);`
	insertVmiMigrationQuery       = `INSERT INTO vmimigrations(name, namespace, uuid, phase, vmiName, targetPod, creationTime, endTimestamp, sourceNode, targetNode, completed, failed, content) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE uuid=VALUES(uuid), targetPod=VALUES(targetPod), creationTime=VALUES(creationTime), endTimestamp=VALUES(endTimestamp), sourceNode=VALUES(sourceNode), targetNode=VALUES(targetNode), completed=VALUES(completed), failed=VALUES(failed);`
)

var (
	defaultUsername = "mysql"
	defaultPassword = "supersecret"
	//defaultHost     = "mysql"
	defaultHost     = "0.0.0.0"
	defaultPort     = "3306"
	defaultdbName   = "objtracker"
)

type databaseInstance struct {
	username string
	password string
	host     string
	port     string
	dbName   string
	db       *sql.DB
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewDatabaseInstance() (*databaseInstance, error) {
	dbInstance := &databaseInstance{
		username: defaultUsername,
		password: defaultPassword,
		host:     defaultHost,
		port:     defaultPort,
		dbName:   defaultdbName,
	}
	ctx, cancel := context.WithCancel(context.Background())
	dbInstance.ctx = ctx
	dbInstance.cancel = cancel
	err := dbInstance.connect()
	if err != nil {
        log.Log.Println("failed to connect to db: ", err)
		return nil, err
	}

	return dbInstance, nil
}

func (d *databaseInstance) Shutdown() (err error) {
	if d.cancel != nil {
		d.cancel()
	}

	if d.db != nil {
		d.db.Close()
	}
	return
}

func (d *databaseInstance) InitTables() (err error) {
	err = d.createTables()
	if err != nil {
		return err
	}

	return nil
}

func (d *databaseInstance) connect() (err error) {

	uri := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", d.username, d.password, d.host, d.port, d.dbName)

	db, err := sql.Open("mysql", uri)
	if err != nil {
		return err
	}

	d.db = db
	ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
	defer cancel()

	err = d.db.PingContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *databaseInstance) createTables() error {
    if err := d.createPodsTable(); err != nil {
		return err
	}
    if err := d.createVmisTable(); err != nil {
		return err
	}
    if err := d.createVmiMigrationsTable(); err != nil {
		return err
	}
	return nil
}

func (d *databaseInstance) createPodsTable() error {

	createPodsTable := `
	CREATE TABLE IF NOT EXISTS pods (
	  keyid varchar(100),
	  kind varchar(100),
	  name varchar(100),
	  namespace varchar(100),
	  uuid varchar(100),
      phase varchar(100),
      activeContainers TINYINT,
      totalContainers TINYINT,
      nodeName varchar(100),
      creationTime datetime,
      content json,
	  createdBy varchar(100),
	  PRIMARY KEY (uuid)
	);
	`
	err := d.execTable(createPodsTable)
	if err != nil {
		return err
	}

	return nil
}

func (d *databaseInstance) createVmisTable() error {

	vmisTableCreate := `
	CREATE TABLE IF NOT EXISTS vmis (
	  name varchar(100),
	  namespace varchar(100),
	  uuid varchar(100),
      reason varchar(100),
      phase varchar(100),
      nodeName varchar(100),
      creationTime datetime,
      content json,
	  createdBy varchar(100),
	  PRIMARY KEY (uuid)
	);
	`
	err := d.execTable(vmisTableCreate)
	if err != nil {
		return err
	}

	return nil
}

func (d *databaseInstance) createVmiMigrationsTable() error {

	vmimsTableCreate := `
	CREATE TABLE IF NOT EXISTS vmimigrations (
	  name varchar(100),
	  namespace varchar(100),
	  uuid varchar(100),
      phase varchar(100),
      vmiName varchar(100),
      targetPod varchar(100),
      creationTime datetime,
      endTimestamp datetime,
      sourceNode varchar(100),
      targetNode varchar(100),
      completed BOOLEAN,
      failed BOOLEAN,
      content json,
	  PRIMARY KEY (uuid)
	);
	`
	err := d.execTable(vmimsTableCreate)
	if err != nil {
		return err
	}

	return nil
}


func (d *databaseInstance) execTable(tableSql string) error {
	ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctx, tableSql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}
func (d *databaseInstance) DropTables() error {
	dropPodsTable := `
	DROP TABLE pods;
	`
	err := d.execTable(dropPodsTable)
	if err != nil {
		return err
	}

	return nil
}

func (d *databaseInstance) getMeta(page int, perPage int, queryString string) (map[string]int, error) {  
	ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
	defer cancel()

    stmt, err := d.db.PrepareContext(ctx, "select count(*) as totalRecords from (" + queryString + ") tmp")
    if err != nil {
        return nil, err
    }
    defer stmt.Close()

    totalRecords := 0

    err = stmt.QueryRow().Scan(&totalRecords)
    if err != nil {
        return nil, err
    }
    
    totalPages := 0

    if perPage != -1 {
        totalPages = totalRecords/perPage
    } else {
        totalPages = 1
    }


    if totalRecords % perPage > 0 {
        totalPages++
    } 

    meta  := map[string]int { 
        "page":        		page,
        "per_page":    		perPage,
        "totalRowCount":    totalRecords,
        "totalPages": 		totalPages,
    }

    if err != nil {
        return nil, err
    }

    return meta, nil
}


func (d *databaseInstance) GetPods(page int, perPage int) (map[string]interface{}, error) {
	queryString := "select uuid, name, namespace, phase, activeContainers, totalContainers, creationTime, createdBy from pods"
    resultsMap, err := d.genericGet(queryString, page, perPage)
	if err != nil {
		return nil, err
	}
    return resultsMap, nil 
}

func (d *databaseInstance) GetVmis(page int, perPage int) (map[string]interface{}, error) {
	queryString := "select uuid, name, namespace, phase, reason, nodeName, creationTime from vmis"
    resultsMap, err := d.genericGet(queryString, page, perPage)
	if err != nil {
		return nil, err
	}
    return resultsMap, nil 
}

func (d *databaseInstance) GetVmiMigrations(page int, perPage int) (map[string]interface{}, error) {

	queryString := "select name, namespace, uuid, phase, vmiName, targetPod, creationTime, endTimestamp, sourceNode, targetNode, completed, failed from vmimigrations"
    resultsMap, err := d.genericGet(queryString, page, perPage)
	if err != nil {
		return nil, err
	}
    return resultsMap, nil 
}

func (d *databaseInstance) genericGet(queryString string, page int, perPage int) (map[string]interface{}, error) {
	response := map[string]interface{}{}
	ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
	defer cancel()

	limit := " "
    if perPage != -1 {
        limit = " limit " + strconv.Itoa((page - 1) * perPage) + ", " + strconv.Itoa(perPage)  
    }

	stmt, err := d.db.PrepareContext(ctx, queryString + limit)
	if err != nil {
		return response, err
	}
	defer stmt.Close()


	rows, err := stmt.Query() 
	if err != nil {
		return response, err
	}

	defer rows.Close()



    columns, err := rows.Columns()
	if err != nil {
		return response, err
	}
	data     := []map[string]interface{}{}
    count    := len(columns)
    values   := make([]interface{}, count)
    scanArgs := make([]interface{}, count)

    for i := range values {
        scanArgs[i] = &values[i]
    }

    for rows.Next() {
        err := rows.Scan(scanArgs...)
        if err != nil {
			return response, err
		}
		tbRecord := map[string]interface{}{}
        for i, col := range columns {
           v     := values[i]
           b, ok := v.([]byte)
           if (ok) {
               tbRecord[col] = string(b)
           } else {
               tbRecord[col] = v
           }
        }
        data = append(data, tbRecord)

    } 

	meta, err := d.getMeta(page, perPage, queryString)
	if err != nil {
		return nil, err
	}
	response["data"] = data
	response["meta"] = meta
    return response, nil 
}

func (d *databaseInstance) getSingleMigrationByUUID(uuid string) (*VirtualMachineInstanceMigration, error) {

    vmim := VirtualMachineInstanceMigration{}
    var startTime time.Time
    var endTime time.Time
     
    query := fmt.Sprintf("SELECT name, namespace, uuid, phase, vmiName, targetPod, creationTime, endTimestamp, sourceNode, targetNode, completed, failed from vmimigrations WHERE uuid='%s'", uuid)
	rows := d.db.QueryRow(query) 
    var targetNode string
    err := rows.Scan(&vmim.Name, &vmim.Namespace, &vmim.UUID, &vmim.Phase, &vmim.VMIName, &vmim.TargetPod, 
                     &startTime, &endTime, &vmim.SourceNode, &targetNode, &vmim.Completed,
                     &vmim.Failed)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("SingleMigrationByUUID can't find anything with this uuid: ", uuid)
            return &vmim, err
        } else {
            log.Log.Println("SingleMigrationByUUID ERROR: ", err, " for uuid: ", uuid)
            return &vmim, err
        }
    } 
    
    vmim.TargetNode = targetNode
    vmim.UUID = uuid
    var startTimePtr metav1.Time
    var endTimePtr metav1.Time

    
    startTimeStr := []string{startTime.Format("2006-01-02T15:04:05Z07:00")}
    if err := metav1.Convert_Slice_string_To_v1_Time(&startTimeStr, &startTimePtr, nil); err != nil {
        log.Log.Println("SingleMigrationByUUID ERROR: failed to convert time", err, " for uuid: ", uuid)
        return &vmim, err
    }
    endTimeStr := []string{endTime.Format("2006-01-02T15:04:05Z07:00")}
    if err := metav1.Convert_Slice_string_To_v1_Time(&endTimeStr, &endTimePtr, nil); err != nil {
        log.Log.Println("SingleMigrationByUUID ERROR: failed to convert time", err, " for uuid: ", uuid)
        return &vmim, err
    }

    vmim.CreationTime = startTimePtr
    vmim.EndTimestamp = endTimePtr
    log.Log.Println("SingleMigrationByUUID: ", vmim)
    return &vmim, nil
}
