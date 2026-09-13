package logging

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// recordingHandler is a minimal slog.Handler used to observe how multiHandler,
// redactionHandler, and runIDHandler drive an underlying handler.
type recordingHandler struct {
	enabledFn func(level slog.Level) bool
	handleErr error

	handled        []slog.Record
	withAttrsCalls [][]slog.Attr
	withGroupCalls []string
}

func (h *recordingHandler) Enabled(_ context.Context, level slog.Level) bool {
	if h.enabledFn == nil {
		return true
	}
	return h.enabledFn(level)
}

func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	h.handled = append(h.handled, r)
	return h.handleErr
}

func (h *recordingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.withAttrsCalls = append(h.withAttrsCalls, attrs)
	return h
}

func (h *recordingHandler) WithGroup(name string) slog.Handler {
	h.withGroupCalls = append(h.withGroupCalls, name)
	return h
}

func attrsOf(t *testing.T, r slog.Record) map[string]slog.Value {
	t.Helper()
	out := make(map[string]slog.Value)
	r.Attrs(func(a slog.Attr) bool {
		out[a.Key] = a.Value
		return true
	})
	return out
}

// --- levelReplacer ---

func TestLevelReplacer(t *testing.T) {
	t.Run("rewrites the critical level to a CRITICAL string", func(t *testing.T) {
		a := slog.Attr{Key: slog.LevelKey, Value: slog.AnyValue(LevelCritical)}
		got := levelReplacer(nil, a)
		if got.Value.Kind() != slog.KindString || got.Value.String() != "CRITICAL" {
			t.Fatalf("levelReplacer() = %+v, want string value CRITICAL", got)
		}
	})

	t.Run("leaves other levels untouched", func(t *testing.T) {
		a := slog.Attr{Key: slog.LevelKey, Value: slog.AnyValue(slog.LevelWarn)}
		got := levelReplacer(nil, a)
		lvl, ok := got.Value.Any().(slog.Level)
		if !ok || lvl != slog.LevelWarn {
			t.Fatalf("levelReplacer() = %+v, want unchanged LevelWarn", got)
		}
	})

	t.Run("leaves non-level keys untouched", func(t *testing.T) {
		a := slog.String("msg", "hello")
		got := levelReplacer(nil, a)
		if got.Key != a.Key || !got.Value.Equal(a.Value) {
			t.Fatalf("levelReplacer() = %+v, want unchanged %+v", got, a)
		}
	})

	t.Run("leaves the level key untouched when its value isn't a slog.Level", func(t *testing.T) {
		a := slog.Attr{Key: slog.LevelKey, Value: slog.StringValue("not-a-level")}
		got := levelReplacer(nil, a)
		if got.Key != a.Key || !got.Value.Equal(a.Value) {
			t.Fatalf("levelReplacer() = %+v, want unchanged %+v", got, a)
		}
	})
}

// --- multiHandler ---

func TestMultiHandlerEnabled(t *testing.T) {
	t.Run("false with no handlers", func(t *testing.T) {
		h := &multiHandler{}
		if h.Enabled(context.Background(), slog.LevelInfo) {
			t.Fatalf("Enabled() = true, want false")
		}
	})

	t.Run("false when every handler is disabled", func(t *testing.T) {
		h := &multiHandler{handlers: []slog.Handler{
			&recordingHandler{enabledFn: func(slog.Level) bool { return false }},
			&recordingHandler{enabledFn: func(slog.Level) bool { return false }},
		}}
		if h.Enabled(context.Background(), slog.LevelInfo) {
			t.Fatalf("Enabled() = true, want false")
		}
	})

	t.Run("true when at least one handler is enabled", func(t *testing.T) {
		h := &multiHandler{handlers: []slog.Handler{
			&recordingHandler{enabledFn: func(slog.Level) bool { return false }},
			&recordingHandler{enabledFn: func(slog.Level) bool { return true }},
		}}
		if !h.Enabled(context.Background(), slog.LevelInfo) {
			t.Fatalf("Enabled() = false, want true")
		}
	})
}

