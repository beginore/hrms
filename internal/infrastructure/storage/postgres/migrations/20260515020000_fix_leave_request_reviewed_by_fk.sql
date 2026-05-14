-- +goose Up
-- +goose StatementBegin

ALTER TABLE leave_requests DROP CONSTRAINT IF EXISTS lr_reviewed_by_fk;
ALTER TABLE leave_requests ADD CONSTRAINT lr_reviewed_by_fk FOREIGN KEY (reviewed_by) REFERENCES users(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE leave_requests DROP CONSTRAINT IF EXISTS lr_reviewed_by_fk;
ALTER TABLE leave_requests ADD CONSTRAINT lr_reviewed_by_fk FOREIGN KEY (reviewed_by) REFERENCES employees(id);

-- +goose StatementEnd
