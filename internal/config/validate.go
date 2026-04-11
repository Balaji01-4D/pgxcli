package config

import "errors"

func validate(cfg Config) error {
	var errs []error

	if cfg.Main.Prompt == "" {
		errs = append(errs, errors.New("prompt must not be empty"))
	}
	if cfg.Main.Style == "" {
		errs = append(errs, errors.New("style must not be empty"))
	}
	if cfg.Main.HistoryFile == "" {
		errs = append(errs, errors.New("history file path must not be empty"))
	}
	if cfg.Main.LogFile == "" {
		errs = append(errs, errors.New("log file path must not be empty"))
	}

	return errors.Join(errs...)
}
