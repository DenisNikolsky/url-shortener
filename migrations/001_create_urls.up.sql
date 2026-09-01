CREATE TABLE urls (
                      id BIGSERIAL PRIMARY KEY,
                      short_code VARCHAR(16) NOT NULL UNIQUE,
                      original_url TEXT NOT NULL,
                      created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                      clicks BIGINT NOT NULL DEFAULT 0
);