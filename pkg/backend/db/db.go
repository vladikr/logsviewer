package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
    "strconv"
    "strings"

	k8sv1 "k8s.io/api/core/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    "logsviewer/pkg/backend/log"

	_ "github.com/go-sql-driver/mysql"
)

func (d *DatabaseInstance) StorePVC(pvc *PersistentVolumeClaim) error {
	ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctx, insertPVCQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()
	madeAt := pvc.CreationTime.Format("2006-01-02 15:04:05.999999")

	_, err = stmt.ExecContext(
        ctx,
        pvc.Name,
        pvc.Namespace,
        pvc.UUID,
        pvc.Reason,
        pvc.Phase,
        pvc.AccessModes,
        pvc.StorageClassName,
        pvc.VolumeName,
        pvc.VolumeMode,
        pvc.Capacity,
        madeAt,
        pvc.Content)
	if err != nil {
		return err
	}

	return nil
} 


func (d *DatabaseInstance) StorePod(pod *Pod) error {
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

func (d *DatabaseInstance) StoreNode(node *Node) error {
	ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctx, insertNodeQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(
        ctx,
        node.Name,
        node.SystemUUID,
        node.Status,
        node.InternalIP,
        node.HostName,
        node.OsImage,
        node.KernelVersion,
        node.KubletVersion,
        node.ContainerRuntimeVersion,
        node.Content)
	if err != nil {
		return err
	}

	return nil
} 

func (d *DatabaseInstance) StoreVmi(vmi *VirtualMachineInstance) error {
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

func (d *DatabaseInstance) StoreVmiMigration(vmim *VirtualMachineInstanceMigration) error {
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
	insertNodeQuery       = `INSERT INTO nodes(name, systemUuid, status, internalIP, hostName, osImage, kernelVersion, kubletVersion, containerRuntimeVersion, content) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE name=VALUES(name);`
	insertPVCQuery       = `INSERT INTO pvcs(name, namespace, uuid, reason, phase, accessModes, storageClassName, volumeName, volumeMode, capacity, creationTime, content) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE uuid=VALUES(uuid);`
)

var (
	defaultUsername = "mysql"
	defaultPassword = "supersecret"
	//defaultHost     = "mysql"
	defaultHost     = "0.0.0.0"
	defaultPort     = "3306"
	defaultdbName   = "objtracker"
)

type DatabaseInstance struct {
	username string
	password string
	host     string
	port     string
	dbName   string
	db       *sql.DB
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewDatabaseInstance() (*DatabaseInstance, error) {
	dbInstance := &DatabaseInstance{
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

func (d *DatabaseInstance) Shutdown() (err error) {
	if d.cancel != nil {
		d.cancel()
	}

	if d.db != nil {
		d.db.Close()
	}
	return
}

func (d *DatabaseInstance) InitTables() (err error) {
	err = d.createTables()
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseInstance) connect() (err error) {

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

func (d *DatabaseInstance) createTables() error {
    if err := d.createPodsTable(); err != nil {
		return err
	}
    if err := d.createNodesTable(); err != nil {
		return err
	}
    if err := d.createVmisTable(); err != nil {
		return err
	}
    if err := d.createVmiMigrationsTable(); err != nil {
		return err
	}
    if err := d.createPVCsTable(); err != nil {
		return err
	}
	return nil
}

func (d *DatabaseInstance) createPodsTable() error {

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

func (d *DatabaseInstance) createVmisTable() error {

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

func (d *DatabaseInstance) createVmiMigrationsTable() error {

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

func (d *DatabaseInstance) createNodesTable() error {

	createNodesTable := `
	CREATE TABLE IF NOT EXISTS nodes (
	  name varchar(100),
	  systemUuid varchar(100),
	  status varchar(100),
	  internalIP varchar(100),
	  hostName varchar(100),
      osImage varchar(100),
      kernelVersion varchar(100),
      kubletVersion varchar(100),
      containerRuntimeVersion varchar(100),
      content json,
	  PRIMARY KEY (name)
	);
	`
	err := d.execTable(createNodesTable)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseInstance) createPVCsTable() error {
	createPVCsTable := `
	CREATE TABLE IF NOT EXISTS pvcs (
	  name varchar(100),
	  namespace varchar(100),
	  uuid varchar(100),
      reason varchar(100),
      phase varchar(100),
      accessModes varchar(100),
      storageClassName varchar(100),
      volumeName varchar(100),
      volumeMode varchar(100),
      capacity varchar(100),
      creationTime datetime,
      content json,
	  PRIMARY KEY (uuid)
	);
	`
	err := d.execTable(createPVCsTable)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseInstance) execTable(tableSql string) error {
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
func (d *DatabaseInstance) DropTables() error {
	dropPodsTable := `
	DROP TABLE pods;
	`
	err := d.execTable(dropPodsTable)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseInstance) getMeta(page int, perPage int, queryString string) (map[string]int, error) {  
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

func (d *DatabaseInstance) GetMigrationQueryParams(migrationUUID string) (QueryResults, error) {
    // looking for a source pod - pod runs on sourceNode createdBy vmiUUID before migration creationTime and after/equal vmi creation time

    results := QueryResults{}
    migration, err := d.getSingleMigrationByUUID(migrationUUID)
    if err != nil {
        return results, err
    }

    vmiUUID, creationTime, err := d.getVMICreationTimeByName(migration.VMIName, migration.Namespace)
    if err != nil {
        return results, err
    }

    targetPodUUID, err := d.getPodUUIDByName(migration.TargetPod, migration.Namespace)
    if err != nil {
        return results, err
    }
    timeLayout := "2006-01-02 15:04:05.999999"
	vmiMadeAt := creationTime.Format(timeLayout)
	migrationMadeAt := migration.CreationTime.Format(timeLayout)
	migrationEndedAt := migration.EndTimestamp.Format(timeLayout)
    
	sourcePodQueryString := fmt.Sprintf("select uuid, name from pods where createdBy='%s' AND nodeName='%s' AND creationTime BETWEEN '%s' and '%s' ORDER BY creationTime ASC LIMIT 1", vmiUUID, migration.SourceNode, vmiMadeAt, migrationMadeAt)
	virtHandlerQueryString := "select name from pods where nodeName='%s' AND name like 'virt-handler%%'"

    results.StartTimestamp, _ = time.Parse(timeLayout, migrationMadeAt)
    results.EndTimestamp, _ = time.Parse(timeLayout, migrationEndedAt)
    results.TargetPod = migration.TargetPod
    results.TargetPodUUID = targetPodUUID
    results.MigrationUUID = migrationUUID
    results.VMIUUID = vmiUUID

    // get source virt-launcher info
	rows := d.db.QueryRow(sourcePodQueryString) 
    err = rows.Scan(&results.SourcePodUUID, &results.SourcePod)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("migration source pod lookup - can't find anything with this uuid: ", vmiUUID)
            return results, err
        } else {
            log.Log.Println("migration source pod lookup, ERROR: ", err, " for uuid: ", vmiUUID)
            return results, err
        }
    } 
    
    // get the source virt-handler
	rows = d.db.QueryRow(fmt.Sprintf(virtHandlerQueryString, migration.SourceNode))
    err = rows.Scan(&results.SourceHandler)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("migration src virt-handler lookup -  can't find virt-handler on node: ", migration.SourceNode)
            return results, err
        } else {
            log.Log.Println("migration src virt-handler lookup - ERROR: ", err, " for nodeName: ", migration.SourceNode)
            return results, err
        }
    } 

    // get the target virt-handler
	rows = d.db.QueryRow(fmt.Sprintf(virtHandlerQueryString, migration.TargetNode))
    err = rows.Scan(&results.TargetHandler)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("migration target virt-handler lookup -  can't find virt-handler on node: ", migration.TargetNode)
            return results, err
        } else {
            log.Log.Println("migration target virt-handler lookup - ERROR: ", err, " for nodeName: ", migration.TargetNode)
            return results, err
        }
    } 
    return results, nil 
}


func (d *DatabaseInstance) GetVMIQueryParams(vmiUUID string, nodeName string) (QueryResults, error) {
    results := QueryResults{VMIUUID: vmiUUID}
 
	sourcePodQueryString := fmt.Sprintf("select uuid, name, namespace, creationTime from pods where createdBy='%s' AND nodeName='%s'", vmiUUID, nodeName)
	virtHandlerQueryString := fmt.Sprintf("select name from pods where nodeName='%s' AND name like 'virt-handler%%'", nodeName)


    // get source virt-launcher info
	rows := d.db.QueryRow(sourcePodQueryString) 
    err := rows.Scan(&results.SourcePodUUID, &results.SourcePod, &results.Namespace, &results.StartTimestamp)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("getVMIQueryParams can't find anything with this uuid: ", vmiUUID)
            return results, err
        } else {
            log.Log.Println("getVMIQueryParams ERROR: ", err, " for uuid: ", vmiUUID)
            return results, err
        }
    } 
    
    // get the relevant virt-handler
	rows = d.db.QueryRow(virtHandlerQueryString) 
    err = rows.Scan(&results.SourceHandler)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("getVMIQueryParams can't find virt-handler on node: ", nodeName)
            return results, err
        } else {
            log.Log.Println("getVMIQueryParams ERROR: ", err, " for nodeName: ", nodeName)
            return results, err
        }
    } 

    return results, nil 
}

func (d *DatabaseInstance) GetPodQueryParams(podUUID string) (QueryResults, error) {
    results := QueryResults{}
 
	sourcePodQueryString := fmt.Sprintf("select uuid, name, namespace, creationTime from pods where uuid='%s'", podUUID)


    // get source virt-launcher info
	rows := d.db.QueryRow(sourcePodQueryString) 
    err := rows.Scan(&results.SourcePodUUID, &results.SourcePod, &results.Namespace, &results.StartTimestamp)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("getPodQueryParams can't find anything with this uuid: ", podUUID)
            return results, err
        } else {
            log.Log.Println("getPodQueryParams ERROR: ", err, " for uuid: ", podUUID)
            return results, err
        }
    } 
    

    return results, nil 
}

