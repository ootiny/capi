package _rt_package_name_

import (
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

var gSqlPostgresCompileArgs = make([]string, 65536)

func init() {
	for i := 0; i < len(gSqlPostgresCompileArgs); i++ {
		gSqlPostgresCompileArgs[i] = fmt.Sprintf("$%d", i+1)
	}
}

type PGSqlAgent struct{}

func NewPGAgent() ISqlAgent {
	return &PGSqlAgent{}
}

func (p *PGSqlAgent) DataSource(host string, port uint16, user string, password string, dbName string) string {
	if dbName == "" {
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s sslmode=disable",
			host, port, user, password,
		)
	} else {
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbName,
		)
	}
}

func (p *PGSqlAgent) HasDatabase(dbName string) string {
	return fmt.Sprintf("SELECT 1 from %s WHERE datname='%s';", "pg_database", dbName)
}

func (p *PGSqlAgent) CreateDatabase(dbName string) string {
	return fmt.Sprintf("CREATE DATABASE \"%s\";", dbName)
}

func (p *PGSqlAgent) DropDatabase(dbName string) string {
	return fmt.Sprintf("DROP DATABASE \"%s\";", dbName)
}

func (p *PGSqlAgent) CreateMetaTable() string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (id text NOT NULL PRIMARY KEY, meta text);", gSqlMetaTableName)

}
func (p *PGSqlAgent) QueryMetaTable() string {
	return fmt.Sprintf("SELECT meta FROM \"%s\" where id = $1;", gSqlMetaTableName)
}

func (p *PGSqlAgent) InsertMetaTable(serviceName string, meta string) string {
	return fmt.Sprintf(
		"INSERT INTO \"%s\" (id, meta) VALUES('%s', '%s');",
		gSqlMetaTableName,
		serviceName,
		meta,
	)
}

func (p *PGSqlAgent) UpdateMetaTable(serviceName string, meta string) string {
	return fmt.Sprintf(
		"UPDATE \"%s\" SET meta = '%s' WHERE id = '%s';",
		gSqlMetaTableName,
		meta,
		serviceName,
	)
}

func (p *PGSqlAgent) CreateServiceTable(serviceName string) string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (id varchar(64) NOT NULL PRIMARY KEY);", serviceName)
}

func (p *PGSqlAgent) AddColumn(serviceName string, columnName string, columnType string) string {
	switch columnType {
	case "LK":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" varchar(64) NOT NULL DEFAULT '';",
			serviceName,
			columnName,
		)
	case "Bool":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" boolean NOT NULL DEFAULT false;",
			serviceName,
			columnName,
		)
	case "Int64":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" bigint NOT NULL DEFAULT 0;",
			serviceName,
			columnName,
		)
	case "Float64":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" double precision NOT NULL DEFAULT 0;",
			serviceName,
			columnName,
		)
	case "Bytes":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" bytea NOT NULL DEFAULT '';",
			serviceName,
			columnName,
		)
	case "String16":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" varchar(16) NOT NULL DEFAULT '';",
			serviceName,
			columnName,
		)
	case "String32":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" varchar(32) NOT NULL DEFAULT '';",
			serviceName,
			columnName,
		)
	case "String64":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" varchar(64) NOT NULL DEFAULT '';",
			serviceName,
			columnName,
		)
	case "String256":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" varchar(256) NOT NULL DEFAULT '';",
			serviceName,
			columnName,
		)
	case "String":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" text NOT NULL DEFAULT '';",
			serviceName,
			columnName,
		)
	case "List<String>":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" text NOT NULL DEFAULT '[]';",
			serviceName,
			columnName,
		)
	case "Map<String>":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" text NOT NULL DEFAULT '{}';",
			serviceName,
			columnName,
		)
	case "LKList":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" text NOT NULL DEFAULT '[]';",
			serviceName,
			columnName,
		)
	case "LKMap":
		return fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" text NOT NULL DEFAULT '{}';",
			serviceName,
			columnName,
		)
	}

	return ""
}

func (p *PGSqlAgent) DropColumn(serviceName string, columnName string) string {
	return fmt.Sprintf("ALTER TABLE \"%s\" DROP COLUMN \"%s\";", serviceName, columnName)
}

func (p *PGSqlAgent) CreateIndex(serviceName string, columnName string) string {
	return fmt.Sprintf(
		"CREATE INDEX %s__index__%s ON \"%s\" (\"%s\");",
		serviceName, columnName, serviceName, columnName,
	)
}

func (p *PGSqlAgent) DropIndex(serviceName string, columnName string) string {
	return fmt.Sprintf(
		"DROP INDEX %s__index__%s;",
		serviceName, columnName,
	)
}

