package repl

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"golang.org/x/term"

	"github.com/google/shlex"
)

func EchoViaPager(writeFn func(io.Writer) error) error {
	stdout := os.Stdout
	stdin := os.Stdin

	if !term.IsTerminal(int(stdin.Fd())) || !term.IsTerminal(int(stdout.Fd())) {
		return writeFn(stdout)
	}

	pagerCmd := getPager()

	if tryPipePager(pagerCmd, writeFn) {
		return nil
	}

	if tryTempfilePager(pagerCmd, writeFn) {
		return nil
	}

	return writeFn(stdout)
}

func terminalChecker(stdin *os.File, stdout *os.File) bool {
	return !term.IsTerminal(int(stdin.Fd())) || !term.IsTerminal(int(stdout.Fd()))
}

func tryPipePager(pagerCmd []string, writeFn func(io.Writer) error) bool {
	cmdPath, err := exec.LookPath(pagerCmd[0])
	if err != nil {
		return false
	}

	cmd := exec.Command(cmdPath, pagerCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false
	}
	if err := cmd.Start(); err != nil {
		return false
	}

	writeErr := writeFn(stdin)
	stdin.Close()

	waiterr := waitIgnoringInterrupt(cmd)

	return writeErr == nil && waiterr == nil
}

func tryTempfilePager(pagerCmd []string, writerFn func(io.Writer) error) bool {
	cmdPath, err := exec.LookPath(pagerCmd[0])
	if err != nil {
		return false
	}
	tmp, err := os.CreateTemp("", "pager-*")
	if err != nil {
		return false
	}
	defer os.Remove(tmp.Name())

	buf := &bytes.Buffer{}
	if err := writerFn(buf); err != nil {
		return false
	}

	if _, err := tmp.Write(buf.Bytes()); err != nil {
		return false
	}
	tmp.Close()

	cmd := exec.Command(cmdPath, tmp.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run() == nil
}

type waiter interface {
	Wait() error
}

func waitIgnoringInterrupt(w waiter) error {
	for {
		err := w.Wait()
		if err == nil {
			return nil
		}
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		return err
	}
}

func getPager() []string {
	if pager := os.Getenv("PAGER"); pager != "" {
		parts, err := shlex.Split(pager)
		if err == nil && len(parts) > 0 {
			return parts
		}
	}

	if runtime.GOOS == "windows" {
		return []string{"more"}
	}

	if _, okay := os.LookupEnv("LESS"); !okay {
		os.Setenv("LESS", "-SRFX")
	}
	return []string{"less"}
}