func (d *DatabaseInstance) GetPods(page int, perPage int, queryDetails *GenericQueryDetails) (map[string]interface{}, error) {
	queryString := "select uuid, name, namespace, phase, activeContainers, totalContainers, creationTime, createdBy from pods"
    if queryDetails != nil {
        conditions := []string{}
        if (*queryDetails != (GenericQueryDetails{})) {
            queryString = fmt.Sprintf("%s where ", queryString)
        }

        if  queryDetails.Name != "" {
            conditions = append(conditions, fmt.Sprintf("name='%s'", queryDetails.Name))
        }
        if  queryDetails.Namespace != "" {
            conditions = append(conditions, fmt.Sprintf("namespace='%s'", queryDetails.Namespace))
        }
        if  queryDetails.UUID != "" {
            conditions = append(conditions, fmt.Sprintf("uuid='%s'", queryDetails.UUID))
        }
        
        queryString = fmt.Sprintf("%s%s", queryString, strings.Join(conditions, " AND "))
    }
    
    log.Log.Println("queryString: ", queryString)
    resultsMap, err := d.genericGet(queryString, page, perPage)
	if err != nil {
		return nil, err
	}
    return resultsMap, nil 
}

func (d *DatabaseInstance) GetNodeObject(nodeUUID string) (*k8sv1.Node, error) {
    var content json.RawMessage
 
	queryString := fmt.Sprintf("select content from nodes where systemUuid = '%s'", nodeUUID)

	rows := d.db.QueryRow(queryString) 
    err := rows.Scan(&content)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("can't find nodes with this uuid: ", nodeUUID)
            return nil, err
        } else {
            log.Log.Println("ERROR: ", err, " for uuid: ", nodeUUID)
            return nil, err
        }
    }

    
    // Unmashal json
    var node k8sv1.Node

    err = json.Unmarshal(content, &node)
    if err != nil {
      log.Log.Fatalln("failed to unmarshal json to node object - ", err)
      return nil, err
    }
 
    return &node, nil 

}

