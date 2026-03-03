CREATE TABLE IF NOT EXISTS urls (
    id           BIGSERIAL    PRIMARY KEY,
    short_code   VARCHAR(10)  NOT NULL UNIQUE,
    original_url TEXT         NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls (original_url);
