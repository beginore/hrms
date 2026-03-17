-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_sessions
(
    "id"            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id"       uuid        NOT NULL REFERENCES users (id),
    "cognito_username" varchar(255) NOT NULL,
    "refresh_token" text        NOT NULL UNIQUE,
    "created_at"    timestamptz NOT NULL DEFAULT now(),
    "expires_at"    timestamptz NOT NULL
    );

CREATE INDEX idx_user_sessions_refresh_token ON user_sessions (refresh_token);
CREATE INDEX idx_user_sessions_user_id ON user_sessions (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP INDEX IF EXISTS idx_user_sessions_refresh_token;
DROP TABLE IF EXISTS user_sessions;
-- +goose StatementEnd