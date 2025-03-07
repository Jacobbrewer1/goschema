package generation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jacobbrewer1/goschema/pkg/entities"
	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
)

// LoadSQL loads all SQL files in the given paths and parses them
func LoadSQL(paths ...string) ([]*entities.Table, error) {
	p := parser.New()
	tables := make([]*entities.Table, 0)

	for _, path := range paths {
		matches, err := filepath.Glob(path)
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			var matchedTables []*entities.Table
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

func parseSQL(p *parser.Parser, path string) ([]*entities.Table, error) {
	sql, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Loop through each line and remove any `with system versioning` clauses
	lines := strings.Split(string(sql), "\n")
	newSql := ""
	for i, line := range lines {
		if strings.Contains(line, "with system versioning") {
			newSql = strings.ReplaceAll(line, "with system versioning", "")
		} else {
			newSql = line
		}
		lines[i] = newSql
	}
	sql = []byte(strings.Join(lines, "\n"))

	stmts, _, err := p.ParseSQL(string(sql))
	if err != nil {
		return nil, fmt.Errorf("error parsing SQL: %w", err)
	}

	tables := make([]*entities.Table, 0, len(stmts))
	for _, stmt := range stmts {
		ct, ok := stmt.(*ast.CreateTableStmt)
		if !ok {
			// We only support create table statements
			continue
		}

		t, err := entities.NewTable(ct)
		if err != nil {
			return nil, fmt.Errorf("error creating table from statement: %w", err)
		}

		tables = append(tables, t)
	}

	return tables, nil
}
