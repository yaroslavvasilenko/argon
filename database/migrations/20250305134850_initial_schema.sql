-- +goose Up
-- +goose StatementBegin

-- Создание основной таблицы объявлений
CREATE TABLE IF NOT EXISTS listings (
    id UUID NOT NULL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    original_description TEXT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    views_count INTEGER NOT NULL,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

-- Создание таблицы для обмена валют
CREATE TABLE currency_exchange_rates (
    symbol VARCHAR(10) PRIMARY KEY,
    exchange_rate DECIMAL(18,8) NOT NULL,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Создание таблиц для кэширования поиска
CREATE TABLE IF NOT EXISTS search_cursors (
    id VARCHAR(64) PRIMARY KEY,
    cursor_data JSONB NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS search_info (
    id VARCHAR(64) PRIMARY KEY,
    search_data JSONB NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_search_cursors_expires_at ON search_cursors(expires_at);
CREATE INDEX IF NOT EXISTS idx_search_info_expires_at ON search_info(expires_at);

-- Создание таблицы для категорий объявлений
CREATE TABLE listing_categories (
    listing_id UUID NOT NULL,
    category_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (listing_id, category_id)
);

-- Создание таблицы для локаций
CREATE TABLE locations (
    id VARCHAR(255) NOT NULL,
    listing_id UUID NOT NULL REFERENCES listings(id),
    name VARCHAR(255) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    radius INTEGER NOT NULL
);

CREATE INDEX idx_locations_coords ON locations(latitude, longitude);
CREATE INDEX idx_locations_listing ON locations(listing_id);

-- Создание таблицы для характеристик объявлений
CREATE TABLE IF NOT EXISTS listing_characteristics (
    listing_id UUID NOT NULL,
    characteristics JSONB NOT NULL,
    
    CONSTRAINT fk_listing_characteristics_listing_id FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE
);

CREATE INDEX idx_listing_characteristics_listing_id ON listing_characteristics(listing_id);

-- Создание таблицы для продвижения объявлений
CREATE TABLE IF NOT EXISTS listing_boosts (
    listing_id UUID NOT NULL,
    boost_type VARCHAR(255) NOT NULL,
    commission FLOAT NOT NULL,
    
    CONSTRAINT fk_listing_boosts_listing_id FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE
);

CREATE INDEX idx_listing_boosts_listing_id ON listing_boosts(listing_id);
CREATE INDEX idx_listing_boosts_boost_type ON listing_boosts(boost_type);

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

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем триггер
DROP TRIGGER IF EXISTS trigger_update_search_vectors ON listings;
DROP FUNCTION IF EXISTS update_search_vectors();

-- Удаляем индексы поиска
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

-- Удаляем индексы
DROP INDEX IF EXISTS idx_listing_boosts_listing_id;
DROP INDEX IF EXISTS idx_listing_boosts_boost_type;
DROP INDEX IF EXISTS idx_listing_characteristics_listing_id;
DROP INDEX IF EXISTS idx_locations_coords;
DROP INDEX IF EXISTS idx_locations_listing;
DROP INDEX IF EXISTS idx_search_cursors_expires_at;
DROP INDEX IF EXISTS idx_search_info_expires_at;

-- Удаляем таблицы в правильном порядке (сначала зависимые таблицы)
DROP TABLE IF EXISTS listing_boosts;
DROP TABLE IF EXISTS listing_characteristics;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS listing_categories;
DROP TABLE IF EXISTS search_cursors;
DROP TABLE IF EXISTS search_info;
DROP TABLE IF EXISTS currency_exchange_rates;
DROP TABLE IF EXISTS listings;

-- +goose StatementEnd
