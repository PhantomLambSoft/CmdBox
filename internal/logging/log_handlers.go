package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

var redactKeys = map[string]bool{
	"token":    true,
	"password": true,
}

func levelReplacer(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		if level, ok := a.Value.Any().(slog.Level); ok && level == LevelCritical {
			a.Value = slog.StringValue("CRITICAL")
		}
	}
	return a
}

type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var err error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if e := handler.Handle(ctx, r.Clone()); e != nil {
				err = e
			}
		}
	}
	return err
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		out[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: out}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	out := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		out[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: out}
}

type redactionHandler struct {
	next slog.Handler
}

func (h *redactionHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *redactionHandler) Handle(ctx context.Context, r slog.Record) error {
	redacted := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		if redactKeys[strings.ToLower(a.Key)] {
			a.Value = slog.StringValue("[REDACTED]")
		}
		redacted.AddAttrs(a)
		return true
	})
	return h.next.Handle(ctx, redacted)
}

func (h *redactionHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redacted := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		if redactKeys[strings.ToLower(a.Key)] {
			a.Value = slog.StringValue("[REDACTED]")
		}
		redacted[i] = a
	}
	return &redactionHandler{next: h.next.WithAttrs(redacted)}
}

func (h *redactionHandler) WithGroup(name string) slog.Handler {
	return &redactionHandler{next: h.next.WithGroup(name)}
}

type runIDHandler struct {
	next  slog.Handler
	runID string
}

func (h *runIDHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *runIDHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(slog.String("run_id", h.runID))
	return h.next.Handle(ctx, r)
}

func (h *runIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &runIDHandler{next: h.next.WithAttrs(attrs), runID: h.runID}
}

func (h *runIDHandler) WithGroup(name string) slog.Handler {
	return &runIDHandler{next: h.next.WithGroup(name), runID: h.runID}
}

func ConfigureLogging(cfg LogConfig, runID string) *slog.Logger {
	consoleHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:       cfg.ConsoleLevel,
		ReplaceAttr: levelReplacer,
	})

	handlers := []slog.Handler{consoleHandler}

	if cfg.FileEnabled {
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    maxSizeMB(cfg.MaxBytes),
			MaxBackups: cfg.Backups,
		}
		fileHandler := slog.NewTextHandler(fileWriter, &slog.HandlerOptions{
			Level:       cfg.FileLevel,
			ReplaceAttr: levelReplacer,
		})
		handlers = append(handlers, fileHandler)
	}

	var handler slog.Handler = &multiHandler{handlers: handlers}
	handler = &redactionHandler{next: handler}
	handler = &runIDHandler{next: handler, runID: runID}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func maxSizeMB(mb int) int {
	if mb < 1 {
		return 1
	}
	return mb
}
