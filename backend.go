package gosql

import "errors"

type ColumnType uint

const (
	TextType ColumnType = iota
	IntType
)

type Cell interface {
	AsText() string
	AsInt() int32
}

type Results struct {
	Columns []struct {
		Type ColumnType
		Name string
	}
	Rows [][]Cell
}

var (
	ErrColumnDoesNotExits = errors.New("column does not exist")
	ErrorInvalidDataType  = errors.New("invalid data type")
)

type Backend interface {
	CreateTable(statement *CreateTableStatement) error
	Insert(*InsertStatement) error
	Select(*SelectStatement) (*Results, error)
}