func TestMultiHandlerHandle(t *testing.T) {
	t.Run("dispatches only to enabled handlers", func(t *testing.T) {
		enabled := &recordingHandler{enabledFn: func(slog.Level) bool { return true }}
		disabled := &recordingHandler{enabledFn: func(slog.Level) bool { return false }}
		h := &multiHandler{handlers: []slog.Handler{enabled, disabled}}

		r := slog.NewRecord(time.Now(), slog.LevelInfo, "hi", 0)
		if err := h.Handle(context.Background(), r); err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if len(enabled.handled) != 1 {
			t.Fatalf("enabled handler received %d records, want 1", len(enabled.handled))
		}
		if len(disabled.handled) != 0 {
			t.Fatalf("disabled handler received %d records, want 0", len(disabled.handled))
		}
	})

	t.Run("returns the last error from an enabled handler and still calls the rest", func(t *testing.T) {
		errFirst := errors.New("first failure")
		errSecond := errors.New("second failure")
		h1 := &recordingHandler{handleErr: errFirst}
		h2 := &recordingHandler{handleErr: errSecond}
		h3 := &recordingHandler{}
		h := &multiHandler{handlers: []slog.Handler{h1, h2, h3}}

		r := slog.NewRecord(time.Now(), slog.LevelInfo, "hi", 0)
		err := h.Handle(context.Background(), r)
		if !errors.Is(err, errSecond) {
			t.Fatalf("Handle() error = %v, want %v", err, errSecond)
		}
		if len(h1.handled) != 1 || len(h2.handled) != 1 || len(h3.handled) != 1 {
			t.Fatalf("expected every handler to be invoked despite errors: %d %d %d",
				len(h1.handled), len(h2.handled), len(h3.handled))
		}
	})

	t.Run("nil when no handler is enabled", func(t *testing.T) {
		h := &multiHandler{handlers: []slog.Handler{
			&recordingHandler{enabledFn: func(slog.Level) bool { return false }, handleErr: errors.New("should not run")},
		}}
		r := slog.NewRecord(time.Now(), slog.LevelInfo, "hi", 0)
		if err := h.Handle(context.Background(), r); err != nil {
			t.Fatalf("Handle() error = %v, want nil", err)
		}
	})

	t.Run("each handler gets an isolated clone of the record", func(t *testing.T) {
		h1 := &recordingHandler{}
		h2 := &recordingHandler{}
		h := &multiHandler{handlers: []slog.Handler{h1, h2}}

		r := slog.NewRecord(time.Now(), slog.LevelInfo, "hi", 0)
		r.AddAttrs(slog.String("shared", "yes"))
		if err := h.Handle(context.Background(), r); err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		// Mutate what h1 received; h2's copy must be unaffected.
		h1.handled[0].AddAttrs(slog.String("only_h1", "yes"))

		if _, ok := attrsOf(t, h2.handled[0])["only_h1"]; ok {
			t.Fatalf("mutating one handler's record leaked into another handler's clone")
		}
	})
}

func TestMultiHandlerWithAttrs(t *testing.T) {
	h1 := &recordingHandler{}
	h2 := &recordingHandler{}
	h := &multiHandler{handlers: []slog.Handler{h1, h2}}

	attrs := []slog.Attr{slog.String("k", "v")}
	got := h.WithAttrs(attrs)

	mh, ok := got.(*multiHandler)
	if !ok {
		t.Fatalf("WithAttrs() returned %T, want *multiHandler", got)
	}
	if len(mh.handlers) != 2 {
		t.Fatalf("WithAttrs() produced %d handlers, want 2", len(mh.handlers))
	}
	if len(h1.withAttrsCalls) != 1 || len(h2.withAttrsCalls) != 1 {
		t.Fatalf("expected WithAttrs to be forwarded to every handler")
	}
}

func TestMultiHandlerWithGroup(t *testing.T) {
	h1 := &recordingHandler{}
	h2 := &recordingHandler{}
	h := &multiHandler{handlers: []slog.Handler{h1, h2}}

	got := h.WithGroup("g")

	mh, ok := got.(*multiHandler)
	if !ok {
		t.Fatalf("WithGroup() returned %T, want *multiHandler", got)
	}
	if len(mh.handlers) != 2 {
		t.Fatalf("WithGroup() produced %d handlers, want 2", len(mh.handlers))
	}
	if len(h1.withGroupCalls) != 1 || h1.withGroupCalls[0] != "g" {
		t.Fatalf("expected WithGroup(%q) to be forwarded to h1, got %v", "g", h1.withGroupCalls)
	}
	if len(h2.withGroupCalls) != 1 || h2.withGroupCalls[0] != "g" {
		t.Fatalf("expected WithGroup(%q) to be forwarded to h2, got %v", "g", h2.withGroupCalls)
	}
}

// --- redactionHandler ---

