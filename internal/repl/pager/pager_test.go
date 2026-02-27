package pager

import (
	"errors"
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
