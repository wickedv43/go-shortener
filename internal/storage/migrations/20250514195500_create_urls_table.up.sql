CREATE TABLE IF NOT EXISTS urls (
    uuid SERIAL NOT NULL,
    short_url TEXT NOT NULL,
    original_url TEXT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);