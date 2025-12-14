# Teleport Lite

Teleport Lite is a lightweight access-controller that lets you manage SSH targets, users, and role-based permissions from a single Go application. It ships with a browser UI (Gin templates + Tailwind), a WebSocket-based SSH console, audit logging, and a tiny Go agent binary that can be deployed on remote machines.

> **Status**: Active development. Expect rough edges and frequent changes.

## Highlights

- **User, role, and permission management** built on Gin + GORM with a MySQL backend.
- **Resource inventory** with token-based agent registration and SSH connect buttons.
- **Embedded terminal** that opens multiple SSH sessions in tabs via xterm.js and Gorilla WebSocket.
- **Audit trail** for every privileged action, including pagination + search.
- **Seed data** for a default organization, roles, permissions, and admin user.
- **Standalone agent** (`cmd/agent`) that can be packaged and run on hosts you want to control.

## Project Layout

```
cmd/
 ├─ api/        # Main HTTP API/server (Gin)
 └─ agent/      # Lightweight resource agent
internal/
 ├─ auth, http, handlers, router  # REST + WebSocket entrypoints
 ├─ models, db, seed              # Persistence layer (GORM)
 ├─ ui                            # Templates, JS, Tailwind classes
 └─ agent                         # Helpers used by the agent binary
migrations/                       # SQL reference docs / ERD
dist/                             # Prebuilt agent archives (if any)
```

## Requirements

- Go 1.25+
- MySQL 8 (or compatible 5.7+)
- Make/Curl (optional, for scripts)
- A modern browser for the UI (Chrome/Firefox/Edge)

## Configuration

`internal/config` loads values from `.env` (via `godotenv`) and falls back to OS environment variables.

```dotenv
MYSQL_DSN="teleport:teleport@tcp(127.0.0.1:3306)/teleport_lite?parseTime=true&loc=Local"
JWT_SECRET="replace-me"
APP_PORT=8080
# Optional: used when registering agents via API
AGENT_REG_TOKEN="dev-token"
```

- `MYSQL_DSN` **required** – standard Go MySQL DSN (`user:pass@tcp(host:port)/db?parseTime=true`).
- `JWT_SECRET` **required** – secret for signing session tokens.
- `APP_PORT` – HTTP port (defaults to `8080` if empty).
- `AGENT_REG_TOKEN` – optional server-side guard for agent registration.

## Getting Started

```bash
git clone https://github.com/<you>/teleport_lite
cd teleport_lite
cp .env.example .env   # if you have one, otherwise create using the variables above

# Install dependencies (handled automatically by Go modules)
go mod tidy

# Start the API / UI server
go run ./cmd/api
```

During startup the server:

1. Connects to MySQL and runs `AutoMigrate`.
2. Seeds a default org, roles, permissions, and an admin user.
3. Launches a background local agent (`agent.RunLocalAgent`) for demo purposes.
4. Exposes the UI on `http://localhost:8080`.

### Default Credentials

```
Email:    admin@example.com
Password: admin123
```

(Change the password immediately after logging in.)

### Running the Agent Manually

To build and run the agent outside the API server:

```bash
go build -o dist/teleport-agent ./cmd/agent
CONTROLLER_URL=http://127.0.0.1:8080 \
AGENT_REG_TOKEN=dev-token \
./dist/teleport-agent
```

Agents poll `/agents/heartbeat`, register via `/agents/register`, and appear under the Resources page once approved.

## UI/UX Notes

- **Users** – create accounts, assign roles, set connect usernames, and reset passwords.
- **Resources** – view registered machines, generate install tokens, and open SSH sessions. The Connect dialog now supports multiple terminal tabs per host.
- **Audit Trail** – search by user/action/resource/IP, view 20 rows at a time, and fetch the next page via cursor-based pagination. Search & cursor state are persisted in cookies so you can refresh and resume where you left off.

## Development Tips

- The frontend lives entirely in `internal/ui` (Go templates + vanilla JS). No build tooling is required.
- When modifying Go files, run `go fmt ./...` and `go test ./...` (tests TBD).
- SQL schemas in `migrations/` are reference documents; the application relies on GORM `AutoMigrate`.
- If you regenerate the agent binaries, place them in `dist/` to keep parity with the UI download hints.

## Roadmap / Ideas

- API docs & Postman collection
- SSO providers / OIDC support
- RBAC editor from the UI
- Agent auto-update pipeline

---

Happy hacking! Report bugs or feature ideas via issues/PRs. Contributions are welcome. 🚀
