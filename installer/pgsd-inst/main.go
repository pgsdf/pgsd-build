package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    step int
}

func initialModel() model {
    return model{step: 0}
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "enter":
            m.step++
        }
    }
    return m, nil
}

func (m model) View() string {
    switch m.step {
    case 0:
        return "PGSD Installer (prototype)\n\nPress Enter to continue, q to quit.\n"
    case 1:
        return "Step 1: Image selection (not implemented)\n\nPress Enter to continue, q to quit.\n"
    case 2:
        return "Step 2: Disk selection (not implemented)\n\nPress Enter to continue, q to quit.\n"
    case 3:
        return "Step 3: Confirmation (not implemented)\n\nPress Enter to continue, q to quit.\n"
    default:
        return "Installation pipeline not yet implemented in this prototype.\nPress q to quit.\n"
    }
}

func main() {
    if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
        fmt.Fprintln(os.Stderr, "error running TUI:", err)
        os.Exit(1)
    }
}
