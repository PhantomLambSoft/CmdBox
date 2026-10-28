package prompts

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/huh/v2"
	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/charmbracelet/x/term"
)

type Prompter struct {
	theme  huh.Theme
	reader *bufio.Reader
}

func NewPrompter(p ui.Palette, useColor bool) *Prompter {
	var theme huh.Theme = huh.ThemeFunc(huh.ThemeBase)
	if useColor {
		theme = ui.HuhTheme(p)
	}
	return &Prompter{
		theme:  theme,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (p *Prompter) PromptForMissingVars(varName string) (string, error) {
	message := fmt.Sprintf("Enter value for <%s>", varName)

	if !isInteractive() {
		return p.promptPlain(message)
	}

	var value string
	err := huh.NewInput().
		Title(message).
		Value(&value).
		WithTheme(p.theme).
		Run()
	if err != nil {
		return "", fmt.Errorf("prompting for %s: %w", varName, err)
	}
	return value, nil
}

func (p *Prompter) PromptConfirmation(message string) (bool, error) {
	var confirmed bool
	err := huh.NewConfirm().Title(message).Value(&confirmed).Run()
	if err != nil {
		return false, err
	}
	return confirmed, nil
}

func (p *Prompter) promptPlain(message string) (string, error) {
	fmt.Fprint(os.Stderr, message+": ")
	line, err := p.reader.ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && line != "") {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func isInteractive() bool {
	return term.IsTerminal(os.Stdin.Fd()) && term.IsTerminal(os.Stdout.Fd())
}
