package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	host      := flag.String("host", "", "SSH target: user@host (omit for local mode)")
	key       := flag.String("key", "", "SSH private key path (required for remote mode)")
	container := flag.String("container", "budget-clear-proxy", "Docker container name prefix")
	flag.Parse()

	var ex Executor
	if *host != "" {
		if *key == "" {
			fmt.Fprintln(os.Stderr, "ERROR: --key is required when using --host")
			os.Exit(1)
		}
		ssh, err := newSSHExecutor(*host, *key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SSH connection failed: %v\n", err)
			os.Exit(1)
		}
		ex = ssh
	} else {
		ex = &LocalExecutor{}
	}
	defer ex.Close()

	p := tea.NewProgram(
		newModel(ex, *container),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
