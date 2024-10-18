package gosql

type Ast struct {
	Statements []*Statement
}

type AStKind uint

const (
	SelectKind AStKind = iota
	CreateTableKind
	InsertKind
)

type Statement struct {
	SelectStatement *SelectStatement
	CreateTableStatement *CreateTableStatement
	InsertStatement *InsertStatement
	Kind AStKind
}


// An insert statement for now, has a table name and a list of values to insert:
type InsertStatement struct {
	table token
	values *[]*expression
}

// An expression is a literal token or (in the future) a function call or inline operation:
type expressionKind uint

const(
	literalKind expressionKind = iota
)

type expression struct {
	literal *token
	kind expressionKind
}

// A create statement, for now, has a table name and a list of column names and types:
type columnDefinition struct {
	name token
	datatype token
}

type CreateTableStatement struct {
	name token
	cols *[]*columnDefinition
}

// A select statemen, for now, has a table name and a list of column names:
type SelectStatement struct {
	item []*expression
	from token
}

