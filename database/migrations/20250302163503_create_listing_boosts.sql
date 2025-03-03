-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS listing_boosts (
    listing_id UUID NOT NULL,
    boost_type VARCHAR(255) NOT NULL,
    commission FLOAT NOT NULL,
    
    CONSTRAINT fk_listing_boosts_listing_id FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE
);

CREATE INDEX idx_listing_boosts_listing_id ON listing_boosts(listing_id);
CREATE INDEX idx_listing_boosts_boost_type ON listing_boosts(boost_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS listing_boosts;
-- +goose StatementEnd
