-- +goose Up
-- +goose StatementBegin

-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    original_description TEXT,
    location_id VARCHAR(255),
    rating FLOAT,
    votes INTEGER,
    available BOOLEAN,
    editable BOOLEAN,
    zitadel_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Создание таблицы для контактов пользователей
CREATE TABLE IF NOT EXISTS user_contacts (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    text VARCHAR(255) NOT NULL,
    link VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_user_contacts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Создание таблицы для изображений пользователей
CREATE TABLE IF NOT EXISTS user_images (
    user_id VARCHAR(255) NOT NULL,
    image_url TEXT NOT NULL,
    
    CONSTRAINT fk_user_images_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Добавление поля location_id в таблицу listings
ALTER TABLE listings ADD COLUMN IF NOT EXISTS location_id VARCHAR(255);

-- Обновление таблицы locations для связи с пользователями
ALTER TABLE locations DROP CONSTRAINT IF EXISTS locations_listing_id_fkey;
ALTER TABLE locations ADD CONSTRAINT locations_listing_id_fkey 
    FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE;

-- Создание индексов для быстрого поиска
CREATE INDEX idx_user_contacts_user_id ON user_contacts(user_id);
CREATE INDEX idx_user_images_user_id ON user_images(user_id);
CREATE INDEX idx_listings_location_id ON listings(location_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаление индексов
DROP INDEX IF EXISTS idx_user_contacts_user_id;
DROP INDEX IF EXISTS idx_user_images_user_id;
DROP INDEX IF EXISTS idx_listings_location_id;

-- Удаление поля location_id из таблицы listings
ALTER TABLE listings DROP COLUMN IF EXISTS location_id;

-- Восстановление ограничения для таблицы locations
ALTER TABLE locations DROP CONSTRAINT IF EXISTS locations_listing_id_fkey;
ALTER TABLE locations ADD CONSTRAINT locations_listing_id_fkey 
    FOREIGN KEY (listing_id) REFERENCES listings(id);

-- Удаление таблиц в правильном порядке (сначала зависимые таблицы)
DROP TABLE IF EXISTS user_images;
DROP TABLE IF EXISTS user_contacts;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
