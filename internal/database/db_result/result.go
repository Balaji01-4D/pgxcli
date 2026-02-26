package dbresult

import (
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type Result interface {
	CommandTag() pgconn.CommandTag
}

type ResultDuration struct {
	ExecutionTime  time.Duration // time until first row is ready
	TTFR           time.Duration // time to first row (client-observed)
	StreamDuration time.Duration // time to drain remaining rows
}
