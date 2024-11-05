package gosql

import (
	"fmt"
	"log/slog"
	"strings"
)

// localizacion del token en el codigo
type location struct {
	line uint
	col  uint
}

// para guardar las palabras clave reservadas de SQL
type keyword string

const (
	selectKeyword keyword = "select"
	fromKeyword   keyword = "from"
	asKeyword     keyword = "as"
	tableKeyword  keyword = "table"
	createKeyword keyword = "create"
	insertKeyword keyword = "insert"
	intoKeyword   keyword = "into"
	valuesKeyword keyword = "values"
	intKeyword    keyword = "int"
	textKeyword   keyword = "text"
	whereKeyword  keyword = "where"
	trueKeyword   keyword = "true"
	falseKeyword  keyword = "false"
	nullKeyword   keyword = "null"
)

// para guardar la sintaxis SQL
type symbol string

const (
	semicolonSymbol  symbol = ";"
	asteriskSymbol   symbol = "*"
	commaSymbol      symbol = ","
	leftParenSymbol  symbol = "("
	rightParenSymbol symbol = ")"
	eqSymbol         symbol = "="
	neqSymbol        symbol = "<>"
	neqSymbol2       symbol = "!="
	concatSymbol     symbol = "||"
	plusSymbol       symbol = "+"
	ltSymbol         symbol = "<"
	lteSymbol        symbol = "<="
	gtSymbol         symbol = ">"
	gteSymbol        symbol = ">="
)

type tokenKind uint

const (
	keywordKind tokenKind = iota
	symbolKind
	identifierKind
	stringKind
	numericKind
	boolKind
	nullKind
)

type token struct {
	value string
	kind  tokenKind
	loc   location
}

// indica la posicion actual del lexer
type cursor struct {
	pointer uint
	loc     location
}

func (t *token) equals(other *token) bool {
	return t.value == other.value && t.kind == other.kind
}

// longestMatch iterates through a source string starting at the given cursor to find the longest
// matching substring among the provided options
func longestMatch(source string, ic cursor, options []string) string {
	var value []byte
	var skipList []int
	var match string

	cur := ic

	for cur.pointer < uint(len(source)) {
		value = append(value, strings.ToLower(string(source[cur.pointer]))...)
		cur.pointer++

	match:
		for i, option := range options {
			for _, skip := range skipList {
				if i == skip {
					continue match
				}
			}

			// Deal with case like INT vs INTO
			if option == string(value) {
				skipList = append(skipList, i)
				if len(option) > len(match) {
					match = option
				}
				continue
			}

			sharesprefix := string(value) == option[:cur.pointer-ic.pointer]
			tooLong := len(value) > len(option)
			if tooLong || !sharesprefix {
				skipList = append(skipList, i)
			}
		}
		if len(skipList) == len(options) {
			break
		}
	}
	return match
}

type lexer func(string, cursor) (*token, cursor, bool)

func lexNumeric(source string, ic cursor) (*token, cursor, bool) {
	cur := ic

	periodFound := false
	expMarkerFound := false

	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]
		cur.loc.col++

		isDigit := c >= '0' && c <= '9'
		isPeriod := c == '.'
		isExpMarker := c == 'e'

		slog.Debug("Character", slog.String("char", string(c)))
		slog.Debug("Pointers", slog.Int("cur pointer", int(cur.pointer)), slog.Int("ic pointer", int(ic.pointer)), slog.Int("cur loc", int(cur.loc.col)))
		slog.Debug("Type", slog.Bool("isDigit", isDigit), slog.Bool("isPeriod", isPeriod), slog.Bool("isExpMarker", isExpMarker))
		// Must start with a digit or period
		if cur.pointer == ic.pointer {
			if !isDigit && !isPeriod {
				return nil, ic, false
			}
			periodFound = isPeriod
			continue
		}

		if isPeriod {
			if periodFound {
				return nil, ic, false
			}

			periodFound = true
			continue
		}

		if isExpMarker {
			if expMarkerFound {
				return nil, ic, false
			}

			// No periods allowed after expMarker
			periodFound = true
			expMarkerFound = true

			// expMarker must be followeb by digits
			if cur.pointer == uint(len(source)-1) {
				return nil, ic, false
			}

			//Si es un valor negativo o positivo pasa el cursor.
			cNext := source[cur.pointer+1]
			if cNext == '-' || cNext == '+' {
				cur.pointer++
				cur.loc.col++
			}
			continue
		}
		if !isDigit {
			break
		}
	}

	// No characters accumulated
	if cur.pointer == ic.pointer {
		return nil, ic, false
	}

	slog.Debug("Final Value", slog.Int("cur pointer", int(cur.pointer)), slog.Int("ic pointer", int(ic.pointer)), slog.String("value", source[ic.pointer:cur.pointer]))
	return &token{
		value: source[ic.pointer:cur.pointer],
		loc:   ic.loc,
		kind:  numericKind,
	}, cur, true
}

// Strings must start and end with a single apostrophe. They can contain a  single apostrophe if it is followed by another single apostrophe.
// We'll put this kind of character delimited lexing logic into a helper function so we can use it again when analyzing identifiers

