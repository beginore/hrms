-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS payroll_cycles (
    id           uuid PRIMARY KEY,
    org_id       uuid        NOT NULL,
    period_start date        NOT NULL,
    period_end   date        NOT NULL,
    status       varchar(20) NOT NULL DEFAULT 'DRAFT',
    created_by   uuid        NOT NULL,
    approved_by  uuid,
    approved_at  timestamptz,
    paid_at      timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (org_id, period_start, period_end)
);
ALTER TABLE payroll_cycles ADD CONSTRAINT pc_org_fk FOREIGN KEY (org_id) REFERENCES organizations(id);
ALTER TABLE payroll_cycles ADD CONSTRAINT pc_created_by_fk FOREIGN KEY (created_by) REFERENCES users(id);
ALTER TABLE payroll_cycles ADD CONSTRAINT pc_approved_by_fk FOREIGN KEY (approved_by) REFERENCES users(id);

CREATE TABLE IF NOT EXISTS payroll_adjustments (
    id          uuid PRIMARY KEY,
    org_id      uuid         NOT NULL,
    employee_id uuid         NOT NULL,
    cycle_id    uuid,
    type        varchar(20)  NOT NULL,
    category    varchar(50)  NOT NULL,
    amount      numeric(12,2) NOT NULL,
    is_taxable  boolean      NOT NULL DEFAULT true,
    reason      text,
    created_by  uuid         NOT NULL,
    created_at  timestamptz  NOT NULL DEFAULT now()
);
ALTER TABLE payroll_adjustments ADD CONSTRAINT pa_org_fk FOREIGN KEY (org_id) REFERENCES organizations(id);
ALTER TABLE payroll_adjustments ADD CONSTRAINT pa_employee_fk FOREIGN KEY (employee_id) REFERENCES employees(id);
ALTER TABLE payroll_adjustments ADD CONSTRAINT pa_cycle_fk FOREIGN KEY (cycle_id) REFERENCES payroll_cycles(id);
ALTER TABLE payroll_adjustments ADD CONSTRAINT pa_created_by_fk FOREIGN KEY (created_by) REFERENCES users(id);

CREATE TABLE IF NOT EXISTS payroll_tax_rules (
    id             uuid PRIMARY KEY,
    org_id         uuid,
    country        varchar(2)   NOT NULL DEFAULT 'KZ',
    name           varchar(100) NOT NULL,
    rate           numeric(8,6) NOT NULL,
    applies_to     varchar(30)  NOT NULL,
    threshold_min  numeric(12,2),
    threshold_max  numeric(12,2),
    is_active      boolean      NOT NULL DEFAULT true,
    effective_from date         NOT NULL,
    effective_to   date,
    created_at     timestamptz  NOT NULL DEFAULT now(),
    updated_at     timestamptz  NOT NULL DEFAULT now()
);
ALTER TABLE payroll_tax_rules ADD CONSTRAINT ptr_org_fk FOREIGN KEY (org_id) REFERENCES organizations(id);

CREATE TABLE IF NOT EXISTS payroll_items (
    id                    uuid PRIMARY KEY,
    cycle_id              uuid         NOT NULL,
    org_id                uuid         NOT NULL,
    employee_id           uuid         NOT NULL,
    base_salary           numeric(12,2) NOT NULL,
    attendance_adjustment numeric(12,2) NOT NULL DEFAULT 0,
    overtime_amount       numeric(12,2) NOT NULL DEFAULT 0,
    bonuses_total         numeric(12,2) NOT NULL DEFAULT 0,
    deductions_total      numeric(12,2) NOT NULL DEFAULT 0,
    taxes_total           numeric(12,2) NOT NULL DEFAULT 0,
    gross_salary          numeric(12,2) NOT NULL,
    net_salary            numeric(12,2) NOT NULL,
    working_days          integer      NOT NULL,
    paid_days             numeric(5,2) NOT NULL,
    unpaid_days           numeric(5,2) NOT NULL,
    late_days             integer      NOT NULL,
    absent_days           integer      NOT NULL,
    overtime_minutes      integer      NOT NULL,
    status                varchar(20)  NOT NULL DEFAULT 'DRAFT',
    calculation_snapshot  jsonb        NOT NULL,
    created_at            timestamptz  NOT NULL DEFAULT now(),
    updated_at            timestamptz  NOT NULL DEFAULT now(),
    UNIQUE (cycle_id, employee_id)
);
ALTER TABLE payroll_items ADD CONSTRAINT pi_cycle_fk FOREIGN KEY (cycle_id) REFERENCES payroll_cycles(id);
ALTER TABLE payroll_items ADD CONSTRAINT pi_org_fk FOREIGN KEY (org_id) REFERENCES organizations(id);
ALTER TABLE payroll_items ADD CONSTRAINT pi_employee_fk FOREIGN KEY (employee_id) REFERENCES employees(id);

