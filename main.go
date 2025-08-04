package main

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*25, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type model struct {
	startTime   time.Time
	timer       time.Duration
	pausedAt    time.Time
	pausedTimer time.Duration
	paused      bool
}

func initialModel() model {
	return model{
		startTime: time.Now(),
		paused:    false,
	}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case " ":
			m.paused = !m.paused
			if m.paused {
				m.pausedAt = m.startTime.Add(m.timer)
			}
			return m, nil
		}
	case tickMsg:
		if m.paused {
			m.pausedTimer = time.Since(m.pausedAt)
		} else {
			m.timer = time.Since(m.startTime.Add(m.pausedTimer))
		}

		return m, tick()
	}
	return m, nil
}

func (m model) View() string {
	s := fmt.Sprintf("Elapsed Time: %s\n", m.timer)
	s += fmt.Sprintf("Paused: %t\n", m.paused)
	s += "\nPress q to quit.\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
