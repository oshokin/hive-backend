-- +goose Up
-- +goose StatementBegin
CREATE TYPE gender_type AS enum (
    'MALE',
    'FEMALE',
    'UNKNOWN'
);

CREATE TYPE job_status AS enum (
    'QUEUED',
    'PROCESSING',
    'CANCELLED',
    'COMPLETED',
    'FAILED'
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

CREATE TABLE randomizing_jobs (
    id serial PRIMARY KEY, -- ID задания, генерируется автоматически
    expected_count bigint NOT NULL, -- Количество добавляемых анкет
    current_count bigint NOT NULL DEFAULT 0, -- Количество уже добавленных анкет
    status job_status NOT NULL DEFAULT 'QUEUED', -- Статус задания
    started_at timestamp DEFAULT NULL, -- Дата / время начала задания
    finished_at timestamp DEFAULT NULL, -- Дата / время окончания задания
    error_message varchar(500) NOT NULL DEFAULT '' -- Текст ошибки, если задание завершилось с ошибкой
);

COMMENT ON TABLE randomizing_jobs IS 'Список заданий на заполнение анкет пользователей';

COMMENT ON COLUMN randomizing_jobs.id IS 'ID задания, генерируется автоматически';

COMMENT ON COLUMN randomizing_jobs.expected_count IS 'Количество добавляемых анкет';

COMMENT ON COLUMN randomizing_jobs.current_count IS 'Количество уже добавленных анкет';

COMMENT ON COLUMN randomizing_jobs.status IS 'Статус задания';

COMMENT ON COLUMN randomizing_jobs.started_at IS 'Дата / время начала задания';

COMMENT ON COLUMN randomizing_jobs.finished_at IS 'Дата / время окончания задания';

COMMENT ON COLUMN randomizing_jobs.error_message IS 'Текст ошибки, если задание завершилось с ошибкой';

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE randomizing_jobs;

DROP TABLE users;

DROP TABLE cities;

DROP TYPE job_status;

DROP TYPE gender_type;

-- +goose StatementEnd