func (d *DatabaseInstance) GetPVCs(page int, perPage int, queryDetails *GenericQueryDetails) (map[string]interface{}, error) {
	queryString := "select name, namespace, uuid, reason, phase, accessModes, storageClassName, volumeName, volumeMode, capacity, creationTime from pvcs"
    if queryDetails != nil {
        conditions := []string{}
        if (*queryDetails != (GenericQueryDetails{})) {
            queryString = fmt.Sprintf("%s where ", queryString)
        }

        if  queryDetails.Name != "" {
            conditions = append(conditions, fmt.Sprintf("name='%s'", queryDetails.Name))
        }
        if  queryDetails.Namespace != "" {
            conditions = append(conditions, fmt.Sprintf("namespace='%s'", queryDetails.Namespace))
        }
        if  queryDetails.UUID != "" {
            conditions = append(conditions, fmt.Sprintf("uuid='%s'", queryDetails.UUID))
        }
        
        queryString = fmt.Sprintf("%s%s", queryString, strings.Join(conditions, " AND "))
    }
    
    log.Log.Println("queryString: ", queryString)
    resultsMap, err := d.genericGet(queryString, page, perPage)
	if err != nil {
		return nil, err
	}
    return resultsMap, nil 
}

func (d *DatabaseInstance) GetVMIObject(vmiUUID string) (*kubevirtv1.VirtualMachineInstance, error) {
    var content json.RawMessage
 
	queryString := fmt.Sprintf("select content from vmis where uuid = '%s'", vmiUUID)

	rows := d.db.QueryRow(queryString) 
    err := rows.Scan(&content)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("can't find VMIs with this uuid: ", vmiUUID)
            return nil, err
        } else {
            log.Log.Println("ERROR: ", err, " for uuid: ", vmiUUID)
            return nil, err
        }
    }

    
    // Unmashal json
    var vmi kubevirtv1.VirtualMachineInstance

    err = json.Unmarshal(content, &vmi)
    if err != nil {
      log.Log.Fatalln("failed to unmarshal json to vmi object - ", err)
      return nil, err
    }
 
    return &vmi, nil 

}

