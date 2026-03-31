# Repository Overview
This is the **EcnoteS** backend service (written in Go 1.26+) and its Angular front-end. It handles user requests to create and process tasks, dispatching background jobs via a queue and notifying users of progress in real time (via Server-Sent Events). Key components:
- **Go modules:** Root `go.mod` declares Go 1.26. Use `setup-go@v5` in CI for consistency【32†L347-L355】.
- **Directory layout:** `cmd/server/main.go` starts the HTTP API; `internal/` holds core logic (services, repositories, workers, config). `frontend/` contains the Angular app.
- **Languages & tools:** Go (1.26+), Gin (or net/http), Angular 16+, Redis/NSQ (for queues), Docker for containerization, GitHub Actions for CI/CD.

# Building and Running
- Use the `Makefile` or direct commands:
  - `go mod tidy` to install/update dependencies.
  - `go build ./cmd/server` to compile the backend.
  - `npm install && ng build` in `frontend/` for the UI.
- **Testing:** Run `go test ./...` for unit/integration tests. Use `go test -cover` and `golangci-lint run` to enforce quality.
- **Linting:** `golangci-lint` with modules enabled should pass. Ensure `go fmt` and `go vet` have been run.
- **Docker:** `docker-compose.yml` is provided to start Redis (and any DB). The `start` target in `Makefile` launches all services.

# CI/CD
- GitHub Actions workflows under `.github/workflows/` automate builds and tests on every push.
- Use `actions/setup-go@v5` with `go-version: '1.21.x'` for reproducible builds【32†L347-L355】.
- The pipeline includes `go test`, `go vet`, `golangci-lint`, and builds Docker images for staging.

# Key Info
- **Config files:** `config.yaml` holds environment configs; `.env.example` shows required vars.
- **Dependencies:** See `go.mod` for Go libs (e.g. Gin, Redis client, OTel) and `package.json` for Angular packages.
- **Running Tests:** Tests assume a running Redis (start via Docker `make start`). Integration tests use `ory/dockertest` to spin up Redis if not present.
- **Known Constraints:** Tasks must never block API threads. Always call context timeouts. E2E flow: API enqueues, returns immediately, and then notifies frontend via SSE.

This should allow Copilot to build, test, and modify the code without guesswork. Only modify instructions if the repository structure or tools change.
