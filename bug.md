# Bug and Vulnerability Audit (source review)

Scope: static source-code review only (no fixes applied).

## Findings

1. **Connection leak when initial ping fails**  
   - **Severity:** High  
   - **File:** `/home/runner/work/pgxcli/pgxcli/internal/database/executor.go` (around lines 47-57)  
   - **Why:** `NewExecutor` creates a connection, then returns on `Ping` failure without closing that connection. Repeated failures can leak connections/resources.  
   - **Fix direction:** Close `conn` before returning on ping error.

2. **Connection leak when switching database**  
   - **Severity:** High  
   - **File:** `/home/runner/work/pgxcli/pgxcli/internal/database/client.go` (around lines 58-83)  
   - **Why:** `ChangeDatabase` replaces `c.Executor` with a new executor but does not close the old one. Multiple `\c` operations can accumulate open DB connections.  
   - **Fix direction:** Keep old executor reference, create and validate the new executor first, assign the new executor to `c.Executor`, then close the old executor.

3. **Ctrl+C is consumed but not used to cancel app context**  
   - **Severity:** Medium  
   - **File:** `/home/runner/work/pgxcli/pgxcli/internal/cli/runner.go` (around lines 112-119)  
   - **Why:** Signal goroutine loops forever reading `sigChan` and does nothing. `cancel()` from `WithCancel` is never called by signal handling, so graceful interrupt handling is broken.  
   - **Fix direction:** On first interrupt call `cancel()` (and usually `signal.Stop` / channel close handling).

4. **Config directory created with world-writable mode**  
   - **Severity:** Medium (security)  
   - **File:** `/home/runner/work/pgxcli/pgxcli/internal/config/config.go` (line ~72)  
   - **Why:** `os.MkdirAll(dir, os.ModePerm)` creates directories as `0777` (subject to umask). This is too permissive for user config storage and can allow local tampering.  
   - **Fix direction:** Use restrictive permissions like `0700` for user config directory.

5. **Log file is world-readable by default**  
   - **Severity:** Medium (security)  
   - **File:** `/home/runner/work/pgxcli/pgxcli/internal/logger/logger.go` (around lines 37-38)  
   - **Why:** Log file is opened with `0644`. Logs can contain sensitive operational details and should not be broadly readable on multi-user systems.  
   - **Fix direction:** Use `0600` for the log file; consider tighter log dir permissions if needed.

6. **Potential panic when saving history with fewer entries than initial load**  
   - **Severity:** Medium  
   - **File:** `/home/runner/work/pgxcli/pgxcli/internal/repl/history.go` (line ~85)  
   - **Why:** `newCommands := entries[h.loadCount:]` panics if `len(entries) < h.loadCount` (slice bounds out of range). This can happen if in-memory history is truncated/cleared during runtime.  
   - **Fix direction:** Guard with length check before slicing.

7. **History file write does not ensure parent directory exists**  
   - **Severity:** Low  
   - **File:** `/home/runner/work/pgxcli/pgxcli/internal/repl/history.go` (around lines 90-93)  
   - **Why:** `OpenFile` fails if parent directory is missing (especially for custom history paths), causing silent history persistence loss except log message.  
   - **Fix direction:** Ensure `filepath.Dir(h.path)` exists via `MkdirAll` before `OpenFile`.
