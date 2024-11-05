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
	SelectStatement      *SelectStatement
	CreateTableStatement *CreateTableStatement
	InsertStatement      *InsertStatement
	Kind                 AStKind
}

// An insert statement for now, has a table name and a list of values to insert:
type InsertStatement struct {
	table  token
	values []*expression
}

// An expression is a literal token or (in the future) a function call or inline operation:
type expressionKind uint

const (
	literalKind expressionKind = iota
	//binaryKind
)

//type binaryExpression struct {
//	A  expression
//	B  expression
//	Op token
//}
//
//func (be binaryExpression) GenerateCode() string {
//	return fmt.Sprintf("(%s %s %s)", be.A.GenerateCode(), be.Op.value, be.B.GenerateCode())
//}

type expression struct {
	literal *token
	kind    expressionKind
	//binary  *binaryExpression
}

//func (e expression) GenerateCode() string {
//	switch e.kind {
//	case literalKind:
//		switch e.literal.kind {
//		case identifierKind:
//			return fmt.Sprintf("\"%s\"", e.literal.value)
//		case stringKind:
//			return fmt.Sprintf("'%s'", e.literal.value)
//		default:
//			return fmt.Sprintf(e.literal.value)
//		}
//case binaryKind:
//	return e.binary.GenerateCode()
//}
//return ""
//}

// A create statement, for now, has a table name and a list of column names and types:
type columnDefinition struct {
	name     token
	datatype token
}

type CreateTableStatement struct {
	name token
	cols []*columnDefinition
}

// A select statemen, for now, has a table name and a list of column names:
type SelectStatement struct {
	item []*expression
	from token
}
