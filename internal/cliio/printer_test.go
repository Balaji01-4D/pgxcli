package cliio

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeWaiter struct {
	errs []error
	i    int
}

func (fw *fakeWaiter) Wait() error {
	err := fw.errs[fw.i]
	fw.i++
	return err
}

func Test_waitIgnoringInterrupt_IgnoreENTR(t *testing.T) {
	fw := &fakeWaiter{
		errs: []error{
			syscall.EINTR,
			syscall.EINTR,
			nil,
		},
	}

	err := waitIgnoringInterrupt(fw)
	assert.Nil(t, err)
}

func Test_waitIgnoringInterrupt_ReturnOtherError(t *testing.T) {
	someErr := errors.New("some error")
	fw := &fakeWaiter{
		errs: []error{
			syscall.EINTR,
			someErr,
		},
	}

	err := waitIgnoringInterrupt(fw)
	assert.Equal(t, someErr, err)
}

func TestSetPagerMode(t *testing.T) {
	p := &PgxPrinter{}

	assert.NoError(t, p.SetPagerMode("AUTO"))
	assert.Equal(t, PagerModeAuto, p.pagerMode)

	assert.NoError(t, p.SetPagerMode("always"))
	assert.Equal(t, PagerModeAlways, p.pagerMode)

	assert.NoError(t, p.SetPagerMode("never"))
	assert.Equal(t, PagerModeNever, p.pagerMode)

	err := p.SetPagerMode("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid pager mode")
}

func TestShouldUsePager(t *testing.T) {
	basePrinter := &PgxPrinter{
		isTerminal:     true,
		pagerSupported: true,
		terminalHeight: 10,
	}

	basePrinter.pagerMode = PagerModeNever
	assert.False(t, basePrinter.shouldUsePager(strings.Repeat("a", 10000)))

	basePrinter.pagerMode = PagerModeAlways
	assert.True(t, basePrinter.shouldUsePager("small output"))

	basePrinter.pagerMode = PagerModeAuto
	assert.False(t, basePrinter.shouldUsePager("small output"))
	assert.True(t, basePrinter.shouldUsePager(strings.Repeat("a", autoPagerMinBytes)))
	assert.True(t, basePrinter.shouldUsePager(strings.Repeat("line\n", 10)))
}

func TestEnsureTrailingNewline(t *testing.T) {
	assert.Equal(t, "", ensureTrailingNewline(""))
	assert.Equal(t, "hello\n", ensureTrailingNewline("hello"))
	assert.Equal(t, "hello\n", ensureTrailingNewline("hello\n"))
}

func TestPrintViaPager_AppendsNewlineWhenNotPresent(t *testing.T) {
	out := &bytes.Buffer{}
	p := &PgxPrinter{
		out:       out,
		errOut:    io.Discard,
		pagerMode: PagerModeNever,
	}

	p.PrintViaPager("SELECT 9")
	assert.Equal(t, "SELECT 9\n", out.String())
}
