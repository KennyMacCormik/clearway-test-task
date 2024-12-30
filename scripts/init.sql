CREATE TABLE IF NOT EXISTS "users" (
    "login" text NOT NULL UNIQUE,
    "pwd" text NOT NULL,
    "created_at" bigint NOT NULL default EXTRACT(EPOCH FROM now()), -- unix timestamp in sec
    "updated_at" bigint, -- unix timestamp in sec
    "deleted_at" bigint, -- unix timestamp in sec
    PRIMARY KEY ("login")
);

CREATE TABLE IF NOT EXISTS "files" (
    "asset_id" text NOT NULL,
    "user_login" text NOT NULL,
    "content_type" text,
    "data" bytea,
    "created_at" bigint NOT NULL default EXTRACT(EPOCH FROM now()), -- unix timestamp in sec
    "updated_at" bigint, -- unix timestamp in sec
    "deleted_at" bigint, -- unix timestamp in sec
    PRIMARY KEY ("asset_id", "user_login"),
    CONSTRAINT fk_user_login FOREIGN KEY ("user_login") REFERENCES "users"("login")
);

-- password: secret
insert into "users" values ('alice', '$2a$04$zkIAKg6l2DAuOMDDkRI9wuK43PjfONy41pgFqI6m8P2lueM13Rg1i') on conflict do nothing ;