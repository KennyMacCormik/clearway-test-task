CREATE TABLE IF NOT EXISTS "users" (
    "login" text NOT NULL UNIQUE,
    "pwd" text NOT NULL,
    "created_at" bigint NOT NULL default EXTRACT(EPOCH FROM now()),
    "updated_at" bigint NOT NULL DEFAULT 0, -- unix timestamp in sec
    "deleted_at" bigint NOT NULL DEFAULT 0, -- unix timestamp in sec
    PRIMARY KEY ("login")
);

CREATE TABLE IF NOT EXISTS "sessions" (
    "id" serial PRIMARY KEY,
    "user_login" text NOT NULL,
    "token" text NOT NULL UNIQUE,
    "iat" bigint NOT NULL,
    "exp" bigint NOT NULL,
    "created_at" bigint NOT NULL default EXTRACT(EPOCH FROM now()),
    "updated_at" bigint NOT NULL DEFAULT 0, -- unix timestamp in sec
    "deleted_at" bigint NOT NULL DEFAULT 0, -- unix timestamp in sec
    CONSTRAINT fk_user_login FOREIGN KEY ("user_login") REFERENCES "users"("login")
);

CREATE TABLE IF NOT EXISTS "files" (
    "id" serial PRIMARY KEY,
    "asset_name" text NOT NULL,
    "user_login" text NOT NULL,
    "content_type" text,
    "data" bytea,
    "created_at" bigint NOT NULL default EXTRACT(EPOCH FROM now()),
    "updated_at" bigint NOT NULL DEFAULT 0, -- unix timestamp in sec
    "deleted_at" bigint NOT NULL DEFAULT 0, -- unix timestamp in sec
    CONSTRAINT fk_user_login FOREIGN KEY ("user_login") REFERENCES "users"("login"),
    CONSTRAINT unique_asset_user_active UNIQUE (asset_name, user_login, deleted_at)
);

CREATE INDEX idx_asset_user ON files (asset_name, user_login);

-- password: secret
insert into "users" values ('alice', '$2a$04$zkIAKg6l2DAuOMDDkRI9wuK43PjfONy41pgFqI6m8P2lueM13Rg1i') on conflict do nothing ;