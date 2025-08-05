package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time
type nextTaskMsg struct{}
type pauseMsg struct{}

// helper function to easily return messages as a tea.Cmd
func Cmd(msg any) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

// forces an update every 200ms, timer does not work without it
func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type task struct {
	name     string
	duration time.Duration
}

// move all timers to own model so that they can be easily reset on next task!
type model struct {
	startTime time.Time
	incTimer  time.Duration

	pausedAt    time.Time
	pausedTimer time.Duration
	paused      bool

	decTimer time.Duration

	tasks       []task
	currentTask task
	error       string
}

func initialModel() model {
	return model{
		paused: false,
		tasks: []task{
			{
				name:     "code",
				duration: time.Minute * 45,
			},
			{
				name:     "stare at wall",
				duration: time.Second * 13,
			},
			{
				name:     "read",
				duration: time.Minute * 25,
			},
		},
	}
}

func (m model) resetTimer() model {
	m.startTime = time.Now()
	m.incTimer = 0
	m.pausedAt = time.Now()
	m.pausedTimer = 0
	m.decTimer = m.currentTask.duration
	return m
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
			return m, Cmd(pauseMsg{})
		case "enter":
			return m, Cmd(nextTaskMsg{})
		}
	case tickMsg:
		// no task (inital state)
		if m.currentTask == (task{}) {
			return m, tick()
		}

		if m.paused {
			m.pausedTimer = time.Since(m.pausedAt)
		} else {
			m.incTimer = time.Since(m.startTime.Add(m.pausedTimer))
		}

		m.decTimer = m.currentTask.duration - m.incTimer
		if m.decTimer < 0 {
			return m, Cmd(nextTaskMsg{})
		}
		return m, tick()
	case pauseMsg:
		m.paused = !m.paused
		if m.paused {
			m.pausedAt = m.startTime.Add(m.incTimer)
		}
		return m, nil
	case nextTaskMsg:
		if m.currentTask != (task{}) {
			notify := exec.Cmd{Path: "/usr/bin/notify-send", Args: []string{"Task Complete", fmt.Sprintf("Completed task: \"%s\"", m.currentTask.name)}}
			err := notify.Run()
			if err != nil {
				m.error = err.Error()
				return m, tea.Quit
			}
		}

		if len(m.tasks) != 0 {
			m.currentTask, m.tasks = m.tasks[0], m.tasks[1:]
		} else {
			m.currentTask = task{}
		}

		m = m.resetTimer()
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	s := fmt.Sprintf("Elapsed Time: %s\n", m.decTimer.Truncate(time.Second))
	s += fmt.Sprintf("Paused: %t\n", m.paused)
	s += fmt.Sprintf("Current Task: %s for %s\n", m.currentTask.name, m.currentTask.duration.Truncate(time.Second))
	s += "Tasks:\n"
	for i, task := range m.tasks {
		s += fmt.Sprintf("    %d. %s - %s\n", i+1, task.name, task.duration.Truncate(time.Second))
	}
	s += "<q> quit <enter> next task\n"

	return s
}

func main() {
	m := initialModel()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	if m.error != "" {
		log.Fatal(m.error)
	}
}