CREATE TABLE IF NOT EXISTS payroll_audit_logs (
    id              uuid PRIMARY KEY,
    org_id          uuid        NOT NULL,
    cycle_id        uuid,
    payroll_item_id uuid,
    actor_user_id   uuid        NOT NULL,
    action          varchar(50) NOT NULL,
    before_state    jsonb,
    after_state     jsonb,
    created_at      timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE payroll_audit_logs ADD CONSTRAINT pal_org_fk FOREIGN KEY (org_id) REFERENCES organizations(id);
ALTER TABLE payroll_audit_logs ADD CONSTRAINT pal_cycle_fk FOREIGN KEY (cycle_id) REFERENCES payroll_cycles(id);
ALTER TABLE payroll_audit_logs ADD CONSTRAINT pal_item_fk FOREIGN KEY (payroll_item_id) REFERENCES payroll_items(id);
ALTER TABLE payroll_audit_logs ADD CONSTRAINT pal_actor_fk FOREIGN KEY (actor_user_id) REFERENCES users(id);

CREATE INDEX IF NOT EXISTS payroll_cycles_org_status_idx ON payroll_cycles(org_id, status);
CREATE INDEX IF NOT EXISTS payroll_items_cycle_idx ON payroll_items(cycle_id);
CREATE INDEX IF NOT EXISTS payroll_adjustments_employee_cycle_idx ON payroll_adjustments(employee_id, cycle_id);
CREATE INDEX IF NOT EXISTS payroll_tax_rules_active_idx ON payroll_tax_rules(org_id, is_active, effective_from);

-- A simple default rule keeps local payroll usable before organisation-specific tax settings exist.
INSERT INTO payroll_tax_rules (
    id, org_id, country, name, rate, applies_to, is_active, effective_from
)
VALUES (
    gen_random_uuid(), NULL, 'KZ', 'Default Individual Income Tax', 0.100000, 'TAXABLE_INCOME', true, '2026-01-01'
)
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS payroll_tax_rules_active_idx;
DROP INDEX IF EXISTS payroll_adjustments_employee_cycle_idx;
DROP INDEX IF EXISTS payroll_items_cycle_idx;
DROP INDEX IF EXISTS payroll_cycles_org_status_idx;

ALTER TABLE payroll_audit_logs DROP CONSTRAINT IF EXISTS pal_actor_fk;
ALTER TABLE payroll_audit_logs DROP CONSTRAINT IF EXISTS pal_item_fk;
ALTER TABLE payroll_audit_logs DROP CONSTRAINT IF EXISTS pal_cycle_fk;
ALTER TABLE payroll_audit_logs DROP CONSTRAINT IF EXISTS pal_org_fk;
DROP TABLE IF EXISTS payroll_audit_logs;

ALTER TABLE payroll_items DROP CONSTRAINT IF EXISTS pi_employee_fk;
ALTER TABLE payroll_items DROP CONSTRAINT IF EXISTS pi_org_fk;
ALTER TABLE payroll_items DROP CONSTRAINT IF EXISTS pi_cycle_fk;
DROP TABLE IF EXISTS payroll_items;

ALTER TABLE payroll_tax_rules DROP CONSTRAINT IF EXISTS ptr_org_fk;
DROP TABLE IF EXISTS payroll_tax_rules;

ALTER TABLE payroll_adjustments DROP CONSTRAINT IF EXISTS pa_created_by_fk;
ALTER TABLE payroll_adjustments DROP CONSTRAINT IF EXISTS pa_cycle_fk;
ALTER TABLE payroll_adjustments DROP CONSTRAINT IF EXISTS pa_employee_fk;
ALTER TABLE payroll_adjustments DROP CONSTRAINT IF EXISTS pa_org_fk;
DROP TABLE IF EXISTS payroll_adjustments;

ALTER TABLE payroll_cycles DROP CONSTRAINT IF EXISTS pc_approved_by_fk;
ALTER TABLE payroll_cycles DROP CONSTRAINT IF EXISTS pc_created_by_fk;
ALTER TABLE payroll_cycles DROP CONSTRAINT IF EXISTS pc_org_fk;
DROP TABLE IF EXISTS payroll_cycles;

-- +goose StatementEnd