func lexCharacterDelimited(source string, ic cursor, delimiter byte) (*token, cursor, bool) {
	cur := ic

	if len(source[cur.pointer:]) == 0 {
		return nil, ic, false
	}

	if source[cur.pointer] != delimiter {
		return nil, ic, false
	}

	cur.loc.col++
	cur.pointer++

	var value []byte
	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]

		if c == delimiter {
			// SQL escapes are via double characters, not backslash
			if cur.pointer+1 >= uint(len(source)) || source[cur.pointer+1] != delimiter {
				cur.pointer++
				cur.loc.col++
				return &token{
					value: string(value),
					loc:   ic.loc,
					kind:  stringKind,
				}, cur, true
			}

			value = append(value, delimiter)
			cur.pointer++
			cur.loc.col++
		}
		value = append(value, c)
		cur.loc.col++
	}
	return nil, ic, false
}

func lexString(source string, ic cursor) (*token, cursor, bool) {
	return lexCharacterDelimited(source, ic, '"')
}

//	Symbols come from a fixed set of strings, so they're easy to compare against.
//	Whitespace should be thrown away.

func lexSymbol(source string, ic cursor) (*token, cursor, bool) {
	c := source[ic.pointer]
	cur := ic
	//Will get overwritteng later if not an ignored syntax
	cur.pointer++
	cur.loc.col++

	switch c {
	//Syntax taht should be thrown away
	case '\n':
		cur.loc.line++
		cur.loc.col = 0
		fallthrough
	case '\t':
		fallthrough
	case ' ':
		return nil, cur, true
	}

	// Syntax that should be kept
	symbols := []symbol{
		eqSymbol,
		neqSymbol,
		neqSymbol2,
		ltSymbol,
		lteSymbol,
		gtSymbol,
		gteSymbol,
		concatSymbol,
		plusSymbol,
		commaSymbol,
		leftParenSymbol,
		rightParenSymbol,
		semicolonSymbol,
		asteriskSymbol,
	}

	var options []string
	for _, s := range symbols {
		options = append(options, string(s))
	}

	// Use `ic`, not `cur`
	match := longestMatch(source, ic, options)
	// Unknown character
	if match == "" {
		return nil, ic, false
	}

	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.col = ic.loc.col + uint(len(match))

	return &token{
		value: match,
		loc:   ic.loc,
		kind:  symbolKind,
	}, cur, true

}

func lexKeyword(source string, ic cursor) (*token, cursor, bool) {
	cur := ic
	keyword := []keyword{
		selectKeyword,
		insertKeyword,
		valuesKeyword,
		asKeyword,
		tableKeyword,
		createKeyword,
		whereKeyword,
		fromKeyword,
		intoKeyword,
		textKeyword,
		trueKeyword,
		falseKeyword,
		nullKeyword,
		intKeyword,
	}

	var options []string
	for _, k := range keyword {
		options = append(options, string(k))
	}

	match := longestMatch(source, ic, options)
	if match == "" {
		return nil, ic, false
	}

	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.col = ic.loc.col + uint(len(match))

	kind := keywordKind
	if match == string(trueKeyword) || match == string(falseKeyword) {
		kind = boolKind
	}

	if match == string(nullKeyword) {
		kind = nullKind
	}

	return &token{
		value: match,
		kind:  kind,
		loc:   ic.loc,
	}, cur, true

}

//	An identifier is either a double-quoted string or a group of characters starting with an alphabetical character
//	and possibly containing numbers and underscores

func lexIdentifier(source string, ic cursor) (*token, cursor, bool) {
	// Handle separately if is a double-quoted identifier
	if token, newCursor, ok := lexCharacterDelimited(source, ic, '"'); ok {
		return token, newCursor, true
	}

	cur := ic

	c := source[cur.pointer]
	// Other characters count too, big ignoring non-ascii for now
	isAlphabetical := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
	if !isAlphabetical {
		return nil, ic, false
	}
	cur.pointer++
	cur.loc.col++

	value := []byte{c}
	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c = source[cur.pointer]

		//Other characters count too, big ignorign non-ascii for now
		isAlphabetical := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
		isNumeric := c >= '0' && c <= '9'
		if isAlphabetical || isNumeric || c == '$' || c == '_' {
			value = append(value, c)
			cur.loc.col++
			continue
		}
		break
	}

	if len(value) == 0 {
		return nil, ic, false
	}

	return &token{
		// Unquioted identifiers are case-insensitive
		value: strings.ToLower(string(value)),
		loc:   ic.loc,
		kind:  identifierKind,
	}, cur, true
}

// lex separa una cadena de entrada en una lista de tokens.
// Este proceso puede ser divido en las siguientes tareas:
//  1. Instanciar un cursor que apunte al principio de la cadena
//  2. Ejecuta todos los lexers en serie.
//  3. Si cualquiera de los lexer genera un token, añade ese token al
//     slice de tokens, updatea el cursor y comienza de nuevo el proceso
//     desde la nueva localizacion del cursor.
func lex(source string) ([]*token, error) {
	tokens := []*token{}
	cur := cursor{}

lex:
	for cur.pointer < uint(len(source)) {
		lexers := []lexer{lexKeyword, lexSymbol, lexString, lexNumeric, lexIdentifier}
		for _, l := range lexers {
			if token, newCursor, ok := l(source, cur); ok {
				cur = newCursor

				if token != nil {
					tokens = append(tokens, token)
				}

				continue lex
			}
		}
		hint := ""
		if len(tokens) > 0 {
			hint = " after " + tokens[len(tokens)-1].value
		}
		return nil, fmt.Errorf("unable to lex token%s, at %d:%d", hint, cur.loc.line, cur.loc.col)
	}

	return tokens, nil
}
