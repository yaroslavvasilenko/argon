-- +goose Up
-- +goose StatementBegin

-- Устанавливаем расширение pg_trgm для нечеткого поиска
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Создаем индекс для быстрого нечеткого поиска по заголовкам
CREATE INDEX IF NOT EXISTS idx_listings_title_trgm ON listings USING GIN (title gin_trgm_ops);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем индекс
DROP INDEX IF EXISTS idx_listings_title_trgm;

-- Удаляем расширение (если оно не используется другими компонентами)
-- DROP EXTENSION IF EXISTS pg_trgm;

-- +goose StatementEnd
