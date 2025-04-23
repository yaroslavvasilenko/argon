-- Создание таблицы для связи изображений с объявлениями
CREATE TABLE IF NOT EXISTS image_links (
    listing_id UUID REFERENCES listings(id),
    image_name TEXT NOT NULL,
    linked BOOLEAN DEFAULT TRUE,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_image_links_listing_id ON image_links(listing_id);
CREATE INDEX IF NOT EXISTS idx_image_links_image_name ON image_links(image_name);
CREATE INDEX IF NOT EXISTS idx_image_links_linked ON image_links(linked);
CREATE INDEX IF NOT EXISTS idx_image_links_updated_at ON image_links(updated_at);
