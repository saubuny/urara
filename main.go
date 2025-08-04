package main

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// move all timers to own model so that they can be easily reset on next task!
type model struct {
	startTime time.Time
	incTimer  time.Duration

	pausedAt    time.Time
	pausedTimer time.Duration
	paused      bool

	timerDuration time.Duration
	decTimer      time.Duration
}

func initialModel() model {
	return model{
		startTime:     time.Now(),
		timerDuration: time.Second * 3,
		decTimer:      time.Second * 3,
		paused:        false,
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
				m.pausedAt = m.startTime.Add(m.incTimer)
			}
			return m, nil
		}
	case tickMsg:
		if m.paused {
			m.pausedTimer = time.Since(m.pausedAt)
		} else {
			m.incTimer = time.Since(m.startTime.Add(m.pausedTimer))
		}

		m.decTimer = m.timerDuration - m.incTimer

		if m.decTimer < 0 {
			return m, tea.Quit
		}
		return m, tick()
	}
	return m, nil
}

func (m model) View() string {
	s := fmt.Sprintf("Elapsed Time: %s\n", m.decTimer.Truncate(time.Second))
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
