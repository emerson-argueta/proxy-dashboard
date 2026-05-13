package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Executor runs shell commands either locally or over SSH.
type Executor interface {
	Run(cmd string) (string, error)
	Close()
}

// LocalExecutor runs commands on the local machine.
type LocalExecutor struct{}

func (l *LocalExecutor) Run(cmd string) (string, error) {
	out, err := exec.Command("sh", "-c", cmd).Output()
	return strings.TrimSpace(string(out)), err
}

func (l *LocalExecutor) Close() {}

// SSHExecutor runs commands on a remote machine over SSH.
type SSHExecutor struct {
	client *ssh.Client
}

func newSSHExecutor(host, keyPath string) (*SSHExecutor, error) {
	if !strings.Contains(host, ":") {
		host = host + ":22"
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read key %s: %w", keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}

	user := "root"
	if idx := strings.Index(host, "@"); idx != -1 {
		user = host[:idx]
		host = host[idx+1:]
	}

	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5e9,
	}

	client, err := ssh.Dial("tcp", host, cfg)
	if err != nil {
		// try resolving via net.ResolveTCPAddr
		addr, rerr := net.ResolveTCPAddr("tcp", host)
		if rerr != nil {
			return nil, fmt.Errorf("connect to %s: %w", host, err)
		}
		client, err = ssh.Dial("tcp", addr.String(), cfg)
		if err != nil {
			return nil, fmt.Errorf("connect to %s: %w", host, err)
		}
	}

	return &SSHExecutor{client: client}, nil
}

func (s *SSHExecutor) Run(cmd string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	out, err := session.Output(cmd)
	return strings.TrimSpace(string(out)), err
}

func (s *SSHExecutor) Close() {
	s.client.Close()
}
