package builder

type DBConnectConfig struct {
	Driver   string `json:"driver" required:"true"`
	Host     string `json:"host" required:"true"`
	Port     uint16 `json:"port" required:"true"`
	User     string `json:"user" required:"true"`
	Password string `json:"password" required:"true"`
	DBName   string `json:"dbName" required:"true"`
}

type DBCacheConfig struct {
	Type string `json:"type" required:"true"`
	Size string `json:"size" required:"true"`
	Addr string `json:"addr"`
}

type DBConfig struct {
	Connect *DBConnectConfig `json:"connect" required:"true"`
	Cache   *DBCacheConfig   `json:"cache" required:"true"`
}

type DBTableColumn struct {
	Type      string          `json:"type"`
	QueryMap  map[string]bool `json:"queryMap"`
	Unique    bool            `json:"unique"`
	Index     bool            `json:"index"`
	Order     bool            `json:"order"`
	Required  bool            `json:"required"`
	LinkTable string          `json:"linkTable"`
}

type DBTableViewColumn struct {
	Name      string `json:"name"`
	LinkTable string `json:"linkTable"`
	LinkView  string `json:"linkView"`
}

type DBTableView struct {
	Columns       []*DBTableViewColumn `json:"columns"`
	ColumnsSelect string               `json:"columnsSelect"`
	CacheSecond   int64                `json:"cacheSecond"`
	Hash          string               `json:"hash"`
}

type DBTable struct {
	Version   string                    `json:"version"`
	Namespace string                    `json:"namespace"`
	Table     string                    `json:"table"`
	Columns   map[string]*DBTableColumn `json:"columns"`
	Views     map[string]*DBTableView   `json:"views"`
	File      string                    `json:"file"`
}
