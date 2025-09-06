package _rt_package_name_

// import (
// 	"database/sql"
// 	"embed"
// 	"fmt"
// 	"io/fs"
// 	"strings"
// 	"sync"
// )

// type SQLManager struct {
// 	agent      ISqlAgent
// 	config     *RTConfig
// 	db         *sql.DB
// 	serviceMap map[string]*DBTable
// 	mutex      *sync.Mutex
// }

// func NewSQLManager(config *RTConfig) (*SQLManager, error) {
// 	agent := ISqlAgent(nil)

// 	switch config.Database.Driver {
// 	case "postgres":
// 		agent = NewPGAgent()
// 	default:
// 		return nil, fmt.Errorf("invalid driver name %s", config.Database.Driver)
// 	}

// 	if db, err := sql.Open(
// 		config.Database.Driver,
// 		agent.DataSourceWithDB(config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password, config.Database.DBName),
// 	); err != nil {
// 		return nil, err
// 	} else {
// 		return &SQLManager{
// 			agent:      agent,
// 			config:     config,
// 			db:         db,
// 			serviceMap: make(map[string]*DBTable),
// 			mutex:      &sync.Mutex{},
// 		}, nil
// 	}
// }

// func (p *SQLManager) CreateDatabaseIfNotExist() (err error) {
// 	dbConfig := p.config.Database
// 	dataSource := p.agent.DataSource(dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password)

// 	if db, err := sql.Open(dbConfig.Driver, dataSource); err != nil {
// 		return err
// 	} else {
// 		defer func() {
// 			err = GetFirstError(err, WrapError(db.Close()))
// 		}()

// 		var exists bool

// 		if e = db.QueryRow(p.agent.HasDatabase(p.dbName)).Scan(&exists); e != nil {
// 			return WrapError(e)
// 		}

// 		if !exists {
// 			if _, e = db.Exec(p.agent.CreateDatabase(p.dbName)); e != nil {
// 				return WrapError(e)
// 			}
// 		}

// 		return nil
// 	}
// }

// func (p *SQLManager) NewTransaction(isolationLevel string, readOnly bool) *SQLTransaction {
// 	return &SQLTransaction{
// 		tx:             nil,
// 		dbMgr:          p,
// 		isolationLevel: isolationLevel,
// 		readOnly:       readOnly,
// 		mutex:          &sync.Mutex{},
// 	}
// }

// func (p *SQLManager) GetService(name string) *SqlServiceMeta {
// 	return p.serviceMap[name]
// }

// func (p *SQLManager) GetViewConfig(service string, view string) *SqlViewConfig {
// 	if serviceMeta, ok := p.serviceMap[service]; !ok {
// 		return nil
// 	} else {
// 		return serviceMeta.Views[view]
// 	}
// }

// func (p *SQLManager) GetColumnType(service string, column string) string {
// 	if serviceMeta, ok := p.serviceMap[service]; !ok {
// 		return ""
// 	} else {
// 		return serviceMeta.GetColumnType(column)
// 	}
// }

// func (p *SQLManager) Open(fSystem *embed.FS) error {
// 	// find all db table configs file
// 	files := make([]string, 0)
// 	if e := fs.WalkDir(
// 		fSystem,
// 		".",
// 		func(path string, d fs.DirEntry, err error) error {
// 			if err != nil {
// 				return err
// 			}

// 			if !d.IsDir() && strings.HasSuffix(path, ".json") {
// 				files = append(files, path)
// 			}

// 			return nil
// 		},
// 	); e != nil {
// 		return WrapError(e)
// 	}

// 	// load db configs
// 	tx := p.NewTransaction(SqlLevelSerializable, false)

// 	for _, file := range files {
// 		if fContent, e := fSystem.ReadFile(file); e != nil {
// 			_ = tx.Close(false)
// 			return WrapError(e)
// 		} else if service, e := NewSqlServiceMeta(string(fContent)); e != nil {
// 			_ = tx.Close(false)
// 			return WrapError(e, file)
// 		} else if _, ok := p.serviceMap[service.Name]; ok {
// 			_ = tx.Close(false)
// 			return Errorf(
// 				"db-manager: duplicated service %s",
// 				service.Name,
// 			)
// 		} else if e := tx.UpdateTable(string(fContent)); e != nil {
// 			_ = tx.Close(false)
// 			return WrapError(e)
// 		} else {
// 			p.serviceMap[service.Name] = service
// 		}
// 	}

// 	// check link
// 	for serviceName, service := range p.serviceMap {
// 		for viewName, viewConfig := range service.Views {
// 			for columnName, linkConfig := range viewConfig.mapLink {
// 				if _, ok := p.serviceMap[linkConfig.service]; !ok {
// 					_ = tx.Close(false)
// 					return Errorf(
// 						"%s: Views.%s.columns: %s invalid link service: %s",
// 						serviceName,
// 						viewName,
// 						columnName,
// 						linkConfig.service,
// 					)
// 				}

// 				if _, ok := p.serviceMap[linkConfig.service].Views[linkConfig.view]; !ok {
// 					_ = tx.Close(false)
// 					return Errorf(
// 						"%s: Views.%s.columns: %s invalid link view: %s",
// 						serviceName,
// 						viewName,
// 						columnName,
// 						linkConfig.view,
// 					)
// 				}
// 			}
// 		}
// 	}

// 	if e := tx.Close(true); e != nil {
// 		return WrapError(e)
// 	} else {
// 		return nil
// 	}
// }

// func (p *SQLManager) Close() error {
// 	if p.db != nil {
// 		ret := WrapError(p.db.Close())
// 		p.db = nil
// 		return ret
// 	} else {
// 		return nil
// 	}
// }
