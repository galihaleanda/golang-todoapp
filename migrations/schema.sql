-- migrations/001_create_users.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(100)  NOT NULL,
    email         VARCHAR(255)  NOT NULL UNIQUE,
    password_hash VARCHAR(255)  NOT NULL,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users (email) WHERE deleted_at IS NULL;


-- migrations/002_create_refresh_tokens.sql
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT         NOT NULL UNIQUE,
    device_id  VARCHAR(255) NOT NULL,
    user_agent TEXT,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id  ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_expires  ON refresh_tokens (expires_at);


-- migrations/003_create_projects.sql
CREATE TYPE project_type AS ENUM ('personal', 'work', 'side_project');

CREATE TABLE IF NOT EXISTS projects (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    type        project_type NOT NULL,
    color       VARCHAR(7)   NOT NULL DEFAULT '#6366F1',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_projects_user_id ON projects (user_id) WHERE deleted_at IS NULL;


-- migrations/004_create_tasks.sql
CREATE TYPE task_status   AS ENUM ('todo', 'in_progress', 'done');
CREATE TYPE task_priority AS ENUM ('low', 'medium', 'high');

CREATE TABLE IF NOT EXISTS tasks (
    id               UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id          UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id       UUID          REFERENCES projects(id) ON DELETE SET NULL,
    title            VARCHAR(255)  NOT NULL,
    description      TEXT          NOT NULL DEFAULT '',
    status           task_status   NOT NULL DEFAULT 'todo',
    priority         task_priority NOT NULL DEFAULT 'medium',
    estimated_hours  NUMERIC(6,2),
    due_date         TIMESTAMPTZ,
    completed_at     TIMESTAMPTZ,
    smart_score      NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX idx_tasks_user_id    ON tasks (user_id)    WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_project_id ON tasks (project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_status     ON tasks (status)     WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_due_date   ON tasks (due_date)   WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_smart_score ON tasks (smart_score DESC) WHERE deleted_at IS NULL;

-- Partial index for overdue query
CREATE INDEX idx_tasks_overdue ON tasks (user_id, due_date)
    WHERE deleted_at IS NULL AND status != 'done';
