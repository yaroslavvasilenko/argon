-- +goose Up
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

-- +goose Down
DROP TABLE IF EXISTS locations;
