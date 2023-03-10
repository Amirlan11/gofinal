CREATE TABLE IF NOT EXISTS foods (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    price integer NOT NULL,
    waittime integer NOT NULL,
    recipe text[] NOT NULL,
    version integer NOT NULL DEFAULT 1
);