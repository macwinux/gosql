package gosql

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

/*
Our in memory backend should store a list of tables. Each table will have a list of columns and rows.
Each column will have a name and type. Each row will have a list of byte arrays.
*/

// MemoryCell Each piece of information store in the database
type MemoryCell []byte

func (mc MemoryCell) AsInt() int32 {
	var i int32
	err := binary.Read(bytes.NewBuffer(mc), binary.BigEndian, &i)
	if err != nil {
		panic(err)
	}
	return i
}

func (mc MemoryCell) AsText() string {
	return string(mc)
}

type table struct {
	columns     []string
	columnTypes []ColumnType
	rows        [][]MemoryCell
}

type MemoryBackend struct {
	tables map[string]*table
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		tables: map[string]*table{},
	}
}

/*
Create Table Support
--------------------
When creating a table, we'll make a new entry in the backend tables map. Then we'll create columns as
specified by the AST
*/

func (mb *MemoryBackend) CreateTable(crt *CreateTableStatement) error {
	t := table{}
	mb.tables[crt.name.value] = &t
	if crt.cols == nil {
		return nil
	}
	for _, col := range crt.cols {
		t.columns = append(t.columns, col.name.value)

		var dt ColumnType
		switch col.datatype.value {
		case "int":
			dt = IntType
		case "text":
			dt = TextType
		default:
			return ErrorInvalidDataType
		}
		t.columnTypes = append(t.columnTypes, dt)
	}
	return nil
}

/*
Insert Support
--------------
Keeping things simple, we'll assume the value passed can be correctly mapped to the type of the column specified
*/

func (mb *MemoryBackend) Insert(inst *InsertStatement) error {
	table, ok := mb.tables[inst.table.value]
	if !ok {
		return ErrTableDoesNotExist
	}
	if inst.values == nil {
		return nil
	}

	row := []MemoryCell{}

	if len(inst.values) != len(table.columns) {
		return ErrMissingValues
	}

	for _, value := range inst.values {
		if value.kind != literalKind {
			fmt.Println("Skipping non-literal.")
			continue
		}
		row = append(row, mb.tokenToCell(value.literal))
	}
	table.rows = append(table.rows, row)
	return nil
}

// tokenToCell helper will write numbers as binary bytes and will write strings as bytes
func (mb *MemoryBackend) tokenToCell(t *token) MemoryCell {
	if t.kind == numericKind {
		buf := new(bytes.Buffer)
		i, err := strconv.Atoi(t.value)
		if err != nil {
			panic(err)
		}
		err = binary.Write(buf, binary.BigEndian, int32(i))
		if err != nil {
			panic(err)
		}
		return MemoryCell(buf.Bytes())
	}
	if t.kind == stringKind {
		return MemoryCell(t.value)
	}
	return nil
}

/*
Select Support
--------------
For select we'll iterate over each row in the table and return the cells according to the columns specified by teh AST

*/

func (mb *MemoryBackend) Select(slct *SelectStatement) (*Results, error) {
	table, ok := mb.tables[slct.from.value]
	if !ok {
		return nil, ErrTableDoesNotExist
	}

	results := [][]Cell{}
	columns := []struct {
		Type ColumnType
		Name string
	}{}

	for i, row := range table.rows {
		result := []Cell{}
		isFirstRow := i == 0

		for _, exp := range slct.item {
			if exp.kind != literalKind {
				// Unsupported, doesn't currently exit, ignore.
				fmt.Println("Skipping non-literal expression.")
				continue
			}

			lit := exp.literal
			if lit.kind == identifierKind {
				found := false
				for i, tableCol := range table.columns {
					if tableCol == lit.value {
						if isFirstRow {
							columns = append(columns, struct {
								Type ColumnType
								Name string
							}{Type: table.columnTypes[i], Name: lit.value})
						}
						result = append(result, row[i])
						found = true
						break
					}
				}
				if !found {
					return nil, ErrColumnDoesNotExist
				}
				continue
			}
			return nil, ErrColumnDoesNotExist
		}
		results = append(results, result)
	}
	return &Results{
		Columns: columns,
		Rows:    results,
	}, nil
}
