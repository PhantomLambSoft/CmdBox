package prompts

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"charm.land/huh/v2"
	"github.com/charmbracelet/x/term"
)

func PromptForMissingVars(varName string) (string, error) {
	message := fmt.Sprintf("Enter value for <%s>", varName)

	if !term.IsTerminal(os.Stdout.Fd()) {
		return promptPlain(message)
	}

	var value string
	err := huh.NewInput().
		Title(message).
		Value(&value).
		Run()
	if err != nil {
		return "", fmt.Errorf("prompting for %s: %w", varName, err)
	}
	return value, nil
}

func promptPlain(message string) (string, error) {
	fmt.Fprint(os.Stderr, message)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\n"), nil
}
