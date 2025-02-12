package generation

import (
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/huandu/xstrings"
	"github.com/jacobbrewer1/goschema/pkg/entities"
)

// Helpers defines the map of functions exposed to generation templates
var Helpers = template.FuncMap{
	"has_primary_key":        hasPrimaryKey,
	"has_autoinc":            hasAutoIncrementingKey,
	"primary_autoinc_column": primaryAutoIncColumn,
	"autoinc_column":         autoIncColumn,
	"non_autoinc_columns":    nonAutoIncColumns,
	"lcfirst":                xstrings.FirstRuneToLower,
	"identity_columns":       identityColumns,
	"non_identity_columns":   nonIdentityColumns,
	"structify":              structify,
	"enum_columns":           enumColumns,
	"unique_column_keys":     uniqueColumnKeys,
	"sorted_columns":         sortedColumns,
	"get_type":               getType,
	"get_tags":               getTags,
}

// hasPrimaryKey returns true if the table has a primary key
func hasPrimaryKey(t *entities.Table) bool {
	return t.PrimaryKey != nil
}

// hasAutoIncrementingKey returns true if the table has an auto-incrementing key
func hasAutoIncrementingKey(t *entities.Table) bool {
	for _, col := range t.Columns {
		if col.AutoIncrementing {
			return true
		}
	}

	return false
}

// primaryAutoIncColumn returns a column that is both auto-incrementing and part of the primary key, if any
func primaryAutoIncColumn(t *entities.Table) *entities.Column {
	for _, col := range t.Columns {
		if col.AutoIncrementing && col.InPrimaryKey {
			return col
		}
	}

	return nil
}

// autoIncColumn returns the auto-incrementing column, if any
func autoIncColumn(t *entities.Table) *entities.Column {
	for _, col := range t.Columns {
		if col.AutoIncrementing {
			return col
		}
	}

	return nil
}

// nonAutoIncColumns returns the non-auto-incrementing columns, if any
func nonAutoIncColumns(t *entities.Table) []*entities.Column {
	cols := make([]*entities.Column, 0, len(t.Columns)-1)
	for _, col := range t.Columns {
		if !col.AutoIncrementing {
			cols = append(cols, col)
		}
	}

	return cols
}

// identityColumns returns the columns that allow a row to be uniquely identified, if any.
// Columns uniquely identifying a row relies on being in at least one of the following:
//   - a primary key
//   - a unique key
func identityColumns(t *entities.Table) []*entities.Column {
	if t.PrimaryKey != nil {
		return t.PrimaryKey.Columns
	}

	// Find the first unique key and return those columns
	for _, key := range t.Keys {
		if strings.HasPrefix(key.Type, "unique") {
			return key.Columns
		}
	}

	return nil
}

// nonIdentityColumns returns the columns that do not allow a row to be uniquely identified, if any.
// Columns that are not uniquely identifying must not be in any of the following:
//   - a primary key
//   - a unique key
func nonIdentityColumns(t *entities.Table) []*entities.Column {
	cols := identityColumns(t)
	if cols == nil {
		return t.Columns
	}

	checker := make(map[*entities.Column]struct{}, len(cols))
	for _, col := range cols {
		checker[col] = struct{}{}
	}

	ret := make([]*entities.Column, 0, len(t.Columns)-len(cols))
	for _, col := range t.Columns {
		if _, ok := checker[col]; !ok {
			ret = append(ret, col)
		}
	}

	return ret
}

// structify attempts to convert a string into a good struct field name
// by following golint conventions
func structify(s string) string {
	s = xstrings.ToCamelCase(s)

	// Capitalize the first letter
	s = xstrings.FirstRuneToUpper(s)

	return s
}

// enumColumns returns the columns which are enum types
func enumColumns(t *entities.Table) []*entities.Column {
	ret := make([]*entities.Column, 0)
	for _, col := range t.Columns {
		if col.Type == "enum" {
			ret = append(ret, col)
		}
	}

	return ret
}

// uniqueColumnKeys returns a list of keys, none of which have the same set of columns as each other
// This is required because you can have mulitple indexs including the exact same columns in the same order,
// but of different types (unique, non-unique, etc). Includes the primary key, if any.
func uniqueColumnKeys(t *entities.Table) []entities.Key {
	m := make(map[string]struct{}, len(t.Keys))
	keys := make([]entities.Key, 0, len(t.Keys))
	if t.PrimaryKey != nil {
		keys = append(keys, *t.PrimaryKey)
		m[fmt.Sprint(t.PrimaryKey.Columns)] = struct{}{}
	}
	for _, key := range t.Keys {
		k := fmt.Sprint(key.Columns)
		if _, ok := m[k]; ok {
			continue
		}
		m[k] = struct{}{}
		keys = append(keys, key)
	}

	return keys
}

// sortedColumns returns a slice of the columns for a given table sorted alphabetically
func sortedColumns(t *entities.Table) []*entities.Column {
	ret := make([]*entities.Column, len(t.Columns))
	copy(ret, t.Columns)
	sort.SliceStable(ret, func(i, j int) bool { return ret[i].Name < ret[j].Name })
	return ret
}

func getType(col *entities.Column) string {
	switch strings.ToLower(col.Type) {
	case "bigint":
		if col.Nullable {
			return NullInt64
		}
		if col.Unsigned {
			return "uint64"
		}
		return "int64"
	case "int":
		if col.Nullable {
			return "usql.NullInt"
		}
		if col.Unsigned {
			return "uint"
		}
		return "int"
	case "tinyint":
		if col.TypeSize == 1 {
			if col.Nullable {
				return "usql.NullBool"
			}
			return "bool"
		}
		if col.Nullable {
			return "usql.NullInt64"
		}
		if col.Unsigned {
			return "uint8"
		}
		return "int8"
	case "smallint":
		if col.Nullable {
			return "usql.NullInt64"
		}
		if col.Unsigned {
			return "uint16"
		}
		return "int16"
	case "mediumint":
		if col.Nullable {
			return "usql.NullInt32"
		}
		if col.Unsigned {
			return "uint32"
		}
		return "int32"
	case "float":
		if col.Nullable {
			return "usql.NullFloat64"
		}
		if col.TypeSize < 25 {
			return "float32"
		}
		return "float64"
	case "decimal", "double":
		if col.Nullable {
			return "usql.NullFloat64"
		}
		return "float64"
	case "char", "varchar", "tinytext", "text", "mediumtext", "longtext", "string":
		if col.Nullable {
			return "usql.NullString"
		}
		return "string"
	case "binary", "varbinary", "tinyblob", "blob", "mediumblob", "longblob":
		return "[]byte"
	case "enum":
		if col.Nullable {
			return "usql.NullEnum"
		}
		return "usql.Enum"
	case "year", "date", "datetime", "timestamp":
		if col.Nullable {
			return "usql.NullTime"
		}
		return "time.Time"
	case "time":
		if col.Nullable {
			return "usql.NullDuration"
		}
		return "usql.Duration"
	case "json":
		return "json.RawMessage"
	default:
		fmt.Println("Unknown type:", col.Type)
		return "unknown"
	}
}

func getTags(col *entities.Column) string {
	tags := fmt.Sprintf("`db:\"%s", col.Name)
	if col.InPrimaryKey {
		tags += ",pk"
	}
	if col.AutoIncrementing {
		tags += ",autoinc"
	}
	tags += "\"`"

	return tags
}
