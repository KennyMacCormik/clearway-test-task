CREATE TABLE IF NOT EXISTS "user" (
    "login" text NOT NULL UNIQUE,
    "pwd" text NOT NULL,
    "createdAt" bigint NOT NULL default EXTRACT(EPOCH FROM now()),
    "updatedAt" timestamptz,
    "deletedAt" timestamptz,
    PRIMARY KEY ("login")
);

CREATE TABLE IF NOT EXISTS "files" (
    "asset_id" text NOT NULL,
    "user_login" text NOT NULL,
    "content_type" text,
    "data" bytea,
    "createdAt" bigint NOT NULL default EXTRACT(EPOCH FROM now()),
    "updatedAt" timestamptz,
    "deletedAt" timestamptz,
    PRIMARY KEY ("asset_id", "user_login"),
    CONSTRAINT fk_user_login FOREIGN KEY ("user_login") REFERENCES "user"("login")
);

-- password: secret
insert into "user" values ('alice', '$2a$04$zkIAKg6l2DAuOMDDkRI9wuK43PjfONy41pgFqI6m8P2lueM13Rg1i') on conflict do nothing ;