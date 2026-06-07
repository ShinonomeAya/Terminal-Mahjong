package tui

import tea "github.com/charmbracelet/bubbletea"

func Run() error {
	program := tea.NewProgram(NewModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := program.Run()
	return err
}
