-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS employees
(
    id uuid PRIMARY KEY,
    org_id uuid NOT NULL,
    user_id uuid NOT NULL,
    department_id uuid NOT NULL,
    position_id uuid NOT NULL,
    role VARCHAR(100) NOT NULL,
    salary_rate NUMERIC(10,2) NOT NULL,
    status VARCHAR(50) NOT NULL
    );

ALTER TABLE employees
    ADD CONSTRAINT employees_org_id_fk
        FOREIGN KEY (org_id) REFERENCES organizations(id);

ALTER TABLE employees
    ADD CONSTRAINT employees_user_id_fk
        FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE employees
    ADD CONSTRAINT employees_department_id_fk
        FOREIGN KEY (department_id) REFERENCES departments(id);

ALTER TABLE employees
    ADD CONSTRAINT employees_position_id_fk
        FOREIGN KEY (position_id) REFERENCES positions(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_position_id_fk;
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_department_id_fk;
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_user_id_fk;
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_org_id_fk;

DROP TABLE IF EXISTS employees;

-- +goose StatementEnd