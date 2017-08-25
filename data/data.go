// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package data

import (
	"database/sql"
	"fmt"
	"github.com/30x/apid-core"
	"github.com/30x/apid-core/api"
	"github.com/30x/apid-core/data/wrap"
	"github.com/30x/apid-core/logger"
	"github.com/Sirupsen/logrus"
	"github.com/mattn/go-sqlite3"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	configDataDriverKey    = "data_driver"
	configDataSourceKey    = "data_source"
	configDataPathKey      = "data_path"
	statCollectionInterval = 10
	commonDBID             = "common"
	commonDBVersion        = "base"

	defaultTraceLevel = "warn"
)

var log, dbTraceLog apid.LogService
var config apid.ConfigService

type dbMapInfo struct {
	db *sql.DB
	closed chan bool
}

var dbMap = make(map[string]*dbMapInfo)
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

func (d *dataService) DB() (apid.DB, error) {
	return d.dbVersionForID(commonDBID, commonDBVersion)
}

func (d *dataService) DBForID(id string) (apid.DB, error) {
	if id == commonDBID {
		return nil, fmt.Errorf("reserved ID: %s", id)
	}
	return d.dbVersionForID(id, commonDBVersion)
}

func (d *dataService) DBVersion(version string) (apid.DB, error) {
	if version == commonDBVersion {
		return nil, fmt.Errorf("reserved version: %s", version)
	}
	return d.dbVersionForID(commonDBID, version)
}

func (d *dataService) DBVersionForID(id, version string) (apid.DB, error) {
	if id == commonDBID {
		return nil, fmt.Errorf("reserved ID: %s", id)
	}
	if version == commonDBVersion {
		return nil, fmt.Errorf("reserved version: %s", version)
	}
	return d.dbVersionForID(id, version)
}

// will set DB to close and delete when no more references
func (d *dataService) ReleaseDB(id, version string) {
	versionedID := VersionedDBID(id, version)

	dbMapSync.Lock()
	defer dbMapSync.Unlock()

	dbm := dbMap[versionedID]
	if dbm.db != nil {
		if strings.EqualFold(config.GetString(logger.ConfigLevel), logrus.DebugLevel.String()) {
			dbm.closed <- true
		}
		log.Warn("SETTING FINALIZER")
		finalizer := Delete(versionedID)
		runtime.SetFinalizer(dbm.db, finalizer)
		dbMap[versionedID] = nil
	} else {
		log.Error("Cannot find DB handle for ver {%s} to release", version)
	}

	return
}

func (d *dataService) dbVersionForID(id, version string) (db *sql.DB, err error) {

	var stoplogchan chan bool
	versionedID := VersionedDBID(id, version)

	dbMapSync.RLock()
	dbm := dbMap[versionedID]
	dbMapSync.RUnlock()
	if dbm != nil && dbm.db != nil {
		return dbm.db, nil
	}

	dbMapSync.Lock()
	defer dbMapSync.Unlock()

	dataPath := DBPath(versionedID)

	if err = os.MkdirAll(path.Dir(dataPath), 0700); err != nil {
		return
	}

	log.Infof("LoadDB: %s", dataPath)
	source := fmt.Sprintf(config.GetString(configDataSourceKey), dataPath)

	wrappedDriverName := "dd:" + config.GetString(configDataDriverKey)
	driver := wrap.NewDriver(&sqlite3.SQLiteDriver{}, dbTraceLog)
	func() {
		// just ignore the "registered twice" panic
		defer func() {
			recover()
		}()
		sql.Register(wrappedDriverName, driver)
	}()

	db, err = sql.Open(wrappedDriverName, source)
	if err != nil {
		log.Errorf("error loading db: %s", err)
		return
	}

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
	if strings.EqualFold(config.GetString(logger.ConfigLevel),
			logrus.DebugLevel.String()) {
		stoplogchan = logDBInfo(versionedID, db)
	}

	db.SetMaxOpenConns(config.GetInt(api.ConfigDBMaxConns))
	db.SetMaxIdleConns(config.GetInt(api.ConfigDBIdleConns))
	db.SetConnMaxLifetime(time.Duration(config.GetInt(api.ConfigDBConnsTimeout)) * time.Second)
	dbInfo := dbMapInfo {db: db, closed: stoplogchan}
	dbMap[versionedID] = &dbInfo
	return
}

func Delete(versionedID string) interface{} {
	return func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Errorf("error closing DB: %v", err)
		}
		dataDir := path.Dir(DBPath(versionedID))
		err = os.RemoveAll(dataDir)
		if err != nil {
			log.Errorf("error removing DB files: %v", err)
		}
		delete(dbMap, versionedID)
	}
}

func VersionedDBID(id, version string) string {
	return path.Join(id, version)
}

func DBPath(id string) string {
	storagePath := config.GetString("local_storage_path")
	relativeDataPath := config.GetString(configDataPathKey)
	return path.Join(storagePath, relativeDataPath, id, "sqlite3")
}

func logDBInfo(versionedId string, db *sql.DB) chan bool {
	stop := make(chan bool)
	go func() {
		for {
			select {
			case <-time.After(time.Duration(statCollectionInterval * time.Second)):
				log.Debugf("Current number of open DB connections for ver {%s} is {%d}",
					versionedId, db.Stats().OpenConnections)
			case <-stop:
				log.Debugf("Stop DB conn. logging for ver {%s}", versionedId)
				return
			}
		}
	}()
	return stop
}