func (d *DatabaseInstance) GetPodObject(podUUID string) (*k8sv1.Pod, error) {
    var content json.RawMessage
 
	queryString := fmt.Sprintf("select content from pods where uuid = '%s'", podUUID)

    // get source virt-launcher info
	rows := d.db.QueryRow(queryString) 
    err := rows.Scan(&content)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("getPodQueryParams can't find anything with this uuid: ", podUUID)
            return nil, err
        } else {
            log.Log.Println("getPodQueryParams ERROR: ", err, " for uuid: ", podUUID)
            return nil, err
        }
    }

    
    // Unmashal json
    var pod k8sv1.Pod

    err = json.Unmarshal(content, &pod)
    if err != nil {
      log.Log.Fatalln("failed to unmarshal json to pod - ", err)
      return nil, err
    }
 
    return &pod, nil 

}

func (d *DatabaseInstance) GetNodes(page int, perPage int) (map[string]interface{}, error) {
	queryString := "select name, systemUuid, status, internalIP, hostName, osImage, kernelVersion, kubletVersion, containerRuntimeVersion from nodes"
    resultsMap, err := d.genericGet(queryString, page, perPage)
	if err != nil {
		return nil, err
	}
    return resultsMap, nil 
}

func (d *DatabaseInstance) GetVmis(page int, perPage int) (map[string]interface{}, error) {
	queryString := "select uuid, name, namespace, phase, reason, nodeName, creationTime from vmis"
    resultsMap, err := d.genericGet(queryString, page, perPage)
	if err != nil {
		return nil, err
	}
    return resultsMap, nil 
}

func (d *DatabaseInstance) GetVmiMigrations(page int, perPage int, vmiDetails *GenericQueryDetails) (map[string]interface{}, error) {

	queryString := "select name, namespace, uuid, phase, vmiName, targetPod, creationTime, endTimestamp, sourceNode, targetNode, completed, failed from vmimigrations"

    if vmiDetails != nil {
        conditions := []string{}
        if (*vmiDetails != (GenericQueryDetails{})) {
            queryString = fmt.Sprintf("%s where ", queryString)
        }

        if  vmiDetails.Name != "" {
            conditions = append(conditions, fmt.Sprintf("vmiName='%s'", vmiDetails.Name))
        }
        if  vmiDetails.Namespace != "" {
            conditions = append(conditions, fmt.Sprintf("namespace='%s'", vmiDetails.Namespace))
        }
        if  vmiDetails.UUID != "" {
            conditions = append(conditions, fmt.Sprintf("uuid='%s'", vmiDetails.UUID))
        }
        
        queryString = fmt.Sprintf("%s%s", queryString, strings.Join(conditions, " AND "))
    }
    
    log.Log.Println("queryString: ", queryString)
    resultsMap, err := d.genericGet(queryString, page, perPage)
	if err != nil {
		return nil, err
	}
    return resultsMap, nil 
}

func (d *DatabaseInstance) genericGet(queryString string, page int, perPage int) (map[string]interface{}, error) {
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

func (d *DatabaseInstance) getPodUUIDByName(name string, namespace string) (string, error) {

    var podUUID string
    query := fmt.Sprintf("SELECT uuid from pods WHERE name='%s' AND namespace='%s'", name, namespace)
	rows := d.db.QueryRow(query) 

    err := rows.Scan(&podUUID)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("failed find a pod with key: ", fmt.Sprintf("%s/%s",name, namespace))
            return podUUID, err
        } else {
            log.Log.Println("ERROR: ", err, " for pod key: ", fmt.Sprintf("%s/%s",name, namespace))
            return podUUID, err
        }
    } 
    
    return podUUID, nil
}

func (d *DatabaseInstance) getVMICreationTimeByName(name string, namespace string) (string, time.Time, error) {

    var creationTime time.Time
    var vmiUUID string
    query := fmt.Sprintf("SELECT uuid, creationTime from vmis WHERE name='%s' AND namespace='%s'", name, namespace)
	rows := d.db.QueryRow(query) 

    err := rows.Scan(&vmiUUID, &creationTime)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("failed find VMI with key: ", fmt.Sprintf("%s/%s",name, namespace))
            return vmiUUID, creationTime, err
        } else {
            log.Log.Println("ERROR: ", err, " for vmi key: ", fmt.Sprintf("%s/%s",name, namespace))
            return vmiUUID, creationTime, err
        }
    } 
    
    return vmiUUID, creationTime, nil
}

func (d *DatabaseInstance) getSingleMigrationByUUID(uuid string) (*VirtualMachineInstanceMigration, error) {

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
