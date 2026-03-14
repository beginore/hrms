-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "departments"
(
    "id" uuid PRIMARY KEY,
    "org_id" uuid NOT NULL,
    "name" VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS "positions"
(
    "id" uuid PRIMARY KEY,
    "org_id" uuid NOT NULL,
    "name" VARCHAR(255) NOT NULL UNIQUE
);

ALTER TABLE departments ADD CONSTRAINT departments_org_id_fk FOREIGN KEY (org_id) REFERENCES organizations(id);
ALTER TABLE positions ADD CONSTRAINT positions_org_id_fk FOREIGN KEY (org_id) REFERENCES organizations(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE positions DROP CONSTRAINT IF EXISTS positions_org_id_fk;
ALTER TABLE departments DROP CONSTRAINT IF EXISTS departments_org_id_fk;

DROP TABLE IF EXISTS "positions";
DROP TABLE IF EXISTS departments;
-- +goose StatementEnd
