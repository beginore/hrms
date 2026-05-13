-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS cities (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE
);

INSERT INTO cities (name) VALUES
    ('Almaty'),
    ('Astana'),
    ('Shymkent'),
    ('Karaganda'),
    ('Aktobe'),
    ('Taraz'),
    ('Pavlodar'),
    ('Oskemen'),
    ('Semey'),
    ('Atyrau'),
    ('Kostanay'),
    ('Kyzylorda'),
    ('Petropavl'),
    ('Oral'),
    ('Aktau'),
    ('Temirtau'),
    ('Turkestan'),
    ('Ekibastuz'),
    ('Rudny'),
    ('Zhezkazgan'),
    ('Balqash'),
    ('Taldykorgan'),
    ('Kentau'),
    ('Zhanaozen'),
    ('Stepnogorsk');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cities;
-- +goose StatementEnd
