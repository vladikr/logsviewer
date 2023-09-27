package db

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "strconv"
    "strings"
    "time"

    k8sv1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    kubevirtv1 "kubevirt.io/api/core/v1"

    "logsviewer/pkg/backend/log"

    _ "github.com/go-sql-driver/mysql"
)

func (d *DatabaseInstance) StoreSubscription(sub *Subscription) error {
    ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
    defer cancel()

    stmt, err := d.db.PrepareContext(ctx, insertSubscriptionQuery)
    if err != nil {
        return err
    }
    defer stmt.Close()
    madeAt := sub.CreationTime.Format("2006-01-02 15:04:05.999999")

    _, err = stmt.ExecContext(
        ctx,
        sub.Name,
        sub.Namespace,
        sub.UUID,
        sub.Source,
        sub.SourceNamespace,
        sub.StartingCSV,
        sub.CurrentCSV,
        sub.InstalledCSV,
        sub.State,
        madeAt,
        sub.Content,
    )
    if err != nil {
        return err
    }

    return nil
}

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
        pod.PVCs,
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

func BoolToString(value bool) string {
    if value {
        return "true"
    }
    return "false"
}

func (d *DatabaseInstance) StoreVm(vm *VirtualMachine) error {
    ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
    defer cancel()

    stmt, err := d.db.PrepareContext(ctx, insertVmQuery)
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.ExecContext(
        ctx,
        vm.Name,
        vm.Namespace,
        vm.UUID,
        BoolToString(vm.Running),
        BoolToString(vm.Created),
        BoolToString(vm.Ready),
        vm.Status,
        vm.Content)
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
                Name:         existngVmim.Name,
                Namespace:    existngVmim.Namespace,
                UUID:         string(migrationState.MigrationUID),
                Phase:        string(existngVmim.Phase),
                VMIName:      string(existngVmim.VMIName),
                TargetPod:    migrationState.TargetPod,
                CreationTime: *migrationState.StartTimestamp,
                EndTimestamp: *migrationState.EndTimestamp,
                SourceNode:   migrationState.SourceNode,
                TargetNode:   migrationState.TargetNode,
                Completed:    migrationState.Completed,
                Failed:       migrationState.Failed,
                Content:      emptyContent}

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

func (d *DatabaseInstance) StoreImportedMustGather(img *ImportedMustGather) error {
    ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
    defer cancel()

    stmt, err := d.db.PrepareContext(ctx, insertImportedMustGatherQuery)
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.ExecContext(
        ctx,
        img.Name,
        img.ImportTime.Format("2006-01-02 15:04:05.999999"),
        img.GatherTime.Format("2006-01-02 15:04:05.999999"),
    )
    if err != nil {
        return err
    }

    return nil
}

var (
    insertPodQuery                = `INSERT INTO pods(keyid, kind, name, namespace, uuid, phase, activeContainers, totalContainers, nodeName, creationTime, pvcs, content, createdBy) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE keyid=VALUES(keyid);`
    insertVmQuery                = `INSERT INTO vms(name, namespace, uuid, running, created, ready, status, content) values (?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE uuid=VALUES(uuid);`
    insertVmiQuery                = `INSERT INTO vmis(name, namespace, uuid, reason, phase, nodeName, creationTime, content) values (?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE uuid=VALUES(uuid);`
    insertVmiMigrationQuery       = `INSERT INTO vmimigrations(name, namespace, uuid, phase, vmiName, targetPod, creationTime, endTimestamp, sourceNode, targetNode, completed, failed, content) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE uuid=VALUES(uuid), targetPod=VALUES(targetPod), creationTime=VALUES(creationTime), endTimestamp=VALUES(endTimestamp), sourceNode=VALUES(sourceNode), targetNode=VALUES(targetNode), completed=VALUES(completed), failed=VALUES(failed);`
    insertNodeQuery               = `INSERT INTO nodes(name, systemUuid, status, internalIP, hostName, osImage, kernelVersion, kubletVersion, containerRuntimeVersion, content) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE name=VALUES(name);`
    insertPVCQuery                = `INSERT INTO pvcs(name, namespace, uuid, reason, phase, accessModes, storageClassName, volumeName, volumeMode, capacity, creationTime, content) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE uuid=VALUES(uuid);`
    insertSubscriptionQuery       = `INSERT INTO subscriptions(name, namespace, uuid, source, sourceNamespace, startingCSV, currentCSV, installedCSV, state, creationTime, content) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE uuid=VALUES(uuid);`
    insertImportedMustGatherQuery = `INSERT INTO importedmustgathers(name, importTime, gatherTime) values (?, ?, ?);`
)

