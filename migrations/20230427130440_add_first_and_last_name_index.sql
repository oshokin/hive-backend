-- +goose Up
-- +goose StatementBegin
CREATE INDEX users_first_name_idx ON users USING btree(first_name text_pattern_ops, last_name text_pattern_ops);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX users_first_name_idx;

-- +goose StatementEnd
