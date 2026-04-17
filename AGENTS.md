# AGENTS.md

## Repo Summary

This repository is a small Go service named `echo`.

Despite the stale auth-related comments in `main.go`, the active code behaves as
an HTTP echo/debug server built with Fiber. It accepts any HTTP method on any
path and returns a JSON payload describing the incoming request.

Current behavior includes:

- Echoing request path, method, params, query, headers, and hostname
- Parsing and echoing request bodies for non-`GET` requests
- Reading multipart form data when present
- Supporting an `X-Delay` request header that sleeps for `N` seconds before
  responding
- Starting a `gops` agent on port `8081`

Default app port comes from `config.yaml` and is currently `8080`.

## Important Files

- `main.go`: process entrypoint; starts `gops` and then the Fiber app
- `routes/routes.go`: Fiber app setup and the single catch-all route
- `routes/middleware.go`: middleware helpers; some are currently unused
- `config/config.go`: configuration loading via Viper
- `config.yaml`: default runtime configuration
- `config-canary.yaml`: alternate profile config loaded when
  `ACTIVE_PROFILE=canary`

## Architecture Notes

- The service currently has one active route: `All("/*")`
- There are no domain-specific auth endpoints in the active code
- `recover` and compression middleware are enabled
- A request logger is instantiated, but in the current code it is registered
  after the catch-all handler, so it does not observe handled requests
- A rate limiter helper exists in `routes/middleware.go`, but it is not wired
  into the app

## Runbook

Use `rtk` for shell commands in this repo.

Common commands:

```bash
rtk test go test ./...
rtk proxy go run .
rtk proxy go build ./...
```

Notes:

- `rtk test go test ./...` is the best quick verification pass
- Use `rtk proxy` for raw Go commands when you want the original command
  behavior without RTK filtering surprises
- The current repository has little to no automated route coverage, so behavior
  checks may require running the app and hitting it manually

## Configuration

Configuration is loaded at process startup by Viper from:

- `config.yaml`
- `config-<ACTIVE_PROFILE>.yaml` when `ACTIVE_PROFILE` is set

Relevant environment variables:

- `ACTIVE_PROFILE`: selects `config-<profile>.yaml`
- `LOG_LEVEL`: one of `PANIC`, `FATAL`, `ERROR`, `WARN`, `INFO`, `DEBUG`,
  `TRACE`

Be aware:

- Config loading happens in package `init()` code
- Middleware code also caches config in `init()`
- If configuration loading is refactored later, startup ordering matters

## Editing Guidance

When changing this service, preserve the current intended role unless the user
explicitly asks for a redesign.

Prefer these principles:

- Keep the app lightweight and easy to reason about
- Treat it as an echo/debug utility unless requirements clearly change
- Be careful with middleware order in Fiber; registration order materially
  changes behavior
- Avoid logging normal request shapes as errors
- Be cautious with request reflection because headers and bodies may contain
  sensitive data
- Be cautious with `X-Delay`; it can be abused to hold worker time open

## Known Code Caveats

These are worth knowing before making changes:

- `routes/routes.go` checks `Listen` errors incorrectly; the condition is
  inverted
- `routes/routes.go` attempts `MultipartForm()` on every request, which creates
  noisy error logs for normal non-multipart traffic
- `routes/middleware.go` contains unused helpers
- The limiter config expects `RateLimit`, but the YAML currently uses `Max`

## Expectations For Future Agents

- Read `routes/routes.go`, `routes/middleware.go`, and `config/config.go`
  before changing request handling
- Use `rtk`-prefixed commands when working in the shell
- Prefer `rtk test go test ./...` after code changes
- Do not assume the auth comments in `main.go` describe the live behavior
- If you introduce real endpoints, update this file to reflect the new
  architecture
