-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS listing_characteristics (
    listing_id UUID NOT NULL,
    characteristics JSONB NOT NULL,
    
    CONSTRAINT fk_listing_characteristics_listing_id FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE
);

CREATE INDEX idx_listing_characteristics_listing_id ON listing_characteristics(listing_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS listing_characteristics;
-- +goose StatementEnd
