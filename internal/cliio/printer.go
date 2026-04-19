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
	PagerModeAuto   = "auto"
	PagerModeAlways = "always"
	PagerModeNever  = "never"

	defaultTerminalHeight = 24
	autoPagerMinBytes     = 4096
)

type Printer interface {
	SetOut(out io.Writer)
	SetErrOut(errOut io.Writer)
	SetPagerMode(mode string) error
	Print(str string)
	PrintError(err error)
	PrintTime(time time.Duration)
	PrintViaPager(str string)
}

type PgxPrinter struct {
	out    io.Writer
	errOut io.Writer

	pagerMode      string
	isTerminal     bool
	terminalHeight int

	pagerPath      string
	pagerArgs      []string
	pagerSupported bool
}

func NewPgxPrinter(out io.Writer, errOut io.Writer) *PgxPrinter {
	p := &PgxPrinter{
		out:            out,
		errOut:         errOut,
		pagerMode:      PagerModeAuto,
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

func (p *PgxPrinter) SetOut(out io.Writer) {
	p.out = out
}

func (p *PgxPrinter) SetErrOut(errOut io.Writer) {
	p.errOut = errOut
}

func (p *PgxPrinter) SetPagerMode(mode string) error {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		normalized = PagerModeAuto
	}

	switch normalized {
	case PagerModeAuto, PagerModeAlways, PagerModeNever:
		p.pagerMode = normalized
		return nil
	default:
		return fmt.Errorf("invalid pager mode %q, expected one of: auto, always, never", mode)
	}
}

func (p *PgxPrinter) Print(str string) {
	printInfo(p.out, str)
}

func (p *PgxPrinter) PrintError(err error) {
	printErr(p.errOut, "%v\n", err)
}

func (p *PgxPrinter) PrintTime(time time.Duration) {
	printTime(p.out, "Time: %.3fs\n", time.Seconds())
}

func (p *PgxPrinter) PrintViaPager(str string) {
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

func EchoViaPager(writeFn func(io.Writer) error) error {
	p := NewPgxPrinter(os.Stdout, os.Stderr)
	return p.echoViaPager(writeFn)
}

func (p *PgxPrinter) shouldUsePager(str string) bool {
	switch p.pagerMode {
	case PagerModeNever:
		return false
	case PagerModeAlways:
		return p.isTerminal && p.pagerSupported
	default:
		if !p.isTerminal || !p.pagerSupported {
			return false
		}
		return len(str) >= autoPagerMinBytes || lineCount(str) > p.autoPagerLineThreshold()
	}
}

func (p *PgxPrinter) autoPagerLineThreshold() int {
	if p.terminalHeight <= 2 {
		return 1
	}
	return p.terminalHeight - 2
}

func lineCount(str string) int {
	if str == "" {
		return 0
	}
	return strings.Count(str, "\n") + 1
}

func ensureTrailingNewline(str string) string {
	if str == "" || strings.HasSuffix(str, "\n") {
		return str
	}
	return str + "\n"
}

func (p *PgxPrinter) echoViaPager(writeFn func(io.Writer) error) error {
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

func (p *PgxPrinter) tryPipePager(writeFn func(io.Writer) error) bool {
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

func (p *PgxPrinter) tryTempfilePager(writerFn func(io.Writer) error) bool {
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
