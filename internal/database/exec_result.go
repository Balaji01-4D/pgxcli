package database

import (
	"fmt"
	"time"
)

// exec → returns status/affected rows (INSERT, UPDATE, CREATE, etc.)

type ExecResult struct {
	RowsAffected int64
	Status       string
	Duration     time.Duration
}

func (e *ExecResult) isResult() {}

func (e *ExecResult) Render() string {
	return fmt.Sprintf(
		"%s\nTime: %s",
		e.Status,
		e.Duration.String(),
	)
}