var (
    defaultUsername = "mysql"
    defaultPassword = "supersecret"
    //defaultHost     = "mysql"
    defaultHost   = "0.0.0.0"
    defaultPort   = "3306"
    defaultdbName = "objtracker"
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
    if err := d.createVmsTable(); err != nil {
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
    if err := d.createSubscriptionsTable(); err != nil {
        return err
    }
    if err := d.createImportedMustGathersTable(); err != nil {
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
      pvcs varchar(200),
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

func (d *DatabaseInstance) createVmsTable() error {

    vmsTableCreate := `
    CREATE TABLE IF NOT EXISTS vms (
      name varchar(100),
      namespace varchar(100),
      uuid varchar(100),
      running varchar(100),
      created varchar(100),
      ready varchar(100),
      status varchar(100),
      content json,
      PRIMARY KEY (uuid)
    );
    `
    err := d.execTable(vmsTableCreate)
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

func (d *DatabaseInstance) createSubscriptionsTable() error {
    createSubscriptionsTable := `
    CREATE TABLE IF NOT EXISTS subscriptions (
      name varchar(100),
      namespace varchar(100),
      uuid varchar(100),
      source varchar(100),
      sourceNamespace varchar(100),
      startingCSV varchar(100),
      currentCSV varchar(100),
      installedCSV varchar(100),
      state varchar(100),
      creationTime datetime,
      content json,
      PRIMARY KEY (uuid)
    );
    `
    err := d.execTable(createSubscriptionsTable)
    if err != nil {
        return err
    }

    return nil
}

func (d *DatabaseInstance) createImportedMustGathersTable() error {
    createImportedMustGathersTable := `
    CREATE TABLE IF NOT EXISTS importedmustgathers (
      name varchar(200),
      importTime datetime,
      gatherTime datetime,
      id int(16) auto_increment, 
      PRIMARY KEY (id)
    );
    `
    err := d.execTable(createImportedMustGathersTable)
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

    stmt, err := d.db.PrepareContext(ctx, "select count(*) as totalRecords from ("+queryString+") tmp")
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
        totalPages = totalRecords / perPage
    } else {
        totalPages = 1
    }

    if totalRecords%perPage > 0 {
        totalPages++
    }

    meta := map[string]int{
        "page":          page,
        "per_page":      perPage,
        "totalRowCount": totalRecords,
        "totalPages":    totalPages,
    }

    if err != nil {
        return nil, err
    }

    return meta, nil
}

func (d *DatabaseInstance) GetResourceStats() (map[string]interface{}, error) {
    queryString := `
-- For pods
SELECT 'pods', phase AS status, COUNT(*) AS count FROM pods GROUP BY phase
-- For nodes
UNION ALL
SELECT 'nodes', status AS status, COUNT(*) AS count FROM nodes GROUP BY status
-- For vmis
UNION ALL
SELECT 'vmis', phase AS status, COUNT(*) AS count FROM vmis GROUP BY phase
-- For vmimigrations
UNION ALL
SELECT 'vmimigrations', phase AS status, COUNT(*) AS count FROM vmimigrations GROUP BY phase
-- For pvcs
UNION ALL
SELECT 'pvcs', phase AS status, COUNT(*) AS count FROM pvcs GROUP BY phase
`

    log.Log.Println("queryString: ", queryString)

    rows, err := d.db.Query(queryString)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    resultsMap := map[string]interface{}{
        "pods":          map[string]int{},
        "nodes":         map[string]int{},
        "vmis":          map[string]int{},
        "vmimigrations": map[string]int{},
        "pvcs":          map[string]int{},
    }

    for rows.Next() {
        var resourceType string
        var status string
        var count int

        err = rows.Scan(&resourceType, &status, &count)
        if err != nil {
            return nil, err
        }

        resultsMap[resourceType].(map[string]int)[status] = count
    }

    return resultsMap, nil
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

    // get a list of PVCs
    pvcsListRes, err := d.GetVMIPVCs(vmiUUID)
    if err != nil {
        return results, err
    }
    data, ok := pvcsListRes["data"].([]map[string]interface{})
    if !ok {
        log.Log.Println("failed to covert pvc data")
        return results, nil
    }

    for _, pvc := range data {
        results.PVCs = append(results.PVCs, fmt.Sprintf("%v", pvc["uuid"]))
    }

    return results, nil
}

func (d *DatabaseInstance) GetFullVMIHistoryQueryParams(vmiUUID string) (QueryResults, error) {
    results := QueryResults{VMIUUID: vmiUUID}
    vmiQueryString := fmt.Sprintf("select creationTime from vmis where uuid = '%s'", vmiUUID)
    sourcePodQueryString := fmt.Sprintf("select uuid, name, nodeName from pods where createdBy='%s'", vmiUUID)
    virtHandlerQueryString := "select name from pods where nodeName IN (%s) AND name like 'virt-handler%%'"

    // get VMIs creation time
    rows := d.db.QueryRow(vmiQueryString)
    if err := rows.Scan(&results.StartTimestamp); err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("getVMIQueryParams can't find any VMIs with this uuid: ", vmiUUID)
            return results, err
        } else {
            log.Log.Println("getVMIQueryParams ERROR: ", err, " for VMI uuid: ", vmiUUID)
            return results, err
        }
    }

    // get all involved virt-launchers
    podsResultsMap, err := d.genericGet(sourcePodQueryString, -1, -1)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("failed to get pods data for vmi uuid: ", vmiUUID)
            return results, err
        } else {
            log.Log.Println("failed to get pods data ERROR: ", err, " for uuid: ", vmiUUID)
            return results, err
        }
    }

    // list of nodes on which the relevant virt-handlers reside
    nodesList := []string{}
    if data, ok := podsResultsMap["data"].([]map[string]interface{}); ok {
        for _, launcher := range data {
            newPod := Pod{
                UUID: fmt.Sprintf("%v", launcher["uuid"]),
                Name: fmt.Sprintf("%v", launcher["name"]),
                NodeName: fmt.Sprintf("%v", launcher["nodeName"]),
            }
            nodesList = append(nodesList, fmt.Sprintf("%v", launcher["nodeName"]))
            results.InvolvedVirtLaunchers = append(results.InvolvedVirtLaunchers, newPod)
        }
    } else {
        log.Log.Println("failed to covert virt-launcher pods data")
        return results, nil
    }

    nodesByte := []byte(fmt.Sprintf(`'%s'`, strings.Join(nodesList, `', '`)))
    virtHandlerQueryString = fmt.Sprintf(virtHandlerQueryString, string(nodesByte))

    // get all involved virt-handlers
    handlersResultsMap, err := d.genericGet(virtHandlerQueryString, -1, -1)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("failed to get virt-handler pods data for vmi uuid: ", vmiUUID)
            return results, err
        } else {
            log.Log.Println("failed to get virt-handler pods data ERROR: ", err, " for uuid: ", vmiUUID)
            return results, err
        }
    }

    if data, ok := handlersResultsMap["data"].([]map[string]interface{}); ok {
        for _, launcher := range data {
            newPod := Pod{
                UUID: fmt.Sprintf("%v", launcher["uuid"]),
                Name: fmt.Sprintf("%v", launcher["name"]),
                NodeName: fmt.Sprintf("%v", launcher["nodeName"]),
            }
            results.InvolvedVirtHandlers = append(results.InvolvedVirtHandlers, newPod)
        }
    } else {
        log.Log.Println("failed to covert virt-handler pods data")
        return results, nil
    }

    // get a list of PVCs
    pvcsListRes, err := d.GetVMIPVCs(vmiUUID)
    if err != nil {
        return results, err
    }

    if data, ok := pvcsListRes["data"].([]map[string]interface{}); ok {

        for _, pvc := range data {
            results.PVCs = append(results.PVCs, fmt.Sprintf("%v", pvc["uuid"]))
        }
    } else {
        log.Log.Println("failed to covert pvc data")
        return results, nil
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

    // get a list of PVCs
    pvcsListRes, err := d.GetVMIPVCs(vmiUUID)
    if err != nil {
        return results, err
    }

    data, ok := pvcsListRes["data"].([]map[string]interface{})
    if !ok {
        log.Log.Println("failed to covert pvc data")
        return results, nil
    }

    for _, pvc := range data {
        results.PVCs = append(results.PVCs, fmt.Sprintf("%v", pvc["uuid"]))
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

    // get a list of PVCs
    pvcsListRes, err := d.GetPodPVCs(podUUID)
    if err != nil {
        return results, err
    }

    data, ok := pvcsListRes["data"].([]map[string]interface{})
    if !ok {
        log.Log.Println("failed to covert pvc data")
        return results, nil
    }
    for _, pvc := range data {
        results.PVCs = append(results.PVCs, fmt.Sprintf("%v", pvc["uuid"]))
    }

    return results, nil
}

func (d *DatabaseInstance) GetPods(page int, perPage int, queryDetails *GenericQueryDetails) (map[string]interface{}, error) {
    queryString := "select uuid, name, namespace, phase, activeContainers, totalContainers, creationTime, createdBy from pods"
    if queryDetails != nil {
        conditions := []string{}
        if *queryDetails != (GenericQueryDetails{}) {
            queryString = fmt.Sprintf("%s where ", queryString)
        }

        if queryDetails.Name != "" {
            conditions = append(conditions, fmt.Sprintf("name='%s'", queryDetails.Name))
        }
        if queryDetails.Namespace != "" {
            conditions = append(conditions, fmt.Sprintf("namespace='%s'", queryDetails.Namespace))
        }
        if queryDetails.UUID != "" {
            conditions = append(conditions, fmt.Sprintf("uuid='%s'", queryDetails.UUID))
        }
        if queryDetails.Status != "" {
            if queryDetails.Status == "healthy" {
                // Running or Succeeded
                conditions = append(conditions, fmt.Sprintf("phase='Running' OR phase='Succeeded'"))
            } else if queryDetails.Status == "unhealthy" {
                // Failed
                conditions = append(conditions, fmt.Sprintf("phase='Failed'"))
            } else if queryDetails.Status == "warning" {
                // other thank Running, Succeeded, Failed
                conditions = append(conditions, fmt.Sprintf("phase!='Running' AND phase!='Succeeded' AND phase!='Failed'"))
            }
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

// GetPVCObject returns a PersistentVolumeClaim yaml object
func (d *DatabaseInstance) GetPVCObject(pvcUUID string) (*k8sv1.PersistentVolumeClaim, error) {
    var content json.RawMessage

    queryString := fmt.Sprintf("select content from pvcs where uuid = '%s'", pvcUUID)

    rows := d.db.QueryRow(queryString)
    err := rows.Scan(&content)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("can't find pvcs with this uuid: ", pvcUUID)
            return nil, err
        } else {
            log.Log.Println("ERROR: ", err, " for uuid: ", pvcUUID)
            return nil, err
        }
    }

    // Unmashal json
    var pvc k8sv1.PersistentVolumeClaim

    err = json.Unmarshal(content, &pvc)
    if err != nil {
        log.Log.Fatalln("failed to unmarshal json to pvc object - ", err)
        return nil, err
    }

    return &pvc, nil

}

func (d *DatabaseInstance) GetPVCs(page int, perPage int, queryDetails *GenericQueryDetails) (map[string]interface{}, error) {
    queryString := "select name, namespace, uuid, reason, phase, accessModes, storageClassName, volumeName, volumeMode, capacity, creationTime from pvcs"
    if queryDetails != nil {
        conditions := []string{}
        if *queryDetails != (GenericQueryDetails{}) {
            queryString = fmt.Sprintf("%s where ", queryString)
        }

        if queryDetails.Name != "" {
            // handle list of claim names
            claimsList := strings.Split(queryDetails.Name, ",")
            if len(claimsList) > 1 {
                queryStr := "name in ("
                for idx, claim := range claimsList {
                    queryStr += fmt.Sprintf("'%s'", claim)
                    if idx != len(claimsList)-1 {
                        queryStr += ", "
                    }
                }
                queryStr += ")"
                conditions = append(conditions, queryStr)
            } else {
                conditions = append(conditions, fmt.Sprintf("name='%s'", queryDetails.Name))
            }
        }
        if queryDetails.Namespace != "" {
            conditions = append(conditions, fmt.Sprintf("namespace='%s'", queryDetails.Namespace))
        }
        if queryDetails.UUID != "" {
            conditions = append(conditions, fmt.Sprintf("uuid='%s'", queryDetails.UUID))
        }
        if queryDetails.Status != "" {
            if queryDetails.Status == "healthy" {
                // Bound
                conditions = append(conditions, fmt.Sprintf("phase='Bound'"))
            } else if queryDetails.Status == "unhealthy" {
                // Lost
                conditions = append(conditions, fmt.Sprintf("phase='Lost'"))
            } else if queryDetails.Status == "warning" {
                // other thank Bound, Lost
                conditions = append(conditions, fmt.Sprintf("phase!='Bound' AND phase!='Lost'"))
            }
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

func (d *DatabaseInstance) GetVMIPVCs(vmiUUID string) (map[string]interface{}, error) {
    podQueryString := fmt.Sprintf("select pvcs from pods where createdBy='%s'", vmiUUID)
    res, err := d.genericGetPVCs(podQueryString)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("GeVMIPVCs can't find anything with this uuid: ", vmiUUID)
            return nil, err
        } else {
            log.Log.Println("GetVMIPVCs ERROR: ", err, " for uuid: ", vmiUUID)
            return nil, err
        }
    }
    return res, nil
}

func (d *DatabaseInstance) GetPodPVCs(podUUID string) (map[string]interface{}, error) {
    queryString := fmt.Sprintf("select pvcs from pods where uuid = '%s'", podUUID)
    res, err := d.genericGetPVCs(queryString)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Log.Println("GetPodPVCs can't find anything with this uuid: ", podUUID)
            return nil, err
        } else {
            log.Log.Println("GetPodPVCs ERROR: ", err, " for uuid: ", podUUID)
            return nil, err
        }
    }
    return res, nil
}

func (d *DatabaseInstance) genericGetPVCs(queryString string) (map[string]interface{}, error) {
    podPvcs := ""
    pvcsList := make(map[string]interface{})

    // get pod pvcs
    rows := d.db.QueryRow(queryString)
    err := rows.Scan(&podPvcs)
    if err != nil {
        return nil, err
    }
    if podPvcs != "" {
        queryDetails := GenericQueryDetails{
            Name: podPvcs,
        }
        pvcsList, err = d.GetPVCs(-1, -1, &queryDetails)
        if err != nil {
            return nil, err
        }
    } else {
        data := []map[string]interface{}{}
        meta := map[string]int{
            "page":          -1,
            "per_page":      -1,
            "totalRowCount": 0,
            "totalPages":    1,
        }
        pvcsList["data"] = data
        pvcsList["meta"] = meta
    }
    return pvcsList, nil
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

func (d *DatabaseInstance) GetNodes(page int, perPage int, queryDetails *GenericQueryDetails) (map[string]interface{}, error) {
    queryString := "select name, systemUuid, status, internalIP, hostName, osImage, kernelVersion, kubletVersion, containerRuntimeVersion from nodes"

    if queryDetails != nil {
        conditions := []string{}
        if *queryDetails != (GenericQueryDetails{}) {
            queryString = fmt.Sprintf("%s where ", queryString)
        }

        if queryDetails.Name != "" {
            conditions = append(conditions, fmt.Sprintf("name='%s'", queryDetails.Name))
        }
        if queryDetails.Namespace != "" {
            conditions = append(conditions, fmt.Sprintf("namespace='%s'", queryDetails.Namespace))
        }
        if queryDetails.UUID != "" {
            conditions = append(conditions, fmt.Sprintf("uuid='%s'", queryDetails.UUID))
        }
        if queryDetails.Status != "" {
            if queryDetails.Status == "healthy" {
                conditions = append(conditions, fmt.Sprintf("status='%s'", "Ready"))
            } else if queryDetails.Status == "unhealthy" {
                conditions = append(conditions, fmt.Sprintf("status='%s'", "NotReady"))
            }
        }

        queryString = fmt.Sprintf("%s%s", queryString, strings.Join(conditions, " AND "))
    }

    resultsMap, err := d.genericGet(queryString, page, perPage)
    if err != nil {
        return nil, err
    }
    return resultsMap, nil
}

func (d *DatabaseInstance) GetVms(page int, perPage int, queryDetails *GenericQueryDetails) (map[string]interface{}, error) {
    queryString := "select uuid, name, namespace, running, created, ready, status from vms"

    if queryDetails != nil {
        conditions := []string{}
        if *queryDetails != (GenericQueryDetails{}) {
            queryString = fmt.Sprintf("%s where ", queryString)
        }

        queryString = fmt.Sprintf("%s%s", queryString, strings.Join(conditions, " AND "))
    }

    resultsMap, err := d.genericGet(queryString, page, perPage)
    if err != nil {
        return nil, err
    }
    return resultsMap, nil
}


func (d *DatabaseInstance) GetVmis(page int, perPage int, queryDetails *GenericQueryDetails) (map[string]interface{}, error) {
    queryString := "select uuid, name, namespace, phase, reason, nodeName, creationTime from vmis"

    if queryDetails != nil {
        conditions := []string{}
        if *queryDetails != (GenericQueryDetails{}) {
            queryString = fmt.Sprintf("%s where ", queryString)
        }

        if queryDetails.Status != "" {
            if queryDetails.Status == "healthy" {
                // Running or Succeeded
                conditions = append(conditions, fmt.Sprintf("phase='%s' OR phase='%s'", "Running", "Succeeded"))
            } else if queryDetails.Status == "unhealthy" {
                // Failed
                conditions = append(conditions, fmt.Sprintf("phase='%s'", "Failed"))
            } else if queryDetails.Status == "warning" {
                // Other than Running, Succeeded, Failed
                conditions = append(conditions, fmt.Sprintf("phase!='%s' AND phase!='%s' AND phase!='%s'", "Running", "Succeeded", "Failed"))
            }
        }

        queryString = fmt.Sprintf("%s%s", queryString, strings.Join(conditions, " AND "))
    }

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
        if *vmiDetails != (GenericQueryDetails{}) {
            queryString = fmt.Sprintf("%s where ", queryString)
        }

        if vmiDetails.Name != "" {
            conditions = append(conditions, fmt.Sprintf("vmiName='%s'", vmiDetails.Name))
        }
        if vmiDetails.Namespace != "" {
            conditions = append(conditions, fmt.Sprintf("namespace='%s'", vmiDetails.Namespace))
        }
        if vmiDetails.UUID != "" {
            conditions = append(conditions, fmt.Sprintf("uuid='%s'", vmiDetails.UUID))
        }
        if vmiDetails.Status != "" {
            if vmiDetails.Status == "healthy" {
                // Running or Succeeded
                conditions = append(conditions, fmt.Sprintf("phase='%s' OR phase='%s'", "Running", "Succeeded"))
            } else if vmiDetails.Status == "unhealthy" {
                // Failed
                conditions = append(conditions, fmt.Sprintf("phase='%s'", "Failed"))
            } else if vmiDetails.Status == "warning" {
                // Other than Running, Succeeded, Failed
                conditions = append(conditions, fmt.Sprintf("phase!='%s' AND phase!='%s' AND phase!='%s'", "Running", "Succeeded", "Failed"))
            }
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

func (d *DatabaseInstance) ListImportedMustGather() (imgList []ImportedMustGather, err error) {
    imgList = []ImportedMustGather{}

    queryString := "select name, importTime, gatherTime from importedmustgathers"

    rows, err := d.db.Query(queryString)
    if err != nil {
        log.Log.Fatalln("failed to query imported must gathers - ", err)
        return
    }

    for rows.Next() {
        img := ImportedMustGather{}
        err = rows.Scan(&img.Name, &img.ImportTime, &img.GatherTime)
        if err != nil {
            log.Log.Fatalln("failed to scan imported must gather - ", err)
            return
        }
        imgList = append(imgList, img)
    }
    return
}

func (d *DatabaseInstance) GetImportedMustGather(name string) (img *ImportedMustGather, exists bool, err error) {
    img = &ImportedMustGather{}

    queryString := fmt.Sprintf("select name, importTime, gatherTime from importedmustgathers where name = '%s' limit 1", name)

    rows := d.db.QueryRow(queryString)
    err = rows.Scan(&img.Name, &img.ImportTime, &img.GatherTime)
    if err != nil {
        exists = false
        if err == sql.ErrNoRows {
            log.Log.Println("No imported must gather found with name: ", name)
            err = nil
            return
        } else {
            log.Log.Fatalln("failed to query imported must gather - ", err)
            return
        }
    }
    exists = true
    return
}

func (d *DatabaseInstance) genericGet(queryString string, page int, perPage int) (map[string]interface{}, error) {
    response := map[string]interface{}{}
    ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
    defer cancel()

    limit := " "
    if perPage != -1 {
        limit = " limit " + strconv.Itoa((page-1)*perPage) + ", " + strconv.Itoa(perPage)
    }

    log.Log.Println("queryString: ", queryString+limit)
    stmt, err := d.db.PrepareContext(ctx, queryString+limit)
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
    data := []map[string]interface{}{}
    count := len(columns)
    values := make([]interface{}, count)
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
            v := values[i]
            b, ok := v.([]byte)
            if ok {
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
            log.Log.Println("failed find a pod with key: ", fmt.Sprintf("%s/%s", name, namespace))
            return podUUID, err
        } else {
            log.Log.Println("ERROR: ", err, " for pod key: ", fmt.Sprintf("%s/%s", name, namespace))
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
            log.Log.Println("failed find VMI with key: ", fmt.Sprintf("%s/%s", name, namespace))
            return vmiUUID, creationTime, err
        } else {
            log.Log.Println("ERROR: ", err, " for vmi key: ", fmt.Sprintf("%s/%s", name, namespace))
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

func (d *DatabaseInstance) GetSubscriptions(page int, perPage int) (map[string]interface{}, error) {
    queryString := "SELECT name, namespace, uuid, source, sourceNamespace, startingCSV, currentCSV, installedCSV, state, creationTime, content from subscriptions"
    resultsMap, err := d.genericGet(queryString, page, perPage)
    if err != nil {
        return nil, err
    }
    return resultsMap, nil
}
