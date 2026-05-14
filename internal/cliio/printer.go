// Package cliio provides utilities for printing output to the terminal
// including support for pagers and colored output.
package cliio

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/fatih/color"
	"github.com/google/shlex"
)

var (
	printErr  = color.New(color.FgHiRed).FprintfFunc()
	printInfo = color.New(color.FgWhite).FprintfFunc()
	printTime = color.New(color.FgHiCyan).FprintfFunc()
)

const (
	pagerModeAuto   = "auto"
	pagerModeAlways = "always"
	pagerModeNever  = "never"

	defaultTerminalHeight = 24
	autoPagerMinBytes     = 4096
)

// Printer defines terminal output behavior for regular, error, timed, and pager output.
type Printer interface {
	SetOut(out io.Writer)
	SetErrOut(errOut io.Writer)
	SetPagerMode(mode string) error
	Print(str string)
	PrintError(err error)
	PrintTime(time time.Duration)
	PrintViaPager(str string)
	ShouldUsePager(str string) bool
}

// pgxPrinter is the default Printer implementation used by the CLI.
type pgxPrinter struct {
	out    io.Writer
	errOut io.Writer

	pagerMode      string
	isTerminal     bool
	terminalHeight int

	pagerPath      string
	pagerArgs      []string
	pagerSupported bool
}

// NewPgxPrinter creates a printer with pager auto-detection.
func NewPgxPrinter(out io.Writer, errOut io.Writer) Printer {
	p := &pgxPrinter{
		out:            out,
		errOut:         errOut,
		pagerMode:      pagerModeAuto,
		isTerminal:     term.IsTerminal(os.Stdin.Fd()) && term.IsTerminal(os.Stdout.Fd()),
		terminalHeight: detectTerminalHeight(os.Stdout.Fd()),
		pagerPath:      "",
		pagerArgs:      nil,
		pagerSupported: false,
	}

	if pagerPath, pagerArgs, ok := resolvePagerCommand(); ok {
		p.pagerPath = pagerPath
		p.pagerArgs = pagerArgs
		p.pagerSupported = true
	}

	return p
}

// SetOut updates the destination for regular output.
func (p *pgxPrinter) SetOut(out io.Writer) {
	p.out = out
}

// SetErrOut updates the destination for error output.
func (p *pgxPrinter) SetErrOut(errOut io.Writer) {
	p.errOut = errOut
}

// SetPagerMode configures pager behavior ("auto", "always", or "never").
func (p *pgxPrinter) SetPagerMode(mode string) error {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		normalized = pagerModeAuto
	}

	switch normalized {
	case pagerModeAuto, pagerModeAlways, pagerModeNever:
		p.pagerMode = normalized
		return nil
	default:
		return fmt.Errorf("invalid pager mode %q, expected one of: auto, always, never", mode)
	}
}

// Print writes regular output to the configured output stream.
func (p *pgxPrinter) Print(str string) {
	printInfo(p.out, str)
}

// PrintError writes an error message to the configured error stream.
func (p *pgxPrinter) PrintError(err error) {
	printErr(p.errOut, "%v\n", err)
}

// PrintTime prints execution duration in seconds.
func (p *pgxPrinter) PrintTime(time time.Duration) {
	printTime(p.out, "Time: %.3fs\n", time.Seconds())
}

// PrintViaPager writes output either directly or through the configured pager.
func (p *pgxPrinter) PrintViaPager(str string) {
	output := ensureTrailingNewline(str)

	if !p.shouldUsePager(str) {
		if _, err := io.WriteString(p.out, output); err != nil {
			p.PrintError(err)
		}
		return
	}

	err := p.echoViaPager(func(w io.Writer) error {
		_, err := io.WriteString(w, output)
		return err
	})
	if err != nil {
		p.PrintError(err)
	}
}

func (p *pgxPrinter) shouldUsePager(str string) bool {
	switch p.pagerMode {
	case pagerModeNever:
		return false
	case pagerModeAlways:
		return p.isTerminal && p.pagerSupported
	default:
		if !p.isTerminal || !p.pagerSupported {
			return false
		}
		return len(str) >= autoPagerMinBytes || lineCount(str) > p.autoPagerLineThreshold()
	}
}

func (p *pgxPrinter) autoPagerLineThreshold() int {
	if p.terminalHeight <= 2 {
		return 1
	}
	return p.terminalHeight - 2
}

func lineCount(str string) int {
	if str == "" {
		return 0
	}
	count := strings.Count(str, "\n")
	if strings.HasSuffix(str, "\n") {
		return count
	}
	return count + 1
}

func ensureTrailingNewline(str string) string {
	if str == "" || strings.HasSuffix(str, "\n") {
		return str
	}
	return str + "\n"
}

func (p *pgxPrinter) echoViaPager(writeFn func(io.Writer) error) error {
	if !p.isTerminal || !p.pagerSupported {
		return writeFn(p.out)
	}

	if p.tryPipePager(writeFn) {
		return nil
	}

	if p.tryTempfilePager(writeFn) {
		return nil
	}

	return writeFn(p.out)
}

func (p *pgxPrinter) tryPipePager(writeFn func(io.Writer) error) bool {
	cmd := exec.Command(p.pagerPath, p.pagerArgs...)
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
	_ = stdin.Close()

	waiterr := waitIgnoringInterrupt(cmd)

	return writeErr == nil && waiterr == nil
}

func (p *pgxPrinter) tryTempfilePager(writerFn func(io.Writer) error) bool {
	tmp, err := os.CreateTemp("", "pager-*")
	if err != nil {
		return false
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()

	if err := writerFn(tmp); err != nil {
		_ = tmp.Close()
		return false
	}
	if err := tmp.Close(); err != nil {
		return false
	}

	cmd := exec.Command(p.pagerPath, append(p.pagerArgs, tmp.Name())...)
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

func detectTerminalHeight(fd uintptr) int {
	_, height, err := term.GetSize(fd)
	if err != nil || height <= 0 {
		return defaultTerminalHeight
	}
	return height
}

func resolvePagerCommand() (string, []string, bool) {
	pagerCmd := getPager()
	if len(pagerCmd) == 0 {
		return "", nil, false
	}

	cmdPath, err := exec.LookPath(pagerCmd[0])
	if err != nil {
		return "", nil, false
	}

	return cmdPath, pagerCmd[1:], true
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
		_ = os.Setenv("LESS", "-SRFX")
	}
	return []string{"less"}
}
