package models

// Key represents a MySQL key (primary, unique, index, etc)
type Key struct {
	Name    string
	Type    string
	Columns []*Column
	Comment string
}
