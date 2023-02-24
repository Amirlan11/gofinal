CREATE TABLE IF NOT EXISTS sales (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    description text NOT NULL,
    duration integer NOT NULL,
    foodsale text[] NOT NULL,
    version integer NOT NULL DEFAULT 1
);
