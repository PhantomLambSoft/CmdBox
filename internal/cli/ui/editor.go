package ui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var ErrEditCanceled = errors.New("edit canceled")

type editor struct {
	ta       textarea.Model
	palette  Palette
	title    string
	sized    bool
	canceled bool
}

func newEditor(initial, title string, palette Palette) editor {
	ta := textarea.New()
	ta.SetValue(initial)
	ta.Focus()
	return editor{ta: ta, palette: palette, title: title}
}

func (e editor) Init() tea.Cmd {
	return nil
}

func (e editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+s":
			return e, tea.Quit
		case "esc":
			e.canceled = true
			return e, tea.Quit
		}
	case tea.WindowSizeMsg:
		e.ta.SetWidth(msg.Width - 2)
		e.ta.SetHeight(msg.Height - 4) // leave room for header
		if !e.sized {
			e.ta.MoveToBegin()
			e.sized = true
		}
	}

	var cmd tea.Cmd
	e.ta, cmd = e.ta.Update(msg)
	return e, cmd
}

//func (e editor) View() tea.View {
//	header := e.palette.PanelTitle.Render(e.title) + "\n" + e.palette.Muted.Render("Ctrl+S save - Esc cancel") + "\n\n"
//	v := tea.NewView(header + e.ta.View())
//	v.AltScreen = true
//	return v
//}

func (e editor) View() tea.View {
	width := GetTerminalWidth()

	instructions := e.palette.Muted.Render("Ctrl+S save, Esc cancel")
	header := lipgloss.NewStyle().
		Width(width - 4).
		Align(lipgloss.Center).
		Render(e.palette.PanelTitle.Render(e.title) + "\n" + instructions)

	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(e.palette.Border).
		Padding(1, 1).
		Width(width - 4).
		Render(header + "\n\n" + e.ta.View())

	v := tea.NewView(box)
	v.AltScreen = true
	return v
}

func EditTextFullscreen(initial, title string, palette Palette) (string, error) {
	e := newEditor(initial, title, palette)
	final, err := tea.NewProgram(e).Run()
	if err != nil {
		return "", err
	}

	result, ok := final.(editor)
	if !ok {
		return "", errors.New("unexpected model type returned from editor program")
	}
	if result.canceled {
		return "", ErrEditCanceled
	}
	return result.ta.Value(), nil
}

func resolveEditor() []string {
	if e := os.Getenv("VISUAL"); e != "" {
		return strings.Fields(e)
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return strings.Fields(e)
	}
	// fallback to OS defaults
	if runtime.GOOS == "windows" {
		return []string{"notepad"}
	}
	for _, candidates := range []string{"nano", "vim", "vi"} {
		if _, err := exec.LookPath(candidates); err == nil {
			return []string{candidates}
		}
	}
	return []string{"vi"}
}

func EditTextExternal(initial, suffix string) (string, error) {
	f, err := os.CreateTemp("", "cb_edit_*"+suffix)
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	path := f.Name()
	defer os.Remove(path)

	if _, err := f.WriteString(initial); err != nil {
		f.Close()
		return "", fmt.Errorf("writing to temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("closing temp file: %w", err)
	}

	editorCmd := resolveEditor()
	args := append(append([]string{}, editorCmd[1:]...), path)
	cmd := exec.Command(editorCmd[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("running editor %q: %w", editorCmd[0], err)
	}

	edited, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading edited file: %w", err)
	}
	return string(edited), nil
}
