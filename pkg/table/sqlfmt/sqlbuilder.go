package sqlfmt

import (
	"strings"

	"github.com/cagnosolutions/go-data/pkg/format"
	"github.com/cagnosolutions/go-data/pkg/table"
)

const (
	stmtCreate = `CREATE TABLE IF NOT EXISTS (
	id INTEGER NOT NULL PRIMARY KEY,
	$field $type NOT NULL DEFAULT $default,
);`
	stmtCreateSize = 125 // rough estimate of stmt size without table info
)

type SQLBuilder struct {
}

func (sql *SQLBuilder) CreateStmt(t *table.Table) string {
	var sb strings.Builder
	var hasPK bool
	sb.Grow(t.Size() + stmtCreateSize)
	sb.WriteString("CREATE TABLE IF NOT EXISTS ")
	sb.WriteString(format.ToSnakeCase(t.Name))
	sb.WriteString(" (\n")
	for i := range t.Fields {
		sb.WriteString("\t")
		sb.WriteString(format.ToSnakeCase(t.Fields[i].Name))
		sb.WriteString(" ")
		sb.WriteString(GetSQLiteType(t.Fields[i].Kind))
		if t.Fields[i].Name == "ID" && !hasPK {
			sb.WriteString(" NOT NULL PRIMARY KEY,\n")
			hasPK = true
			continue
		}
		if i < len(t.Fields)-1 {
			sb.WriteString(" NOT NULL,\n")
			continue
		}
		sb.WriteString(" NOT NULL\n")
	}
	sb.WriteString(");")
	return sb.String()
}
