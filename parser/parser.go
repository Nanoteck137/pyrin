package parser

import (
	"fmt"
	"io"

	"github.com/nanoteck137/pyrin/ast"
	"github.com/nanoteck137/pyrin/lexer"
	"github.com/nanoteck137/pyrin/token"
)

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

type ErrorList []*Error

func (p *ErrorList) Add(message string) {
	*p = append(*p, &Error{
		Message: message,
	})
}

type Parser struct {
	tokenizer *lexer.Tokenizer
	token     token.Token
	errors    ErrorList
}

func New(reader io.Reader) *Parser {
	tokenizer := lexer.New(reader)

	return &Parser{
		tokenizer: tokenizer,
		token:     tokenizer.NextToken(),
	}
}

func (p *Parser) error(message string) {
	p.errors.Add(message)
}

func (p *Parser) next() {
	p.token = p.tokenizer.NextToken()
}

func (p *Parser) expect(token token.Kind) {
	if p.token.Kind != token {
		p.error(fmt.Sprintf("Expected token: %v got %v", token, p.token.Kind))
	}

	p.next()
}

func (p *Parser) parseTypespec() ast.Typespec {
	switch p.token.Kind {
	case token.Ident:
		ident := p.token.Ident
		p.next()
		return &ast.IdentTypespec{
			Ident: ident,
		}
	case token.LBracket:
		p.next()
		p.expect(token.RBracket)

		elementTy := p.parseTypespec()

		return &ast.ArrayTypespec{
			Element: elementTy,
		}
	}

	p.error(fmt.Sprintf("Unknown token for type: %v", p.token.Kind))
	return nil
}

func (p *Parser) parseStructField() *ast.Field {
	name := p.token.Ident
	p.expect(token.Ident)

	var ty ast.Typespec
	unset := false
	if p.token.Kind == token.Unset {
		p.next()
		unset = true
	} else {
		ty = p.parseTypespec()
	}

	p.expect(token.Semicolon)
	return &ast.Field{
		Name:  name,
		Type:  ty,
		Unset: unset,
	}
}

func (p *Parser) parseStructDecl() *ast.StructDecl {
	name := p.token.Ident
	p.next()

	extend := ""

	if p.token.Kind == token.DoubleColon {
		p.next()

		extend = p.token.Ident
		p.expect(token.Ident)
	}

	p.expect(token.LBrace)

	var fields []*ast.Field
	for p.token.Kind == token.Ident {
		field := p.parseStructField()
		fields = append(fields, field)
	}

	p.expect(token.RBrace)

	return &ast.StructDecl{
		Name:   name,
		Extend: extend,
		Fields: fields,
	}
}

func (p *Parser) ParseDecl() ast.Decl {
	switch p.token.Kind {
	case token.Ident:
		return p.parseStructDecl()
	default:
		// TODO(patrik): Remove
		panic("Unknown decl")
	}

	return nil
}

func (p *Parser) Parse() []ast.Decl {
	var decls []ast.Decl
	for p.token.Kind != token.Eof {
		decl := p.ParseDecl()
		decls = append(decls, decl)
	}

	return decls
}
