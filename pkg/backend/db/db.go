package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
        pod.Content)
	if err != nil {
		return err
	}

	return nil
} 

var (
	insertPodQuery       = `INSERT INTO pods(keyid, kind, name, namespace, uuid, phase, activeContainers, totalContainers, nodeName, creationTime, content) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE keyid=VALUES(keyid);`
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

	uri := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", d.username, d.password, d.host, d.port, d.dbName)

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
	  PRIMARY KEY (uuid)
	);
	`
	err := d.execTable(createPodsTable)
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

/*func (d *databaseInstance) StoreObjectChange(object *Object, content json.RawMessage) error {
	// TimeString - given a time, return the MySQL standard string representation
	madeAt := time.Now().Format("2006-01-02 15:04:05.999999")
	ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctx, insertObjectQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, object.Key, object.Kind, object.Name, object.Namespace, object.UUID)
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(d.ctx, 1*time.Second)
	defer cancel()

	stmt, err = d.db.PrepareContext(ctx, insertObjectChangeQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, object.Key, madeAt, object.UUID, content)
	if err != nil {
		return err
	}

	return nil
} */

