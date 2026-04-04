# First release readiness TODO

This repository is **not fully ready** for a first public release yet.

## Missing items

- [ ] Replace placeholder `SECURITY.md` content with a real security policy (supported versions and vulnerability reporting process).
- [ ] Align CI Go versions with `go.mod` (`go 1.25.4`), since `.github/workflows/go.yml` still tests on `1.24`.
- [ ] Add a release note baseline (`CHANGELOG.md`) for the first tagged release.
- [ ] Define and document an explicit first-release checklist (version tag format, release artifacts, and verification steps).
- [ ] Install and document local lint prerequisites (`golangci-lint`) for contributors using `make precommit`.

## Audit note

I completed a full-file audit of the repository before generating this TODO.
