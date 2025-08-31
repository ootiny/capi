package builder

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type DBTableColumnMeta struct {
	Type     string   `json:"type"`
	Query    []string `json:"query"`
	Unique   bool     `json:"unique"`
	Index    bool     `json:"index"`
	Order    bool     `json:"order"`
	Required bool     `json:"required"`
}

func (p *DBTableColumnMeta) ToDBTableColumn() (*DBTableColumn, error) {
	// parse type
	strType := ""
	strTable := ""
	switch p.Type {
	case "PK", "Bool", "Int64", "Float64",
		"String", "String16", "String32", "String64", "String256",
		"List<String>", "Map<String>":
		strType = p.Type
		strTable = ""
	default:
		if strings.HasPrefix(p.Type, DBPrefix) {
			strType = "LK"
			strTable = NamespaceToTableName(p.Type)
		} else if strings.HasPrefix(p.Type, "List<") && strings.HasSuffix(p.Type, ">") {
			innerType := p.Type[5 : len(p.Type)-1]
			if strings.HasPrefix(innerType, DBPrefix) {
				strType = "LKList"
				strTable = NamespaceToTableName(innerType)
			} else {
				return nil, fmt.Errorf("invalid column type: %s", p.Type)
			}
		} else if strings.HasPrefix(p.Type, "Map<") && strings.HasSuffix(p.Type, ">") {
			innerType := p.Type[4 : len(p.Type)-1]
			if strings.HasPrefix(innerType, DBPrefix) {
				strType = "LKMap"
				strTable = NamespaceToTableName(innerType)
			} else {
				return nil, fmt.Errorf("invalid column type: %s", p.Type)
			}
		} else {
			return nil, fmt.Errorf("invalid column type: %s", p.Type)
		}
	}

	// build query map
	queryMap := map[string]bool{}
	for _, v := range p.Query {
		queryMap[v] = true
	}

	return &DBTableColumn{
		Type:      strType,
		QueryMap:  queryMap,
		Unique:    p.Unique,
		Index:     p.Index,
		Order:     p.Order,
		Required:  p.Required,
		LinkTable: strTable,
	}, nil
}

type DBTableViewMeta struct {
	Cache   string   `json:"cache"`
	Columns []string `json:"columns"`
}

type DBTableMeta struct {
	Version      string                        `json:"version"`
	Table        string                        `json:"table"`
	Columns      map[string]*DBTableColumnMeta `json:"columns"`
	Views        map[string]*DBTableViewMeta   `json:"views"`
	__filepath__ string
}

func (p *DBTableMeta) GetFilePath() string {
	return p.__filepath__
}

func (p *DBTableMeta) ToAPIMeta() (*APIMeta, error) {
	fnDBTypeToAPIType := func(v string, klass string) string {
		switch v {
		case "PK":
			return "String"
		case "Bool":
			return "Bool"
		case "Int64":
			return "Int64"
		case "Float64":
			return "Float64"
		case "String", "String32", "String64", "String256":
			return "String"
		case "List<String>":
			return "List<String>"
		case "Map<String>":
			return "Map<String>"
		default:
			if strings.HasPrefix(v, "List<") && strings.HasSuffix(v, ">") {
				innerType := v[5 : len(v)-1]
				return fmt.Sprintf("List<%s@%s>", innerType, klass)
			} else if strings.HasPrefix(v, "Map<") && strings.HasSuffix(v, ">") {
				innerType := v[4 : len(v)-1]
				return fmt.Sprintf("Map<%s@%s>", innerType, klass)
			} else if strings.HasPrefix(v, DBPrefix) {
				return fmt.Sprintf("%s@%s", v, klass)
			} else {
				return v
			}
		}
	}

	definitions := map[string]*APIDefinitionMeta{}

	for name, view := range p.Views {
		attributes := []*APIDefinitionAttributeMeta{}

		for _, column := range view.Columns {
			columnName := ""
			columnType := ""
			columnArray := strings.Split(column, "@")
			if len(columnArray) == 1 {
				columnName = columnArray[0]
				columnType = fnDBTypeToAPIType(p.Columns[columnName].Type, "")
			} else {
				columnName = columnArray[0]
				columnType = fnDBTypeToAPIType(p.Columns[columnName].Type, columnArray[1])
			}

			attributes = append(attributes, &APIDefinitionAttributeMeta{
				Name:     columnName,
				Type:     columnType,
				Required: p.Columns[columnName].Required,
			})
		}
		definitions[name] = &APIDefinitionMeta{
			Attributes: attributes,
		}
	}

	return &APIMeta{
		Version:     "rt.db.v1",
		Namespace:   p.Table,
		Definitions: definitions,
	}, nil
}

