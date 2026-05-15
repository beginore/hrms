-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks
    ADD COLUMN report_description text,
    ADD COLUMN report_url         text,
    ADD COLUMN submitted_at       timestamptz;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks
    DROP COLUMN IF EXISTS submitted_at,
    DROP COLUMN IF EXISTS report_url,
    DROP COLUMN IF EXISTS report_description;
-- +goose StatementEnd