func (p *PGSqlAgent) CreateUnique(serviceName string, columnName string) string {
	return fmt.Sprintf(
		"ALTER TABLE \"%s\" ADD CONSTRAINT %s__unique__%s UNIQUE (\"%s\");",
		serviceName, serviceName, columnName, columnName,
	)
}

func (p *PGSqlAgent) DropUnique(serviceName string, columnName string) string {
	return fmt.Sprintf(
		"ALTER TABLE \"%s\" DROP CONSTRAINT %s__unique__%s;",
		serviceName, serviceName, columnName,
	)
}

func (p *PGSqlAgent) Insert(serviceName string, keys []string) string {
	return fmt.Sprintf(
		"INSERT INTO \"%s\" (%s) VALUES(%s);",
		serviceName,
		"\""+strings.Join(keys, "\",\"")+"\"",
		strings.Join(gSqlPostgresCompileArgs[:len(keys)], ","),
	)
}

func (p *PGSqlAgent) Update(serviceName string, keys []string) string {
	sets := make([]string, len(keys))
	for i, key := range keys {
		sets[i] = "\"" + key + "\" = " + gSqlPostgresCompileArgs[i]
	}

	return fmt.Sprintf(
		"UPDATE \"%s\" SET %s WHERE id = %s;",
		serviceName,
		strings.Join(sets, ","),
		gSqlPostgresCompileArgs[len(keys)],
	)
}

func (p *PGSqlAgent) Delete(serviceName string) string {
	return fmt.Sprintf("DELETE FROM \"%s\" WHERE id = $1;", serviceName)
}

func (p *PGSqlAgent) QueryOrderBy(serviceName string, query *SqlQuery) string {
	queryOrders := query.GetOrders()

	if len(queryOrders) == 0 {
		return ""
	}

	orders := make([]string, len(queryOrders))

	for i, order := range queryOrders {
		if order.asc {
			orders[i] = "\"" + order.name + "\" ASC"
		} else {
			orders[i] = "\"" + order.name + "\" DESC"
		}
	}

	return strings.Join(orders, ", ")
}

func (p *PGSqlAgent) QuerySelect(serviceName string, columns []string) string {
	return "\"" + strings.Join(columns, "\",\"") + "\""
}

func (p *PGSqlAgent) QueryWhere(serviceName string, argStartPos int, query *SqlQuery) (string, []any, error) {
	queryWheres := query.GetWheres()
	whereSqls := make([]string, len(queryWheres))
	args := make([]any, 0)

	if len(queryWheres) == 0 {
		return "", args, nil
	}

	pos := argStartPos

	for i, where := range queryWheres {
		sql := ""

		if i != 0 {
			sql = where.GetConcat() + " "
		}

		op := where.GetOp()
		columnName := where.GetColumnName()

		switch op {
		case SqlEqual, SqlNotEqual, SqlGreaterThan, SqlLessThan, SqlGreaterEqual, SqlLessEqual:
			sql += "\"" + columnName + "\" " + string(op) + " " + gSqlPostgresCompileArgs[pos]
			args = append(args, where.GetValue())
			pos++
		case SqlLike:
			sql += "\"" + columnName + "\" " + string(op) + " '%' || " + gSqlPostgresCompileArgs[pos] + " || '%'"
			args = append(args, where.GetValue())
			pos++
		case SqlIn, SqlNotIn:
			containArgs := []any(nil)

			if v, ok := where.GetValue().([]any); ok {
				containArgs = v
			} else if v, ok := where.GetValue().([]string); ok {
				containArgs = make([]any, len(v))
				for i, it := range v {
					containArgs[i] = it
				}
			} else if v, ok := where.GetValue().([]int64); ok {
				containArgs = make([]any, len(v))
				for i, it := range v {
					containArgs[i] = it
				}
			} else if v, ok := where.GetValue().([]int); ok {
				containArgs = make([]any, len(v))
				for i, it := range v {
					containArgs[i] = it
				}
			} else {
				return "", nil, fmt.Errorf("invalid in args %T", where.GetValue())
			}

			argSql := make([]string, 0)
			for _, it := range containArgs {
				argSql = append(argSql, gSqlPostgresCompileArgs[pos])
				args = append(args, it)
				pos++
			}
			sql += "\"" + columnName + "\" " + string(op) + " (" + strings.Join(argSql, ",") + ")"
		case SqlChild:
			if subSql, subArgs, e := p.QueryWhere(columnName, argStartPos+pos, where.GetValue().(*SqlQuery)); e != nil {
				return "", nil, e
			} else {
				sql += subSql
				args = append(args, subArgs...)
				pos += len(subArgs)
			}
		}

		whereSqls[i] = sql
	}

	return "(" + strings.Join(whereSqls, " ") + ")", args, nil
}
