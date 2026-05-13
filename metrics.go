package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Metrics struct {
	CollectedAt  time.Time
	Container    ContainerMetrics
	System       SystemMetrics
	Database     DatabaseMetrics
	Requests     []RequestEntry
	Error        string
}

type ContainerMetrics struct {
	Name   string
	Status string // "running" | "stopped" | "unknown"
	CPU    string
	Memory string
	Uptime string
}

type SystemMetrics struct {
	DiskTotal string
	DiskUsed  string
	DiskFree  string
	DiskPct   string
}

type DatabaseMetrics struct {
	Size         string
	ActorCount   string
	LogCount     string
	TodayCalls   string
	TotalBalance string
}

type RequestEntry struct {
	Time     string
	Method   string
	Path     string
	Status   string
	Duration string
}

var startedRe  = regexp.MustCompile(`Started (\w+) "([^"]+)"`)
var completedRe = regexp.MustCompile(`Completed (\d+)[^i]+in (\d+ms)`)

func collectMetrics(ex Executor, containerPrefix string) *Metrics {
	m := &Metrics{CollectedAt: time.Now()}

	// Find container
	name, err := ex.Run(fmt.Sprintf(`docker ps --format '{{.Names}}' --filter name=%s | head -1`, containerPrefix))
	if err != nil || name == "" {
		m.Container = ContainerMetrics{Status: "stopped"}
		m.Error = "container not found"
		return m
	}
	m.Container.Name = name
	m.Container.Status = "running"

	// CPU + memory
	stats, err := ex.Run(fmt.Sprintf(`docker stats --no-stream --format '{{.CPUPerc}} | {{.MemUsage}}' %s`, name))
	if err == nil {
		parts := strings.SplitN(stats, "|", 2)
		if len(parts) == 2 {
			m.Container.CPU = strings.TrimSpace(parts[0])
			m.Container.Memory = strings.TrimSpace(parts[1])
		}
	}

	// Uptime
	uptime, err := ex.Run(fmt.Sprintf(`docker inspect --format '{{.State.StartedAt}}' %s`, name))
	if err == nil {
		if t, err := time.Parse(time.RFC3339Nano, uptime); err == nil {
			m.Container.Uptime = formatDuration(time.Since(t))
		}
	}

	// Disk
	disk, err := ex.Run(`df -h / | tail -1 | awk '{print $2, $3, $4, $5}'`)
	if err == nil {
		parts := strings.Fields(disk)
		if len(parts) == 4 {
			m.System.DiskTotal = parts[0]
			m.System.DiskUsed = parts[1]
			m.System.DiskFree = parts[2]
			m.System.DiskPct = parts[3]
		}
	}

	// DB size
	dbSize, err := ex.Run(fmt.Sprintf(`docker exec %s du -sh /rails/storage/production.sqlite3 2>/dev/null | cut -f1`, name))
	if err == nil {
		m.Database.Size = dbSize
	}

	// SQLite queries
	sqliteQuery := func(query string) string {
		out, err := ex.Run(fmt.Sprintf(`docker exec %s sqlite3 /rails/storage/production.sqlite3 "%s" 2>/dev/null`, name, query))
		if err != nil {
			return "—"
		}
		return out
	}

	m.Database.ActorCount   = sqliteQuery("SELECT COUNT(*) FROM actors;")
	m.Database.LogCount     = sqliteQuery("SELECT COUNT(*) FROM capability_logs;")
	m.Database.TodayCalls   = sqliteQuery("SELECT COUNT(*) FROM capability_logs WHERE date(invoked_at) = date('now');")
	balance                 := sqliteQuery("SELECT COALESCE(ROUND(SUM(balance_cents)/100.0, 2), 0) FROM actors;")
	m.Database.TotalBalance  = "$" + balance

	// Recent logs
	logs, err := ex.Run(fmt.Sprintf(`docker logs --tail 80 %s 2>&1`, name))
	if err == nil {
		m.Requests = parseRequests(logs)
	}

	return m
}

func parseRequests(logs string) []RequestEntry {
	lines := strings.Split(logs, "\n")
	type pending struct{ method, path string }
	pending_map := map[string]pending{}
	var entries []RequestEntry

	for _, line := range lines {
		if m := startedRe.FindStringSubmatch(line); m != nil {
			// extract request id from [...] prefix if present
			id := extractRequestID(line)
			pending_map[id] = pending{method: m[1], path: m[2]}
		}
		if m := completedRe.FindStringSubmatch(line); m != nil {
			id := extractRequestID(line)
			ts := extractTimestamp(line)
			p := pending_map[id]
			entries = append(entries, RequestEntry{
				Time:     ts,
				Method:   p.method,
				Path:     p.path,
				Status:   m[1],
				Duration: m[2],
			})
			delete(pending_map, id)
		}
	}

	// Return last 12, most recent first
	if len(entries) > 12 {
		entries = entries[len(entries)-12:]
	}
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries
}

var reqIDRe = regexp.MustCompile(`\[([a-f0-9\-]{36})\]`)
var tsRe    = regexp.MustCompile(`at \d{4}-\d{2}-\d{2} (\d{2}:\d{2}:\d{2})`)

func extractRequestID(line string) string {
	if m := reqIDRe.FindStringSubmatch(line); m != nil {
		return m[1]
	}
	return line
}

func extractTimestamp(line string) string {
	if m := tsRe.FindStringSubmatch(line); m != nil {
		return m[1]
	}
	return "—"
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