func (p *DBTableMeta) ToDBTable() (*DBTable, error) {
	// convert columns
	columns := map[string]*DBTableColumn{}
	for name, column := range p.Columns {
		if dbColumn, err := column.ToDBTableColumn(); err != nil {
			return nil, err
		} else {
			columns[name] = dbColumn
		}
	}

	viewNames := []string{}
	for name := range p.Views {
		viewNames = append(viewNames, name)
	}
	sort.Strings(viewNames)

	// convert views
	views := map[string]*DBTableView{}
	for name, view := range p.Views {
		viewHashArr := []string{}
		viewColumns := []*DBTableViewColumn{}

		// get view index by viewNames
		viewIndex := uint64(0)
		for i, viewName := range viewNames {
			if viewName == name {
				viewIndex = uint64(i)
				break
			}
		}

		for _, viewColumn := range view.Columns {
			columnArray := strings.Split(viewColumn, "@")
			if column, ok := columns[columnArray[0]]; !ok {
				return nil, fmt.Errorf("views.%s column %s not found", name, columnArray[0])
			} else {
				if len(columnArray) == 1 {
					viewColumns = append(viewColumns, &DBTableViewColumn{
						Name:      columnArray[0],
						LinkTable: "",
						LinkView:  "",
					})
					viewHashArr = append(viewHashArr, columnArray[0])
				} else if len(columnArray) == 2 {
					viewColumns = append(viewColumns, &DBTableViewColumn{
						Name:      columnArray[0],
						LinkTable: column.LinkTable,
						LinkView:  columnArray[1],
					})
					viewHashArr = append(viewHashArr, columnArray[0]+"@"+column.LinkTable+"@"+columnArray[1])
				} else {
					return nil, fmt.Errorf("views.%s column %s invalid", name, viewColumn)
				}
			}
		}

		// sort viewHashArr
		sort.Strings(viewHashArr)

		// build columnsSelect
		columnsSelectList := []string{}
		for _, viewColumn := range viewColumns {
			columnsSelectList = append(columnsSelectList, viewColumn.Name)
		}

		if len(columnsSelectList) == 0 {
			return nil, fmt.Errorf("views.%s has no columns", name)
		}

		// parse cache
		cacheSecond, err := TimeStringToDuration(view.Cache)
		if err != nil {
			return nil, fmt.Errorf("views.%s cache invalid: %w", name, err)
		}

		// make md5Hash
		views[name] = &DBTableView{
			Columns:       viewColumns,
			ColumnsSelect: strings.Join(columnsSelectList, ","),
			CacheSecond:   int64(cacheSecond / time.Second),
			Hash:          GetViewHash(viewIndex+1, strings.Join(viewHashArr, ":")),
		}
	}

	return &DBTable{
		Version:   "rt.dbservice.v1",
		Table:     NamespaceToTableName(p.Table),
		Columns:   columns,
		Views:     views,
		Namespace: p.Table,
		File:      p.GetFilePath(),
	}, nil
}

func LoadDBTableMeta(filePath string) (*DBTableMeta, error) {
	var meta DBTableMeta
	if err := UnmarshalConfig(filePath, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse meta file: %w", err)
	} else {
		meta.__filepath__ = filePath
		return &meta, nil
	}
}
