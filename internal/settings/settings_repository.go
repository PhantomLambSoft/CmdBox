package settings

import (
	"bytes"
	"os"

	"github.com/PhantomLambSoft/CmdBox/internal/atomicfile"
	"github.com/pelletier/go-toml/v2"
)

type Repository struct {
	path string
}

func NewSettingsRepository(path string) *Repository {
	return &Repository{
		path: path,
	}
}

func (r *Repository) Load(into *Settings) error {
	data, err := os.ReadFile(r.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return toml.NewDecoder(bytes.NewReader(data)).EnableUnmarshalerInterface().Decode(into)
}

func (r *Repository) Save(settings Settings) error {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).EnableMarshalerInterface().Encode(settings); err != nil {
		return err
	}
	return atomicfile.WriteFile(r.path, buf.String())
}
