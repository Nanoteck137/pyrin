package utils

import (
	"fmt"
	"io"
	"strings"

	"github.com/nanoteck137/pyrin/tools/resolve"
)

type NameMapping func(name string) string
type ReplacementFunc func(name string) string

func ReplacePathArgs(path string, nameMapping NameMapping, replacementFunc ReplacementFunc) (string, []string) {
	var args []string
	parts := strings.Split(path, "/")

	for i, p := range parts {
		if len(p) == 0 {
			continue
		}

		if p[0] == ':' {
			name := p[1:]
			if nameMapping != nil {
				name = nameMapping(name)
			}
			
			args = append(args, name)

			parts[i] = replacementFunc(name)
		}
	}

	newPath := strings.Join(parts, "/")

	return newPath, args
}

func TypeToString(ty resolve.Type) (string, error) {
	switch ty := ty.(type) {
	case *resolve.TypeString:
		return "string", nil
	case *resolve.TypeInt:
		return "int", nil
	case *resolve.TypeBoolean:
		return "bool", nil
	case *resolve.TypeArray:
		s, err := TypeToString(ty.ElementType)
		if err != nil {
			return "", err
		}
		return "[]" + s, nil
	case *resolve.TypePtr:
		s, err := TypeToString(ty.BaseType)
		if err != nil {
			return "", err
		}
		return "*" + s, nil
	case *resolve.TypeStruct:
		return ty.Name, nil
	case *resolve.TypeSameStruct:
		return ty.Type.Name, nil
	default:
		return "", fmt.Errorf("Unknown resolved type: %T", ty)
	}
}

type CodeWriter struct {
	Writer    io.Writer
	IndentStr string
	indent    int
}

func (w *CodeWriter) Indent() {
	w.indent += 1
}

func (w *CodeWriter) Unindent() {
	if w.indent == 0 {
		return
	}

	w.indent -= 1
}

func (w *CodeWriter) WriteIndent() error {
	for i := 0; i < w.indent; i++ {
		_, err := w.Writer.Write(([]byte)(w.IndentStr))
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *CodeWriter) Writef(format string, a ...any) error {
	_, err := fmt.Fprintf(w.Writer, format, a...)
	return err
}

func (w *CodeWriter) IWritef(format string, a ...any) error {
	err := w.WriteIndent()
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w.Writer, format, a...)

	return err
}
