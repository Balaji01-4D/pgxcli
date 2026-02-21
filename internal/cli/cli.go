package cli

import (
	"context"
	"fmt"
	"log/slog"
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
	logger *slog.Logger
}

func (p *pgxCLI) start(ctx context.Context, db, user string) error {
	var connector postgres.Connector
	var err error

	if strings.Contains(db, "://") || strings.Contains(db, "=") {
		p.logger.Debug("using connection string mode")
		connector, err = postgres.NewPGConnectorFromConnString(db)
		if err != nil {
			p.logger.Error("invalid connection string", "error", err)
			return fmt.Errorf("invalid connection string: %w", err)
		}
	} else {
		p.logger.Debug("using field-based connection",
			"host", opts.Host,
			"port", opts.Port,
			"database", db,
			"user", user,
		)
		var password string

		if opts.NeverPrompt {
			password = os.Getenv("PGPASSWORD")
		}

		if opts.ForcePrompt && password == "" {
			pwd, err := p.repl.ReadPassword(user)
			if err != nil {
				p.logger.Error("failed to read password", "error", err)
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
			p.logger.Error("failed to create connector", "error", err)
			return fmt.Errorf("failed to create connector: %w", err)
		}
	}

	p.logger.Debug("attempting database connection")
	ConnErr := p.client.Connect(ctx, connector)
	if ConnErr != nil {
		if shouldAskForPassword(ConnErr, opts.NeverPrompt) {
			p.logger.Debug("connection failed, prompting for password")
			pwd, err := p.repl.ReadPassword(user)
			if err != nil {
				p.logger.Error("failed to read password on retry", "error", err)
				return fmt.Errorf("failed to read password: %v", err)
			}
			connector.UpdatePassword(pwd)
			ConnRetryErr := p.client.Connect(ctx, connector)
			if ConnRetryErr != nil {
				p.logger.Error("connection retry failed", "error", ConnRetryErr)
				return ConnRetryErr
			}
		} else {
			p.logger.Error("database connection failed", "error", ConnErr)
			return ConnErr
		}
	}

	if !p.client.IsConnected() {
		p.logger.Error("not connected to any database after connection attempt")
		return fmt.Errorf("not connected to any database")
	}

	p.logger.Info("database connection established",
		"database", db,
		"user", user,
	)

	p.repl.Run(ctx)

	return nil
}

func (p *pgxCLI) close(ctx context.Context) error {
	p.logger.Debug("closing application")
	if err := p.client.Close(ctx); err != nil {
		p.logger.Error("failed to close database connection", "error", err)
		return err
	}
	p.repl.Close()
	p.logger.Info("application closed")
	return nil
}
