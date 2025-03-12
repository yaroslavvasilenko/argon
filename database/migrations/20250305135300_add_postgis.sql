-- +goose Up
-- +goose StatementBegin

-- Установка расширения PostGIS для поддержки географических запросов
CREATE EXTENSION IF NOT EXISTS postgis;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаление расширения PostGIS
DROP EXTENSION IF EXISTS postgis;

-- +goose StatementEnd
