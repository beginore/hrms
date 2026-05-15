-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks DROP CONSTRAINT tasks_created_by_fk;
ALTER TABLE tasks ADD CONSTRAINT tasks_created_by_fk FOREIGN KEY (created_by) REFERENCES users(id);
ALTER TABLE tasks DROP CONSTRAINT tasks_reviewed_by_fk;
ALTER TABLE tasks ADD CONSTRAINT tasks_reviewed_by_fk FOREIGN KEY (reviewed_by) REFERENCES users(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks DROP CONSTRAINT tasks_reviewed_by_fk;
ALTER TABLE tasks ADD CONSTRAINT tasks_reviewed_by_fk FOREIGN KEY (reviewed_by) REFERENCES employees(id);
ALTER TABLE tasks DROP CONSTRAINT tasks_created_by_fk;
ALTER TABLE tasks ADD CONSTRAINT tasks_created_by_fk FOREIGN KEY (created_by) REFERENCES employees(id);
-- +goose StatementEnd
