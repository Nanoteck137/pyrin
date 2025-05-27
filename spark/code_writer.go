package spark

import (
	"fmt"
	"io"
)

type CodeWriter struct {
	io.Writer

	indentStr string
	indent    int
}

func NewCodeWriter(w io.Writer, indentStr string) CodeWriter {
	return CodeWriter{
		Writer:    w,
		indentStr: indentStr,
		indent:    0,
	}
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
		_, err := w.Write(([]byte)(w.indentStr))
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

func (w *CodeWriter) IndentWritef(format string, a ...any) error {
	err := w.WriteIndent()
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w.Writer, format, a...)
	return err
}
