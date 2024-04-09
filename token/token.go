package token

type Kind int

const (
	Unknown Kind = iota

	Ident

	LBrace
	RBrace
	LBracket
	RBracket
	
	Semicolon
	Colon
	DoubleColon
	Asterisk

	Unset

	Eof
)

type Pos struct {
	Line int
	Column int
}

type Token struct {
	Kind Kind
	Ident string
	Pos Pos
}

func Lookup(ident string) Kind {
	if ident == "unset" {
		return Unset
	}

	return Ident
}
