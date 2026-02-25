# 📝 Todo App — Go REST API

A production-grade Todo application built with Go, following best practices: clean architecture, dependency injection, JWT auth with multi-device support, smart priority scoring, and productivity analytics.

---

**Design principles applied:**
- **Hexagonal / Clean Architecture** — domain layer has zero external dependencies
- **Dependency Inversion** — services depend on interfaces, not concrete repos
- **Repository Pattern** — swap databases without touching business logic
- **Explicit error types** — sentinel errors for domain errors, wrapped errors for infra
- **Context propagation** — every I/O function accepts `context.Context`
- **Graceful shutdown** — SIGTERM/SIGINT handled cleanly

---

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- Docker & Docker Compose

### 1. Clone & configure
```bash
git clone https://github.com/galihaleanda/todo-app
cd todo-app
cp .env.example .env
# Edit .env — especially JWT secrets in production!
```

### 2. Start infrastructure
```bash
make docker-up
```

### 3. Run the app
```bash
make run
```

### 4. Or build & run binary
```bash
make build
./bin/todo-app
```

---

## 📡 API Reference

Base URL: `http://localhost:8080/api/v1`

All protected routes require: `Authorization: Bearer <access_token>`

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/register` | Create account |
| POST | `/auth/login` | Login (returns JWT pair) |
| POST | `/auth/refresh` | Rotate tokens |
| POST | `/auth/logout` | Revoke tokens |

**Register**
```json
POST /auth/register
{
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "password": "secretpass"
}
```

**Login**
```json
POST /auth/login
{
  "email": "budi@example.com",
  "password": "secretpass",
  "device_id": "browser-chrome-mac"
}
```

### Projects

| Method | Path | Description |
|--------|------|-------------|
| POST | `/projects` | Create project |
| GET | `/projects` | List my projects |
| GET | `/projects/:id` | Get project |
| PATCH | `/projects/:id` | Update project |
| DELETE | `/projects/:id` | Delete project |

```json
POST /projects
{
  "name": "My Side Project",
  "description": "Building something cool",
  "type": "side_project",
  "color": "#6366F1"
}
```

Project types: `personal` · `work` · `side_project`

### Tasks

| Method | Path | Description |
|--------|------|-------------|
| POST | `/tasks` | Create task |
| GET | `/tasks` | List tasks (filterable) |
| GET | `/tasks/:id` | Get task |
| PATCH | `/tasks/:id` | Update task |
| DELETE | `/tasks/:id` | Delete task |

**Query filters for `GET /tasks`:**
```
?status=todo|in_progress|done
?priority=low|medium|high
?project_id=<uuid>
?overdue=true
?search=<text>
?page=1&limit=20
```

**Create task**
```json
POST /tasks
{
  "title": "Implement OAuth2",
  "description": "Add Google login to the API",
  "priority": "high",
  "estimated_hours": 4.5,
  "due_date": "2025-03-01T00:00:00Z",
  "project_id": "uuid-here"
}
```

**Response includes `smart_score`** — a computed urgency score based on:
- Manual priority weight (low=10, medium=20, high=30)
- Due date proximity (up to +50 points, escalates when overdue)
- Current status (in_progress +15)
- Quick-win boost for tasks ≤1 hour estimated

### Analytics

| Method | Path | Description |
|--------|------|-------------|
| GET | `/analytics/dashboard` | Full productivity dashboard |
| GET | `/analytics/daily?from=YYYY-MM-DD&to=YYYY-MM-DD` | Daily breakdown |

**Dashboard response:**
```json
{
  "total_tasks": 42,
  "completed_tasks": 30,
  "completion_rate_percent": 71.4,
  "overdue_tasks": 3,
  "completed_this_week": 8,
  "avg_completion_time_hours": 3.2,
  "most_productive_day": "Wednesday",
  "weekly_breakdown": [...],
  "high_priority_pending": 2,
  "medium_priority_pending": 5,
  "low_priority_pending": 4
}
```

---

## 🛠 Makefile Targets

```bash
make run           # Run in development
make build         # Compile binary to bin/
make test          # Run tests with coverage
make lint          # Run golangci-lint
make tidy          # go mod tidy + verify
make migrate-up    # Apply SQL migrations
make migrate-down  # Rollback last migration
make docker-up     # Start postgres + redis
make docker-down   # Stop containers
```

---

## 🔒 Security Notes

- Passwords hashed with bcrypt (cost=10)
- Separate JWT secrets for access and refresh tokens
- Refresh tokens stored in DB (rotated on every use)
- Multi-device support via `device_id`
- Soft delete — data preserved for audit
- Config validation prevents weak secrets in production