func TestRedactionHandlerEnabled(t *testing.T) {
	next := &recordingHandler{enabledFn: func(level slog.Level) bool { return level >= slog.LevelWarn }}
	h := &redactionHandler{next: next}

	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatalf("Enabled(Info) = true, want false")
	}
	if !h.Enabled(context.Background(), slog.LevelWarn) {
		t.Fatalf("Enabled(Warn) = false, want true")
	}
}

func TestRedactionHandlerHandle(t *testing.T) {
	t.Run("redacts configured keys case-insensitively", func(t *testing.T) {
		next := &recordingHandler{}
		h := &redactionHandler{next: next}

		r := slog.NewRecord(time.Now(), slog.LevelInfo, "hi", 0)
		r.AddAttrs(slog.String("Token", "abc123"), slog.String("PASSWORD", "hunter2"), slog.String("user", "alice"))

		if err := h.Handle(context.Background(), r); err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
		if len(next.handled) != 1 {
			t.Fatalf("next handler received %d records, want 1", len(next.handled))
		}

		got := attrsOf(t, next.handled[0])
		if got["Token"].String() != "[REDACTED]" {
			t.Fatalf("Token = %v, want [REDACTED]", got["Token"])
		}
		if got["PASSWORD"].String() != "[REDACTED]" {
			t.Fatalf("PASSWORD = %v, want [REDACTED]", got["PASSWORD"])
		}
		if got["user"].String() != "alice" {
			t.Fatalf("user = %v, want unchanged alice", got["user"])
		}
	})

	t.Run("preserves time, level, and message", func(t *testing.T) {
		next := &recordingHandler{}
		h := &redactionHandler{next: next}

		when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		r := slog.NewRecord(when, slog.LevelError, "boom", 0)

		if err := h.Handle(context.Background(), r); err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		got := next.handled[0]
		if !got.Time.Equal(when) {
			t.Fatalf("Time = %v, want %v", got.Time, when)
		}
		if got.Level != slog.LevelError {
			t.Fatalf("Level = %v, want %v", got.Level, slog.LevelError)
		}
		if got.Message != "boom" {
			t.Fatalf("Message = %q, want %q", got.Message, "boom")
		}
	})

	t.Run("propagates the next handler's error", func(t *testing.T) {
		wantErr := errors.New("write failed")
		next := &recordingHandler{handleErr: wantErr}
		h := &redactionHandler{next: next}

		r := slog.NewRecord(time.Now(), slog.LevelInfo, "hi", 0)
		if err := h.Handle(context.Background(), r); !errors.Is(err, wantErr) {
			t.Fatalf("Handle() error = %v, want %v", err, wantErr)
		}
	})
}

func TestRedactionHandlerWithAttrs(t *testing.T) {
	t.Run("redacts matching bound attrs before forwarding", func(t *testing.T) {
		next := &recordingHandler{}
		h := &redactionHandler{next: next}

		got := h.WithAttrs([]slog.Attr{slog.String("password", "hunter2"), slog.String("user", "alice")})

		rh, ok := got.(*redactionHandler)
		if !ok {
			t.Fatalf("WithAttrs() returned %T, want *redactionHandler", got)
		}
		if rh.next != next {
			t.Fatalf("WithAttrs() did not wrap the same next handler")
		}
		if len(next.withAttrsCalls) != 1 {
			t.Fatalf("next.WithAttrs called %d times, want 1", len(next.withAttrsCalls))
		}

		forwarded := next.withAttrsCalls[0]
		byKey := make(map[string]string, len(forwarded))
		for _, a := range forwarded {
			byKey[a.Key] = a.Value.String()
		}
		if byKey["password"] != "[REDACTED]" {
			t.Fatalf("forwarded password = %q, want [REDACTED]", byKey["password"])
		}
		if byKey["user"] != "alice" {
			t.Fatalf("forwarded user = %q, want unchanged alice", byKey["user"])
		}
	})

	t.Run("does not mutate the caller's attr slice", func(t *testing.T) {
		next := &recordingHandler{}
		h := &redactionHandler{next: next}

		attrs := []slog.Attr{slog.String("token", "secret")}
		h.WithAttrs(attrs)

		if attrs[0].Value.String() != "secret" {
			t.Fatalf("caller's attrs slice was mutated: %v", attrs)
		}
	})
}

func TestRedactionHandlerWithGroup(t *testing.T) {
	next := &recordingHandler{}
	h := &redactionHandler{next: next}

	got := h.WithGroup("g")

	rh, ok := got.(*redactionHandler)
	if !ok {
		t.Fatalf("WithGroup() returned %T, want *redactionHandler", got)
	}
	if rh.next != next {
		t.Fatalf("WithGroup() did not wrap the same next handler")
	}
	if len(next.withGroupCalls) != 1 || next.withGroupCalls[0] != "g" {
		t.Fatalf("expected WithGroup(%q) forwarded to next, got %v", "g", next.withGroupCalls)
	}
}

