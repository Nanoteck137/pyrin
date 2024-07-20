package util

import (
	"fmt"
	"io"

	"github.com/nanoteck137/pyrin/resolve"
)

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
	Writer io.Writer
	indent int
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
		_, err := w.Writer.Write(([]byte)("  "))
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
