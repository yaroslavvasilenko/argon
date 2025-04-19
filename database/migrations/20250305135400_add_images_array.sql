-- +goose Up
-- +goose StatementBegin

-- Добавление поля images в таблицу listings
ALTER TABLE listings
ADD COLUMN images TEXT[];

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаление поля images из таблицы listings
ALTER TABLE listings
DROP COLUMN IF EXISTS images;

-- +goose StatementEnd