// --- runIDHandler ---

func TestRunIDHandlerEnabled(t *testing.T) {
	next := &recordingHandler{enabledFn: func(level slog.Level) bool { return level >= slog.LevelError }}
	h := &runIDHandler{next: next, runID: "run-1"}

	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatalf("Enabled(Info) = true, want false")
	}
	if !h.Enabled(context.Background(), slog.LevelError) {
		t.Fatalf("Enabled(Error) = false, want true")
	}
}

func TestRunIDHandlerHandle(t *testing.T) {
	next := &recordingHandler{}
	h := &runIDHandler{next: next, runID: "run-42"}

	r := slog.NewRecord(time.Now(), slog.LevelInfo, "hi", 0)
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if len(next.handled) != 1 {
		t.Fatalf("next handler received %d records, want 1", len(next.handled))
	}
	got := attrsOf(t, next.handled[0])
	if got["run_id"].String() != "run-42" {
		t.Fatalf("run_id = %v, want run-42", got["run_id"])
	}
}

func TestRunIDHandlerWithAttrs(t *testing.T) {
	next := &recordingHandler{}
	h := &runIDHandler{next: next, runID: "run-42"}

	got := h.WithAttrs([]slog.Attr{slog.String("k", "v")})

	rh, ok := got.(*runIDHandler)
	if !ok {
		t.Fatalf("WithAttrs() returned %T, want *runIDHandler", got)
	}
	if rh.runID != "run-42" {
		t.Fatalf("runID = %q, want run-42", rh.runID)
	}
	if len(next.withAttrsCalls) != 1 {
		t.Fatalf("next.WithAttrs called %d times, want 1", len(next.withAttrsCalls))
	}
}

func TestRunIDHandlerWithGroup(t *testing.T) {
	next := &recordingHandler{}
	h := &runIDHandler{next: next, runID: "run-42"}

	got := h.WithGroup("g")

	rh, ok := got.(*runIDHandler)
	if !ok {
		t.Fatalf("WithGroup() returned %T, want *runIDHandler", got)
	}
	if rh.runID != "run-42" {
		t.Fatalf("runID = %q, want run-42", rh.runID)
	}
	if len(next.withGroupCalls) != 1 || next.withGroupCalls[0] != "g" {
		t.Fatalf("expected WithGroup(%q) forwarded to next, got %v", "g", next.withGroupCalls)
	}
}

// --- maxSizeMB ---

func TestMaxSizeMB(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"negative clamps to 1", -5, 1},
		{"zero clamps to 1", 0, 1},
		{"one is unchanged", 1, 1},
		{"large value is unchanged", 500, 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maxSizeMB(tt.in); got != tt.want {
				t.Fatalf("maxSizeMB(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

// --- ConfigureLogging (integration) ---

// withCapturedStderr redirects os.Stderr for the duration of fn and returns what was written to it.
func withCapturedStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer: %v", err)
	}
	os.Stderr = orig

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading captured stderr: %v", err)
	}
	return buf.String()
}

// restoreDefaultLogger saves and restores slog's package-level default logger, since
// ConfigureLogging calls slog.SetDefault and would otherwise leak state across tests.
func restoreDefaultLogger(t *testing.T) {
	t.Helper()
	orig := slog.Default()
	t.Cleanup(func() { slog.SetDefault(orig) })
}

func TestConfigureLoggingConsoleOnly(t *testing.T) {
	restoreDefaultLogger(t)

	cfg := LogConfig{
		ConsoleLevel: slog.LevelWarn,
		FileEnabled:  false,
	}

	var logger *slog.Logger
	output := withCapturedStderr(t, func() {
		logger = ConfigureLogging(cfg, "run-abc")
		logger.Info("hidden info", "password", "hunter2")
		logger.Warn("visible warning", "password", "hunter2", "user", "alice")
	})

	if logger == nil {
		t.Fatalf("ConfigureLogging() returned nil logger")
	}
	if slog.Default() != logger {
		t.Fatalf("ConfigureLogging() did not set slog.Default() to the returned logger")
	}

	if strings.Contains(output, "hidden info") {
		t.Fatalf("output contains a record below ConsoleLevel: %q", output)
	}
	if !strings.Contains(output, "visible warning") {
		t.Fatalf("output missing the record at ConsoleLevel: %q", output)
	}
	if !strings.Contains(output, "run_id=run-abc") {
		t.Fatalf("output missing run_id attr: %q", output)
	}
	if strings.Contains(output, "hunter2") {
		t.Fatalf("output leaked an unredacted secret: %q", output)
	}
	if !strings.Contains(output, "password=[REDACTED]") {
		t.Fatalf("output missing redacted password attr: %q", output)
	}
	if !strings.Contains(output, "user=alice") {
		t.Fatalf("output missing unredacted user attr: %q", output)
	}
}

