package dbresult

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// exec → returns status/affected rows (INSERT, UPDATE, CREATE, etc.)

type ExecResult struct {
	commandTag pgconn.CommandTag
	Duration   ResultDuration
}

func NewExecResult(cmdTag pgconn.CommandTag, executionTime time.Duration) *ExecResult {
	return &ExecResult{
		commandTag: cmdTag,
		Duration: ResultDuration{
			ExecutionTime: executionTime,
		},
	}
}

func (e *ExecResult) CommandTag() pgconn.CommandTag {
	return e.commandTag
}

func (e *ExecResult) RowsAffected() int64 {
	return e.commandTag.RowsAffected()
}

func (e *ExecResult) String() string {
	return e.commandTag.String()
}

func (e *ExecResult) DurationString() string {
	return fmt.Sprintf("Execution Time: %s", e.Duration.ExecutionTime.String())
}

func (e *ExecResult) Close() error {
	// No resources to clean up for ExecResult
	return nil
}
