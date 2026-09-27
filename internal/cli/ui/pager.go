package ui

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

type pagerModel struct {
	vp       viewport.Model
	palette  Palette
	pageStep int
	lineStep int
}

func (m pagerModel) Init() tea.Cmd {
	return nil
}

func (m pagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			m.vp.ScrollDown(m.lineStep)
		case "k", "up":
			m.vp.ScrollUp(m.lineStep)
		case "ctrl+d", "pgdown":
			m.vp.ScrollDown(m.pageStep)
		case "ctrl+u", "pgup":
			m.vp.ScrollUp(m.pageStep)
		case "g":
			m.vp.GotoTop()
		case "G":
			m.vp.GotoBottom()
		}
	case tea.WindowSizeMsg:
		m.vp.SetWidth(msg.Width)
		m.vp.SetHeight(msg.Height - 1) // -1 for footer line
	}
	return m, nil
}

func (m pagerModel) View() tea.View {
	footer := m.palette.Muted.Render(" j/k ↑/↓ scroll · ctrl-d/ctrl-u page · g/G top/bottom · q quit ")
	v := tea.NewView(m.vp.View() + "\n" + footer)
	v.AltScreen = true
	return v
}

// RunPager initializes and runs a pager interface with the provided content, styling palette, and scrolling settings.
func RunPager(content string, palette Palette, pageStep, lineStep int) error {
	vp := viewport.New(viewport.WithWidth(0), viewport.WithHeight(0))
	vp.SetContent(content)

	m := pagerModel{vp: vp, palette: palette, pageStep: pageStep, lineStep: lineStep}
	_, err := tea.NewProgram(m).Run()
	return err
}