func TestConfigureLoggingCriticalLevel(t *testing.T) {
	restoreDefaultLogger(t)

	cfg := LogConfig{ConsoleLevel: slog.LevelWarn, FileEnabled: false}

	var logger *slog.Logger
	output := withCapturedStderr(t, func() {
		logger = ConfigureLogging(cfg, "run-crit")
		logger.Log(context.Background(), LevelCritical, "meltdown")
	})

	if !strings.Contains(output, "level=CRITICAL") {
		t.Fatalf("output missing rewritten CRITICAL level: %q", output)
	}
}

func TestConfigureLoggingFileEnabled(t *testing.T) {
	restoreDefaultLogger(t)

	// Deliberately not t.TempDir(): lumberjack keeps the file open for the life of the
	// process (ConfigureLogging never exposes a Close), which makes Windows refuse to
	// remove it and fails t.TempDir()'s automatic cleanup. Best-effort remove instead.
	dir, err := os.MkdirTemp("", "cmdbox-log-handlers-test")
	if err != nil {
		t.Fatalf("os.MkdirTemp() error = %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	logPath := filepath.Join(dir, "cmdbox.log")
	cfg := LogConfig{
		ConsoleLevel: LevelCritical + 1, // silence the console entirely
		FileEnabled:  true,
		FileLevel:    slog.LevelDebug,
		FilePath:     logPath,
		MaxBytes:     10,
		Backups:      1,
	}

	_ = withCapturedStderr(t, func() {
		logger := ConfigureLogging(cfg, "run-file")
		logger.Info("to the file", "token", "abc123", "user", "bob")
	})

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "to the file") {
		t.Fatalf("file missing log message: %q", content)
	}
	if !strings.Contains(content, "run_id=run-file") {
		t.Fatalf("file missing run_id attr: %q", content)
	}
	if strings.Contains(content, "abc123") {
		t.Fatalf("file leaked an unredacted secret: %q", content)
	}
	if !strings.Contains(content, "token=[REDACTED]") {
		t.Fatalf("file missing redacted token attr: %q", content)
	}
	if !strings.Contains(content, "user=bob") {
		t.Fatalf("file missing unredacted user attr: %q", content)
	}
}

func TestConfigureLoggingFileDisabledWritesNoFile(t *testing.T) {
	restoreDefaultLogger(t)

	dir, err := os.MkdirTemp("", "cmdbox-log-handlers-test")
	if err != nil {
		t.Fatalf("os.MkdirTemp() error = %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	logPath := filepath.Join(dir, "cmdbox.log")
	cfg := LogConfig{
		ConsoleLevel: slog.LevelDebug,
		FileEnabled:  false,
		FilePath:     logPath,
	}

	_ = withCapturedStderr(t, func() {
		logger := ConfigureLogging(cfg, "run-nofile")
		logger.Info("console only")
	})

	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Fatalf("expected no log file to be created, stat err = %v", err)
	}
}

func TestConfigureLoggingRespectsIndependentConsoleAndFileLevels(t *testing.T) {
	restoreDefaultLogger(t)

	dir, err := os.MkdirTemp("", "cmdbox-log-handlers-test")
	if err != nil {
		t.Fatalf("os.MkdirTemp() error = %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	logPath := filepath.Join(dir, "cmdbox.log")
	cfg := LogConfig{
		ConsoleLevel: slog.LevelError, // console: errors only
		FileEnabled:  true,
		FileLevel:    slog.LevelDebug, // file: everything
		FilePath:     logPath,
		MaxBytes:     10,
		Backups:      1,
	}

	consoleOutput := withCapturedStderr(t, func() {
		logger := ConfigureLogging(cfg, "run-split")
		logger.Debug("debug detail")
	})

	if strings.Contains(consoleOutput, "debug detail") {
		t.Fatalf("console output should have suppressed the debug record: %q", consoleOutput)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	if !strings.Contains(string(data), "debug detail") {
		t.Fatalf("file should have recorded the debug record: %q", data)
	}
}
