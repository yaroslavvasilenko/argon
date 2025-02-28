-- +goose Up
CREATE TABLE listing_categories (
    listing_id UUID NOT NULL,
    category_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (listing_id, category_id)
);

-- +goose Down
DROP TABLE listing_categories;
