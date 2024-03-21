package models

import (
	"log"
	"log/slog"

	"github.com/pingcap/tidb/ast"
)

// Table represents a MySQL table definition
type Table struct {
	Name        string
	Columns     []*Column
	colMap      map[string]*Column
	PrimaryKey  *Key
	Keys        []Key
	Constraints []Constraint
	Comment     string
}

func (t *Table) setColumns(ct *ast.CreateTableStmt) error {
	for i, col := range ct.Cols {
		t.Columns[i] = &Column{
			Name: col.Name.String(),
		}
		t.Columns[i].setTypeInfo(col.Tp)
		t.Columns[i].setFlags(col.Tp.Flag)
		if err := t.Columns[i].setOptions(col); err != nil {
			return err
		}
		t.colMap[t.Columns[i].Name] = t.Columns[i]
	}

	return nil
}

func (t *Table) addForeignKeyConstraint(con *ast.Constraint) {
	c := Constraint{Name: con.Name}
	if con.Refer != nil {
		c.ReferenceTable = con.Refer.Table.Name.String()
	}
	if con.Option != nil {
		c.Comment = con.Option.Comment
	}
	c.setReferences(con)
	t.Constraints = append(t.Constraints, c)
}

func (t *Table) setPrimaryKey(con *ast.Constraint) {
	t.PrimaryKey = &Key{Name: con.Name, Type: "primary"}
	if t.PrimaryKey.Name == "" {
		t.PrimaryKey.Name = "primary"
	}
	if con.Option != nil {
		t.PrimaryKey.Comment = con.Option.Comment
	}
	t.PrimaryKey.Columns = make([]*Column, len(con.Keys))
	for i, col := range con.Keys {
		t.PrimaryKey.Columns[i] = t.colMap[col.Column.String()]
		t.PrimaryKey.Columns[i].InPrimaryKey = true
		t.PrimaryKey.Columns[i].InUniqueKey = true
	}
}

func (t *Table) addKey(con *ast.Constraint) {
	k := Key{Name: con.Name}
	var uniq bool
	switch con.Tp {
	case ast.ConstraintKey:
		k.Type = "key"
	case ast.ConstraintIndex:
		k.Type = "index"
	case ast.ConstraintUniq:
		k.Type = "unique"
		uniq = true
	case ast.ConstraintUniqKey:
		k.Type = "unique_key"
		uniq = true
	case ast.ConstraintUniqIndex:
		k.Type = "unique_index"
		uniq = true
	case ast.ConstraintFulltext:
		k.Type = "fulltext"
	default:
		slog.Warn("unknown key type", slog.Int("type", int(con.Tp)))
	}
	if con.Option != nil {
		k.Comment = con.Option.Comment
	}

	k.Columns = make([]*Column, len(con.Keys))
	for i, col := range con.Keys {
		var ok bool
		if k.Columns[i], ok = t.colMap[col.Column.String()]; !ok {
			log.Printf("warning: found index for invalid field %q\n", col.Column.String())
			return
		}

		k.Columns[i].InUniqueKey = uniq
	}

	t.Keys = append(t.Keys, k)
}

// setTableOptions sets the options for the table
func (t *Table) setTableOptions(ct *ast.CreateTableStmt) error {
	for _, opt := range ct.Options {
		switch opt.Tp {
		case ast.TableOptionComment:
			t.Comment = opt.StrValue
		case ast.TableOptionEngine,
			ast.TableOptionCharset:
			// ignore
		default:
			slog.Warn("unknown table option", slog.Int("type", int(opt.Tp)))
		}
	}

	return nil
}

// NewTable returns a Table struct representing the result of a MySQL CREATE TABLE statement
func NewTable(ct *ast.CreateTableStmt) (*Table, error) {
	table := &Table{
		Name:    ct.Table.Name.String(),
		Columns: make([]*Column, len(ct.Cols)),
		colMap:  make(map[string]*Column, len(ct.Cols)),
	}

	if err := table.setTableOptions(ct); err != nil {
		return nil, err
	}

	if err := table.setColumns(ct); err != nil {
		return nil, err
	}

	for _, con := range ct.Constraints {
		switch con.Tp {
		case ast.ConstraintForeignKey:
			table.addForeignKeyConstraint(con)
		case ast.ConstraintPrimaryKey:
			table.setPrimaryKey(con)
		default:
			table.addKey(con)
		}
	}

	return table, nil
}
