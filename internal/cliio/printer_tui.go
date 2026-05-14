package cliio

import "os/exec"

func (p *pgxPrinter) ShouldUsePager(str string) bool {
	return p.shouldUsePager(str)
}

func ResolvePagerCommand() (path string, args []string, ok bool) {
	return resolvePagerCommand()
}

func PagerCmd(content string) (*exec.Cmd, bool) {
	pagerPath, pagerArgs, ok := resolvePagerCommand()
	if !ok {
		return nil, false
	}

	cmd := exec.Command(pagerPath, pagerArgs...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, false
	}

	go func() {
		defer stdin.Close()
		_, _ = stdin.Write([]byte(ensureTrailingNewline(content)))
	}()

	return cmd, true
}

func ShouldUsePagerFunc(str string, terminalHeight int) bool {
	if len(str) >= autoPagerMinBytes {
		return true
	}
	threshold := terminalHeight - 2
	if threshold < 1 {
		threshold = 1
	}
	return lineCount(str) > threshold
}
