-- +goose Up
-- +goose StatementBegin

-- Переименовываем поле text в original_description
ALTER TABLE listings RENAME COLUMN text TO original_description;

-- Создаем таблицы для полнотекстового поиска
CREATE TABLE listings_search_en (
    listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    title_vector tsvector,
    description_vector tsvector,
    PRIMARY KEY (listing_id)
);

CREATE TABLE listings_search_ru (
    listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    title_vector tsvector,
    description_vector tsvector,
    PRIMARY KEY (listing_id)
);

CREATE TABLE listings_search_es (
    listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    title_vector tsvector,
    description_vector tsvector,
    PRIMARY KEY (listing_id)
);

-- Создаем индексы для быстрого поиска
CREATE INDEX idx_listings_search_en_title ON listings_search_en USING GIN (title_vector);
CREATE INDEX idx_listings_search_en_desc ON listings_search_en USING GIN (description_vector);

CREATE INDEX idx_listings_search_ru_title ON listings_search_ru USING GIN (title_vector);
CREATE INDEX idx_listings_search_ru_desc ON listings_search_ru USING GIN (description_vector);

CREATE INDEX idx_listings_search_es_title ON listings_search_es USING GIN (title_vector);
CREATE INDEX idx_listings_search_es_desc ON listings_search_es USING GIN (description_vector);

-- Создаем триггерную функцию для автоматического обновления поисковых векторов
CREATE OR REPLACE FUNCTION update_search_vectors()
RETURNS TRIGGER AS $$
BEGIN
    -- Английский
    INSERT INTO listings_search_en (listing_id, title_vector, description_vector)
    VALUES (
        NEW.id,
        to_tsvector('english', COALESCE(NEW.title, '')),
        to_tsvector('english', COALESCE(NEW.original_description, ''))
    )
    ON CONFLICT (listing_id) DO UPDATE SET
        title_vector = to_tsvector('english', COALESCE(NEW.title, '')),
        description_vector = to_tsvector('english', COALESCE(NEW.original_description, ''));

    -- Русский
    INSERT INTO listings_search_ru (listing_id, title_vector, description_vector)
    VALUES (
        NEW.id,
        to_tsvector('russian', COALESCE(NEW.title, '')),
        to_tsvector('russian', COALESCE(NEW.original_description, ''))
    )
    ON CONFLICT (listing_id) DO UPDATE SET
        title_vector = to_tsvector('russian', COALESCE(NEW.title, '')),
        description_vector = to_tsvector('russian', COALESCE(NEW.original_description, ''));

    -- Испанский
    INSERT INTO listings_search_es (listing_id, title_vector, description_vector)
    VALUES (
        NEW.id,
        to_tsvector('spanish', COALESCE(NEW.title, '')),
        to_tsvector('spanish', COALESCE(NEW.original_description, ''))
    )
    ON CONFLICT (listing_id) DO UPDATE SET
        title_vector = to_tsvector('spanish', COALESCE(NEW.title, '')),
        description_vector = to_tsvector('spanish', COALESCE(NEW.original_description, ''));

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Создаем триггер
CREATE TRIGGER trigger_update_search_vectors
    AFTER INSERT OR UPDATE OF title, original_description
    ON listings
    FOR EACH ROW
    EXECUTE FUNCTION update_search_vectors();

-- Заполняем поисковые таблицы существующими данными
INSERT INTO listings_search_en (listing_id, title_vector, description_vector)
SELECT 
    id,
    to_tsvector('english', COALESCE(title, '')),
    to_tsvector('english', COALESCE(original_description, ''))
FROM listings
ON CONFLICT DO NOTHING;

INSERT INTO listings_search_ru (listing_id, title_vector, description_vector)
SELECT 
    id,
    to_tsvector('russian', COALESCE(title, '')),
    to_tsvector('russian', COALESCE(original_description, ''))
FROM listings
ON CONFLICT DO NOTHING;

INSERT INTO listings_search_es (listing_id, title_vector, description_vector)
SELECT 
    id,
    to_tsvector('spanish', COALESCE(title, '')),
    to_tsvector('spanish', COALESCE(original_description, ''))
FROM listings
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем триггер
DROP TRIGGER IF EXISTS trigger_update_search_vectors ON listings;
DROP FUNCTION IF EXISTS update_search_vectors();

-- Удаляем индексы
DROP INDEX IF EXISTS idx_listings_search_en_title;
DROP INDEX IF EXISTS idx_listings_search_en_desc;
DROP INDEX IF EXISTS idx_listings_search_ru_title;
DROP INDEX IF EXISTS idx_listings_search_ru_desc;
DROP INDEX IF EXISTS idx_listings_search_es_title;
DROP INDEX IF EXISTS idx_listings_search_es_desc;

-- Удаляем таблицы поиска
DROP TABLE IF EXISTS listings_search_en;
DROP TABLE IF EXISTS listings_search_ru;
DROP TABLE IF EXISTS listings_search_es;

-- Возвращаем старое название колонки
ALTER TABLE listings RENAME COLUMN original_description TO text;

-- +goose StatementEnd
