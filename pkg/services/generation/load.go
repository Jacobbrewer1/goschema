package generation

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Jacobbrewer1/goschema/pkg/models"
	"github.com/pingcap/tidb/ast"
	"github.com/pingcap/tidb/parser"
)

// LoadSQL loads all SQL files in the given paths and parses them
func LoadSQL(paths ...string) ([]*models.Table, error) {
	p := parser.New()
	tables := make([]*models.Table, 0)

	for _, path := range paths {
		matches, err := filepath.Glob(path)
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			matchedTables := make([]*models.Table, 0)
			if fi, err := os.Stat(m); err != nil {
				return nil, err
			} else if fi.IsDir() {
				sqlMatches, sErr := filepath.Glob(filepath.Join(m, "*.sql"))
				if sErr != nil {
					return nil, fmt.Errorf("error globbing %s: %w", m, sErr)
				}
				for _, sqlPath := range sqlMatches {
					if matchedTables, err = parseSQL(p, sqlPath); err != nil {
						return nil, fmt.Errorf("error parsing %s: %w", sqlPath, err)
					}
					tables = append(tables, matchedTables...)
				}
			} else {
				if matchedTables, err = parseSQL(p, m); err != nil {
					return nil, fmt.Errorf("error parsing %s: %w", m, err)
				}
				tables = append(tables, matchedTables...)
			}
		}
	}

	return tables, nil
}

func parseSQL(p *parser.Parser, path string) ([]*models.Table, error) {
	sql, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	stmts, err := p.Parse(string(sql), "", "")
	if err != nil {
		return nil, fmt.Errorf("error parsing SQL: %w", err)
	}

	tables := make([]*models.Table, 0, len(stmts))
	for _, stmt := range stmts {
		ct, ok := stmt.(*ast.CreateTableStmt)
		if !ok {
			// We only support create table statements
			continue
		}

		t, err := models.NewTable(ct)
		if err != nil {
			return nil, fmt.Errorf("error creating table from statement: %w", err)
		}

		tables = append(tables, t)
	}

	return tables, nil
}
