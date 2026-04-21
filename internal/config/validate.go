package config

import (
	"errors"
	"strings"
)

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

	pagerMode := strings.ToLower(strings.TrimSpace(cfg.Main.Pager))
	if pagerMode == "" {
		errs = append(errs, errors.New("pager mode must not be empty"))
	} else {
		switch pagerMode {
		case "auto", "always", "never":
		default:
			errs = append(errs, errors.New("pager mode must be one of: auto, always, never"))
		}
	}
	onError := cfg.Main.OnError
	if onError == "" {
		errs = append(errs, errors.New("on_error action must not be empty"))
	} else if !onError.IsValid() {
		errs = append(errs, errors.New("on_error action must be one of: STOP, RESUME"))
	}

	return errors.Join(errs...)
}
