// D10 read-only harness: builds the 1D expectancy table through the PRODUCTION
// code path (expectancy.LoadAndBuildAt) against a read-only DSN. Writes nothing.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"nofx/expectancy"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	gl "gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
	_ "nofx/store/sqlitedriver" // the ONE sqlite registration site (DS-102 fold, CTO 1790305899255)
)

func main() {
	dsn := os.Args[1]
	sqldb, err := sql.Open("sqlite", dsn)
	if err != nil {
		fmt.Println("open:", err)
		os.Exit(1)
	}
	db, err := gorm.Open(sqliteConnDialector{db: sqldb}, &gorm.Config{Logger: gl.Default.LogMode(gl.Silent)})
	if err != nil {
		fmt.Println("open:", err)
		os.Exit(1)
	}
	now := time.Now()
	t, err := expectancy.LoadAndBuildAt(db, now)
	if err != nil {
		fmt.Println("build:", err)
		os.Exit(1)
	}
	fmt.Println(t.BootLine())
	b, _ := json.MarshalIndent(t, "", " ")
	fmt.Println(string(b))
}

// sqliteConnDialector is the minimal gorm Dialector for the ONE
// registration site: it wraps a *sql.DB opened with the "sqlite" driver that
// nofx/store/sqlitedriver registered (DS-102 fold, CTO 1790305899255). It
// mirrors gorm.io/driver/sqlite's Dialector{Conn: ...} behaviour for queries:
// the version check, the default callbacks, and the sqlite clause builders.
// The harness never migrates, so Migrator is the zero value.
type sqliteConnDialector struct{ db *sql.DB }

func (d sqliteConnDialector) Name() string { return "sqlite" }

func (d sqliteConnDialector) Initialize(gdb *gorm.DB) error {
	gdb.ConnPool = d.db
	var version string
	if err := gdb.ConnPool.QueryRowContext(context.Background(), "select sqlite_version()").Scan(&version); err != nil {
		return err
	}
	if compareVersion(version, "3.35.0") >= 0 {
		callbacks.RegisterDefaultCallbacks(gdb, &callbacks.Config{
			CreateClauses:        []string{"INSERT", "VALUES", "ON CONFLICT", "RETURNING"},
			UpdateClauses:        []string{"UPDATE", "SET", "FROM", "WHERE", "RETURNING"},
			DeleteClauses:        []string{"DELETE", "FROM", "WHERE", "RETURNING"},
			LastInsertIDReversed: true,
		})
	} else {
		callbacks.RegisterDefaultCallbacks(gdb, &callbacks.Config{LastInsertIDReversed: true})
	}
	for k, v := range d.ClauseBuilders() {
		if _, ok := gdb.ClauseBuilders[k]; !ok {
			gdb.ClauseBuilders[k] = v
		}
	}
	return nil
}

func (d sqliteConnDialector) ClauseBuilders() map[string]clause.ClauseBuilder {
	return map[string]clause.ClauseBuilder{
		"INSERT": func(c clause.Clause, builder clause.Builder) {
			if insert, ok := c.Expression.(clause.Insert); ok {
				if stmt, ok := builder.(*gorm.Statement); ok {
					stmt.WriteString("INSERT ")
					if insert.Modifier != "" {
						stmt.WriteString(insert.Modifier)
						stmt.WriteByte(' ')
					}
					stmt.WriteString("INTO ")
					if insert.Table.Name == "" {
						stmt.WriteQuoted(stmt.Table)
					} else {
						stmt.WriteQuoted(insert.Table)
					}
					return
				}
			}
			c.Build(builder)
		},
		"LIMIT": func(c clause.Clause, builder clause.Builder) {
			if limit, ok := c.Expression.(clause.Limit); ok {
				lmt := -1
				if limit.Limit != nil && *limit.Limit >= 0 {
					lmt = *limit.Limit
				}
				if lmt >= 0 || limit.Offset > 0 {
					builder.WriteString("LIMIT ")
					builder.WriteString(strconv.Itoa(lmt))
				}
				if limit.Offset > 0 {
					builder.WriteString(" OFFSET ")
					builder.WriteString(strconv.Itoa(limit.Offset))
				}
			}
		},
		"FOR": func(c clause.Clause, builder clause.Builder) {
			if _, ok := c.Expression.(clause.Locking); ok {
				return // sqlite has no row-level locking
			}
			c.Build(builder)
		},
	}
}

type sqliteMigrator struct{ migrator.Migrator }

func (d sqliteConnDialector) Migrator(gdb *gorm.DB) gorm.Migrator {
	return sqliteMigrator{migrator.Migrator{Config: migrator.Config{
		DB:                          gdb,
		Dialector:                   d,
		CreateIndexAfterCreateTable: true,
	}}}
}

func (d sqliteConnDialector) DataTypeOf(*schema.Field) string { return "" }

func (d sqliteConnDialector) DefaultValueOf(*schema.Field) clause.Expression { return nil }

func (d sqliteConnDialector) BindVarTo(writer clause.Writer, _ *gorm.Statement, _ interface{}) {
	writer.WriteByte('?')
}

func (d sqliteConnDialector) QuoteTo(writer clause.Writer, str string) {
	var (
		underQuoted, selfQuoted bool
		continuousBacktick      int8
		shiftDelimiter          int8
	)
	for _, v := range []byte(str) {
		switch v {
		case '`':
			continuousBacktick++
			if continuousBacktick == 2 {
				writer.WriteString("``")
				continuousBacktick = 0
			}
		case '.':
			if continuousBacktick > 0 || !selfQuoted {
				shiftDelimiter = 0
				underQuoted = false
				continuousBacktick = 0
				writer.WriteString("`")
			}
			writer.WriteByte(v)
			continue
		default:
			if shiftDelimiter-continuousBacktick <= 0 && !underQuoted {
				writer.WriteString("`")
				underQuoted = true
				if selfQuoted = continuousBacktick > 0; selfQuoted {
					continuousBacktick--
				}
			}
			for ; continuousBacktick > 0; continuousBacktick-- {
				writer.WriteString("``")
			}
			writer.WriteByte(v)
		}
		shiftDelimiter++
	}
	if continuousBacktick > 0 && !selfQuoted {
		writer.WriteString("``")
	}
	writer.WriteString("`")
}

func (d sqliteConnDialector) Explain(sql string, vars ...interface{}) string {
	return gl.ExplainSQL(sql, nil, `"`, vars...)
}

// compareVersion mirrors gorm.io/driver/sqlite's own check.
func compareVersion(v1, v2 string) int {
	for i, j := 0, 0; i < len(v1) || j < len(v2); i, j = i+1, j+1 {
		for i < len(v1) && v1[i] < '0' || i < len(v1) && v1[i] > '9' {
			i++
		}
		for j < len(v2) && v2[j] < '0' || j < len(v2) && v2[j] > '9' {
			j++
		}
		var n1, n2 int
		for i < len(v1) && v1[i] >= '0' && v1[i] <= '9' {
			n1 = n1*10 + int(v1[i]-'0')
			i++
		}
		for j < len(v2) && v2[j] >= '0' && v2[j] <= '9' {
			n2 = n2*10 + int(v2[j]-'0')
			j++
		}
		if n1 != n2 {
			if n1 > n2 {
				return 1
			}
			return -1
		}
	}
	return 0
}
