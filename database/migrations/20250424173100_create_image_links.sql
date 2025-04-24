-- Создание таблицы для связи изображений с объявлениями
CREATE TABLE IF NOT EXISTS image_links (
    listing_id UUID REFERENCES listings(id),
    image_name TEXT NOT NULL,
    linked BOOLEAN DEFAULT TRUE,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
