package entities

import (
	"github.com/pingcap/tidb/pkg/parser/ast"
)

// Constraint represents a MySQL foreign key constraint
type Constraint struct {
	Name           string
	ReferenceTable string
	References     map[string]string
	Comment        string
}

func (c *Constraint) setReferences(con *ast.Constraint) {
	c.References = make(map[string]string, len(con.Keys))
	for i, col := range con.Keys {
		c.References[col.Column.String()] = con.Refer.IndexPartSpecifications[i].Column.String()
	}
}
