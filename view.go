package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	black  = lipgloss.Color("#000000")
	white  = lipgloss.Color("#FFFFFF")
	yellow = lipgloss.Color("#FACC15")
	green  = lipgloss.Color("#22C55E")
	red    = lipgloss.Color("#EF4444")
	gray   = lipgloss.Color("#6B7280")
	dkgray = lipgloss.Color("#374151")

	titleStyle = lipgloss.NewStyle().
			Background(black).
			Foreground(yellow).
			Bold(true).
			Padding(0, 2)

	sectionHeaderStyle = lipgloss.NewStyle().
				Foreground(yellow).
				Bold(true).
				MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(gray)

	valueStyle = lipgloss.NewStyle().
			Foreground(white).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(dkgray).
			Padding(0, 1)

	statusRunning = lipgloss.NewStyle().Foreground(green).Bold(true).Render("● RUNNING")
	statusStopped = lipgloss.NewStyle().Foreground(red).Bold(true).Render("● STOPPED")

	helpStyle = lipgloss.NewStyle().Foreground(gray)
)

func (m model) View() string {
	if m.loading && m.metrics == nil {
		return "\n  Loading…\n"
	}

	met := m.metrics
	w := m.width
	if w < 40 {
		w = 100
	}

	var b strings.Builder

	// Title bar
	status := statusStopped
	uptime := ""
	if met.Container.Status == "running" {
		status = statusRunning
		uptime = "  Uptime: " + valueStyle.Render(met.Container.Uptime)
	}
	title := titleStyle.Render("PROXY DASHBOARD") +
		"  " + status + uptime +
		"  " + labelStyle.Render(met.CollectedAt.Format("15:04:05"))
	b.WriteString(title + "\n")
	b.WriteString(strings.Repeat("─", w) + "\n")

	// Three columns: System | Database | Actors
	sysBox := renderSection("SYSTEM", []row{
		{"CPU", met.Container.CPU},
		{"Memory", met.Container.Memory},
		{"Disk Used", met.System.DiskUsed + " / " + met.System.DiskTotal},
		{"Disk Free", met.System.DiskFree + " (" + met.System.DiskPct + ")"},
	})

	dbBox := renderSection("DATABASE", []row{
		{"File Size", met.Database.Size},
		{"Actors", met.Database.ActorCount},
		{"Total Logs", met.Database.LogCount},
		{"Today Calls", met.Database.TodayCalls},
	})

	actorBox := renderSection("BILLING", []row{
		{"Total Balance", met.Database.TotalBalance},
		{"Container", shortName(met.Container.Name)},
	})

	revenueBox := renderSection("REVENUE", []row{
		{"Charged", met.Database.Revenue},
		{"Provider Cost", met.Database.ProviderCost},
		{"Profit", lipgloss.NewStyle().Foreground(green).Bold(true).Render(met.Database.Profit)},
	})

	colWidth := (w - 8) / 4
	cols := lipgloss.JoinHorizontal(lipgloss.Top,
		boxStyle.Width(colWidth).Render(sysBox),
		boxStyle.Width(colWidth).Render(dbBox),
		boxStyle.Width(colWidth).Render(actorBox),
		boxStyle.Width(colWidth).Render(revenueBox),
	)
	b.WriteString(cols + "\n")

	// Recent requests
	b.WriteString(strings.Repeat("─", w) + "\n")
	b.WriteString(sectionHeaderStyle.Render("  RECENT REQUESTS") + "\n")

	if met.Error != "" {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(red).Render(met.Error) + "\n")
	} else if len(met.Requests) == 0 {
		b.WriteString("  " + labelStyle.Render("No requests yet") + "\n")
	} else {
		for _, r := range met.Requests {
			b.WriteString(renderRequest(r) + "\n")
		}
	}

	b.WriteString(strings.Repeat("─", w) + "\n")
	b.WriteString(helpStyle.Render("  r refresh  q quit  refreshes every 5s") + "\n")

	return b.String()
}

type row struct{ label, value string }

func renderSection(title string, rows []row) string {
	var b strings.Builder
	b.WriteString(sectionHeaderStyle.Render(title) + "\n")
	for _, r := range rows {
		if r.value == "" || r.value == " / " || r.value == " ()" {
			continue
		}
		b.WriteString(labelStyle.Render(fmt.Sprintf("  %-12s", r.label)) +
			valueStyle.Render(r.value) + "\n")
	}
	return b.String()
}

func renderRequest(r RequestEntry) string {
	statusColor := green
	if r.Status != "" && r.Status[0] >= '4' {
		statusColor = red
	}
	statusStr := lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(r.Status)

	method := fmt.Sprintf("%-4s", r.Method)
	path := r.Path
	if len(path) > 40 {
		path = path[:37] + "…"
	}

	return fmt.Sprintf("  %s  %s  %-4s  %-40s  %s",
		labelStyle.Render(r.Time),
		statusStr,
		labelStyle.Render(method),
		valueStyle.Render(path),
		labelStyle.Render(r.Duration),
	)
}

func shortName(name string) string {
	// Trim long commit hash suffix
	parts := strings.Split(name, "-")
	if len(parts) > 4 {
		hash := parts[len(parts)-1]
		if len(hash) > 8 {
			return strings.Join(parts[:len(parts)-1], "-") + "-" + hash[:8] + "…"
		}
	}
	return name
}

// suppress unused import
var _ = time.Now
