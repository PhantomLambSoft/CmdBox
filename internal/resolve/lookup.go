package resolve

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/repository"
)

// Lookup defines methods for retrieving command and variable records by their respective identifiers.
type Lookup interface {
	GetCommand(alias string) (CommandRecord, error)
	GetVariable(name string) (VariableRecord, error)
}

type RepoLookup struct {
	cmdRepo repository.CommandRepository
	varRepo repository.VariableRepository
}

func NewLookup(cmdRepo repository.CommandRepository, varRepo repository.VariableRepository) Lookup {
	return &RepoLookup{cmdRepo, varRepo}
}

func (rl *RepoLookup) GetCommand(alias string) (CommandRecord, error) {
	cmd, err := rl.cmdRepo.GetByAlias(alias, nil)
	if err != nil {
		return CommandRecord{}, fmt.Errorf("looking up command alias: %w", err)
	}
	return CommandRecord{cmd.Alias, cmd.Template}, nil
}

func (rl *RepoLookup) GetVariable(name string) (VariableRecord, error) {
	variable, err := rl.varRepo.GetByName(name, nil)
	if err != nil {
		return VariableRecord{}, fmt.Errorf("looking up variable name: %w", err)
	}
	return VariableRecord{variable.Name, variable.Value}, nil
}

// MemoizedLookup is a caching layer that wraps a Lookup to store command and variable records for faster access.
// The cmdCache stores mappings between command aliases and their records.
// The varCache stores mappings between variable names and their records.
// The inner field refers to the actual Lookup implementation being wrapped and relied upon when a cache miss occurs.
type MemoizedLookup struct {
	inner    Lookup
	cmdCache map[string]*CommandRecord
	varCache map[string]*VariableRecord
}

func NewMemoizedLookup(inner Lookup) Lookup {
	return &MemoizedLookup{
		inner,
		make(map[string]*CommandRecord),
		make(map[string]*VariableRecord),
	}
}

func (ml *MemoizedLookup) GetCommand(alias string) (CommandRecord, error) {
	value, ok := ml.cmdCache[alias]
	if ok {
		return *value, nil
	}
	cmd, err := ml.inner.GetCommand(alias)
	if err != nil {
		return CommandRecord{}, fmt.Errorf("looking up command alias: %w", err)
	}
	ml.cmdCache[alias] = &cmd
	return cmd, nil
}

func (ml *MemoizedLookup) GetVariable(name string) (VariableRecord, error) {
	value, ok := ml.varCache[name]
	if ok {
		return *value, nil
	}
	variable, err := ml.inner.GetVariable(name)
	if err != nil {
		return VariableRecord{}, fmt.Errorf("looking up variable name: %w", err)
	}
	ml.varCache[name] = &variable
	return variable, nil
}

func (ml *MemoizedLookup) Clear() {
	ml.cmdCache = make(map[string]*CommandRecord)
	ml.varCache = make(map[string]*VariableRecord)
}
