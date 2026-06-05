-- +goose Up
-- +goose StatementBegin
ALTER TABLE invites
    ADD COLUMN IF NOT EXISTS department_id uuid,
    ADD COLUMN IF NOT EXISTS position_id uuid,
    ADD COLUMN IF NOT EXISTS salary_rate numeric(10, 2),
    ADD COLUMN IF NOT EXISTS employee_status varchar(50);

ALTER TABLE invites
    ADD CONSTRAINT invites_department_id_fk
        FOREIGN KEY (department_id) REFERENCES departments(id);

ALTER TABLE invites
    ADD CONSTRAINT invites_position_id_fk
        FOREIGN KEY (position_id) REFERENCES positions(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE invites DROP CONSTRAINT IF EXISTS invites_position_id_fk;
ALTER TABLE invites DROP CONSTRAINT IF EXISTS invites_department_id_fk;

ALTER TABLE invites
    DROP COLUMN IF EXISTS employee_status,
    DROP COLUMN IF EXISTS salary_rate,
    DROP COLUMN IF EXISTS position_id,
    DROP COLUMN IF EXISTS department_id;
-- +goose StatementEnd
