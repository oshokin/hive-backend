-- +goose Up
-- +goose StatementBegin
CREATE TYPE gender_type AS enum (
    'MALE',
    'FEMALE',
    'UNKNOWN'
);

CREATE TABLE cities (
    id smallserial PRIMARY KEY,
    name varchar(100) NOT NULL UNIQUE
);

COMMENT ON TABLE cities IS 'Список городов';

COMMENT ON COLUMN cities.id IS 'ID города, генерируется автоматически';

COMMENT ON COLUMN cities.name IS 'Название города';

CREATE TABLE users (
    id bigserial PRIMARY KEY, -- ID пользователя, генерируется автоматически
    email varchar(100) NOT NULL UNIQUE, -- E-mail
    password_hash varchar(100) NOT NULL, -- Хеш пароля
    city_id smallint NOT NULL REFERENCES cities (id) ON DELETE SET NULL ON UPDATE CASCADE, -- Город
    first_name varchar(30) NOT NULL, -- Имя
    last_name varchar(60) NOT NULL, -- Фамилия
    birthdate date NOT NULL, -- Дата рождения
    gender gender_type NOT NULL DEFAULT 'UNKNOWN', -- Пол
    interests varchar(500) NOT NULL DEFAULT '' -- Интересы
);

COMMENT ON TABLE users IS 'Список пользователей';

COMMENT ON COLUMN users.id IS 'ID пользователя, генерируется автоматически';

COMMENT ON COLUMN users.email IS 'E-mail';

COMMENT ON COLUMN users.password_hash IS 'Хеш пароля';

COMMENT ON COLUMN users.city_id IS 'Город';

COMMENT ON COLUMN users.first_name IS 'Имя';

COMMENT ON COLUMN users.last_name IS 'Фамилия';

COMMENT ON COLUMN users.birthdate IS 'Дата рождения';

COMMENT ON COLUMN users.gender IS 'Пол';

COMMENT ON COLUMN users.interests IS 'Интересы';

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE users;

DROP TABLE cities;

DROP TYPE gender_type;

-- +goose StatementEnd
