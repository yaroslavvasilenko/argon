-- +goose Up
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

-- +goose Down
DROP TABLE IF EXISTS search_cursors;
DROP TABLE IF EXISTS search_info;
