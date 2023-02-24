CREATE TABLE IF NOT EXISTS contacts (
     id bigserial PRIMARY KEY,
     created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email text NOT NULL,
    subject text NOT NULL,
    message text NOT NULL,
    version integer NOT NULL DEFAULT 1
    );
