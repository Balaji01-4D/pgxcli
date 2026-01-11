package cli

import (
	"context"

	"github.com/balaji01-4d/pgxcli/internal/config"
	postgres "github.com/balaji01-4d/pgxcli/internal/database"
	"github.com/balaji01-4d/pgxcli/internal/repl"
)

const version = "v0.1.0"

type pgxcli struct {
	config *config.Config
	client *postgres.Client
	repl   *repl.Repl
}

func (p pgxcli) close(ctx context.Context) {
	p.client.Close(ctx)
	p.repl.Close()
}
