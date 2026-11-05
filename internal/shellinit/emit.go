package shellinit

import (
	"io"
	"io/fs"
	"os"
	"strings"
)

const EmitFileEnv = "CMDBOX_EMIT_FILE"

func WriteEmit(command string, stdout io.Writer, getEnv func(string) string, writeFile func(string, []byte, fs.FileMode) error) error {
	text := strings.Trim(command, "\n") + "\n"
	if path := getEnv(EmitFileEnv); path != "" {
		return writeFile(path, []byte(text), 0o600)
	}
	_, err := io.WriteString(stdout, text)
	return err
}

func Emit(command string) error {
	return WriteEmit(command, os.Stdout, os.Getenv, os.WriteFile)
}
