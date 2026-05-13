package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const refreshInterval = 5 * time.Second

type model struct {
	ex        Executor
	container string
	metrics   *Metrics
	loading   bool
	width     int
	height    int
}

type metricsMsg *Metrics
type tickMsg time.Time

func newModel(ex Executor, container string) model {
	return model{
		ex:        ex,
		container: container,
		loading:   true,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchMetrics(),
		scheduleTick(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, m.fetchMetrics()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case metricsMsg:
		m.metrics = (*Metrics)(msg)
		m.loading = false
		return m, nil

	case tickMsg:
		return m, tea.Batch(m.fetchMetrics(), scheduleTick())
	}

	return m, nil
}

func (m model) fetchMetrics() tea.Cmd {
	return func() tea.Msg {
		return metricsMsg(collectMetrics(m.ex, m.container))
	}
}

func scheduleTick() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
