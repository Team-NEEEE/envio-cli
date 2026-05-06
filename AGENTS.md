# AGENTS.md

## Scope

These instructions apply to the entire `envio-cli` repository.

## Response Rules

- Respond in Korean by default.
- Base claims on repository files, command output, or primary documentation.
- Do not invent logs, test results, tickets, command output, file states, or business context.
- If a fact cannot be verified, state that it is not verified.
- Keep changes focused on the user's requested scope.

## Project Shape

- `cmd/envio/main.go` is the executable entrypoint.
- `internal/cli` owns Cobra command wiring, shared flags, help rendering, and input-error rendering.
- `internal/command` owns command result contracts such as `Result`, `AppError`, and `Reporter`.
- `internal/api` owns backend request/response contracts and HTTP client behavior.
- `internal/ui` owns plain, JSON, and TUI rendering.
- `internal/i18n` owns localized message loading and locale files.
- `internal/config` owns reusable environment parsing helpers.
- Placeholder `.gitkeep` files are intentional. Do not remove them unless the user explicitly asks.

## Design Boundaries

- Commands should return `command.Result` or `command.AppError`; they should not print UI directly.
- Rendering belongs in `internal/ui`, not in command/domain code.
- Cobra command registration and CLI input validation belong in `internal/cli`.
- Backend JSON response models and HTTP fallback behavior belong in `internal/api`.
- User-visible strings that need reuse or localization belong in `internal/i18n/locales`.
- Sensitive values must not be rendered in plain, JSON, or TUI output unless the user explicitly asks for that behavior.

## Adding Or Extending CLI Commands

When adding a new CLI command or subcommand:

1. Inspect existing `internal/cli` command patterns and tests first.
2. Define the command behavior as a small command/domain flow that returns `command.Result` or `command.AppError`.
3. Wire the Cobra command in `internal/cli`.
4. Add or update help text and gh-style usage rendering when the command is user-facing.
5. Keep hidden/internal commands hidden from normal help output.
6. Validate input errors through the input-error path so invalid command usage shows usage and allowed values.
7. Add localized labels/messages in both `active.en.json` and `active.ko.json` when user-visible text changes.
8. Add focused tests for command routing, validation, help output, and error rendering.

Do not add feature-specific command wording to foundation code unless the branch scope explicitly asks for that feature.

## Adding Backend API Calls

When adding a backend API call:

1. Define request and response types under `internal/api`.
2. Use the existing common response shape: `Response[T]`, `ErrorResponse`, and `FieldError`.
3. Keep backend error handling compatible with the common error response.
4. Preserve fallback diagnostics for non-standard HTTP responses such as empty bodies, HTML errors, proxy errors, and invalid JSON.
5. Keep API code independent from CLI rendering.
6. Add tests for success, backend error JSON, non-JSON failure, and empty-body behavior when applicable.

## Adding Or Changing Output Rendering

When changing output:

1. Preserve output mode selection rules unless the user asks to change them.
2. Keep plain output readable for CI and non-TTY environments.
3. Keep JSON output stable for debugging and automation.
4. Keep TUI behavior isolated to `internal/ui`.
5. Ensure sensitive summary values are hidden consistently across renderers.
6. Update renderer tests for plain, JSON, and TUI behavior when contracts change.

## i18n Rules

- Use the embedded locale file workflow in `internal/i18n`.
- Update both English and Korean locale files for user-visible strings.
- Keep default CLI language behavior consistent with existing tests.
- Avoid hardcoded reusable UI strings in command or renderer code.

## Go Coding Rules

- Prefer small, cohesive functions and explicit names.
- Use interfaces at package boundaries where they improve testability.
- Avoid duplicated literals when a named constant improves maintenance.
- Use structured JSON APIs instead of ad hoc string parsing.
- Keep package responsibilities narrow.
- Run `gofmt` on changed Go files.

## Verification

For code changes, run the relevant subset first, then the full set when practical:

```powershell
go test ./...
go vet ./...
go build ./cmd/envio
```

After dependency changes, run:

```powershell
go mod tidy
go test ./...
go vet ./...
go build ./cmd/envio
```

If a command cannot be run, report that clearly and explain why.

## Git And Commit Messages

- Follow `.github/git-commit-instructions.md` when generating commit messages.
- If the current branch contains a JIRA key, include it as `Ref:`.
- Do not include `.omc/`, generated binaries, temporary files, or `nul` in commits unless explicitly requested.
- Split commits by logical purpose, not by directory or file count alone.
