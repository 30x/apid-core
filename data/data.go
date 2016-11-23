package data

import (
	"database/sql"
	"fmt"
	"github.com/30x/apid"
	"github.com/mattn/go-sqlite3"
	"path"
	"sync"
	"os"
	"github.com/30x/apid/data/wrap"
)

const (
	configDataDriverKey = "data_driver"
	configDataSourceKey = "data_source"
	configDataPathKey   = "data_path"

	commonDBID = "_apid_common_"

	defaultTraceLevel = "warn"
)

var log, dbTraceLog apid.LogService
var config apid.ConfigService

var dbMap = make(map[string]*sql.DB)
var dbMapSync sync.RWMutex

func CreateDataService() apid.DataService {
	config = apid.Config()
	log = apid.Log().ForModule("data")

	// we don't want to trace normally
	config.SetDefault("DATA_TRACE_LOG_LEVEL", defaultTraceLevel)
	dbTraceLog = apid.Log().ForModule("data_trace")

	config.SetDefault(configDataDriverKey, "sqlite3")
	config.SetDefault(configDataSourceKey, "file:%s")
	config.SetDefault(configDataPathKey, "sqlite")

	return &dataService{}
}

type dataService struct {
}

func (d *dataService) DB() (*sql.DB, error) {
	return d.DBForID(commonDBID)
}

func (d *dataService) DBForID(id string) (db *sql.DB, err error) {

	dbMapSync.RLock()
	db = dbMap[id]
	dbMapSync.RUnlock()
	if db != nil {
		return
	}

	dbMapSync.Lock()
	defer dbMapSync.Unlock()

	db = dbMap[id]
	if db != nil {
		return
	}

	storagePath := config.GetString("local_storage_path")
	relativeDataPath := config.GetString(configDataPathKey)
	dataPath := path.Join(storagePath, relativeDataPath)

	if err = os.MkdirAll(dataPath, 0700); err != nil {
		return
	}

	dataFile := path.Join(dataPath, id)
	log.Infof("LoadDB: %s", dataFile)
	dataSource := fmt.Sprintf(config.GetString(configDataSourceKey), dataFile)

	wrappedDriverName := "dd:" + config.GetString(configDataDriverKey)
	dataDriver := wrap.WrapDriver{&sqlite3.SQLiteDriver{}, dbTraceLog}
	sql.Register(wrappedDriverName, &dataDriver)

	db, err = sql.Open(wrappedDriverName, dataSource)
	if err != nil {
		log.Errorf("error loading db: %s", err)
		return
	}

	log.Infof("Sqlite DB path used by apid: %s", dataPath)
	err = db.Ping()
	if err != nil {
		log.Errorf("error pinging db: %s", err)
		return
	}

	sqlString := "PRAGMA journal_mode=WAL;"
	_, err = db.Exec(sqlString)
	if err != nil {
		log.Errorf("error setting journal_mode: %s", err)
		return
	}

	sqlString = "PRAGMA foreign_keys = ON;"
	_, err = db.Exec(sqlString)
	if err != nil {
		log.Errorf("error enabling foreign_keys: %s", err)
		return
	}

	dbMap[id] = db
	return
}
