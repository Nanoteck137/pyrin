package trail

import (
	"io"
	"log/slog"
	"os"

	"github.com/golang-cz/devslog"
)

type Logger struct {
	*slog.Logger
}

type Options struct {
	Debug bool
	Level slog.Level
	Out   io.Writer
}

func NewLogger(opts *Options) *Logger {
	if opts.Out == nil {
		opts.Out = os.Stderr
	}

	handlerOptions := &slog.HandlerOptions{Level: opts.Level}
	var handler slog.Handler
	if opts.Debug {
		handler = devslog.NewHandler(opts.Out, &devslog.Options{HandlerOptions: handlerOptions})
	} else {
		handler = slog.NewJSONHandler(opts.Out, handlerOptions)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}
