-- +goose Up
create table users(
    id bigserial primary key,
    login varchar(100) not null,
    email varchar(254) not null,
    password_hash text not null,
    first_name varchar(50) not null,
	last_name varchar(50) not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz,

    CONSTRAINT users_login_unique UNIQUE (login),
    CONSTRAINT users_email_unique UNIQUE (email)
);

-- +goose Down
drop table if exists users;
