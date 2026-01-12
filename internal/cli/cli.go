package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/balaji01-4d/pgxcli/internal/config"
	postgres "github.com/balaji01-4d/pgxcli/internal/database"
	"github.com/balaji01-4d/pgxcli/internal/repl"
)

const version = "v0.1.0"

type pgxCLI struct {
	config *config.Config
	client *postgres.Client
	repl   *repl.Repl
}

func (p *pgxCLI) start(ctx context.Context, db, user string) error {
	var connector postgres.Connector
	var err error

	if strings.Contains(db, "://") || strings.Contains(db, "=") {
		connector, err = postgres.NewPGConnectorFromConnString(db)
		if err != nil {
			return fmt.Errorf("invalid connection string: %w", err)
		}
	} else {
		var password string

		if opts.NeverPrompt {
			password = os.Getenv("PGPASSWORD")
		}

		if opts.ForcePrompt && password == "" {
			pwd, err := p.repl.ReadPassword()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			password = pwd
		}

		connector, err = postgres.NewPGConnectorFromFields(
			opts.Host,
			db,
			user,
			password,
			opts.Port,
		)
		if err != nil {
			return fmt.Errorf("failed to create connector: %w", err)
		}
	}

	ConnErr := p.client.Connect(ctx, connector)
	if ConnErr != nil {
		if shouldAskForPassword(ConnErr, opts.NeverPrompt) {
			pwd, err := p.repl.ReadPassword()
			if err != nil {
				return fmt.Errorf("failed to read password: %v", err)
			}
			connector.UpdatePassword(pwd)
			ConnErr = p.client.Connect(ctx, connector)
			if ConnErr != nil {
				return err
			}
		} else {
			return ConnErr
		}
	}

	if !p.client.IsConnected() {
		return fmt.Errorf("not connected to any database")
	}

	p.repl.Run(ctx)

	return nil
}

func (p *pgxCLI) close(ctx context.Context) {
	p.client.Close(ctx)
	p.repl.Close()
}
