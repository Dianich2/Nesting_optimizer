-- +goose Up
create table users(
    id bigserial primary key,
    login varchar(100) not null unique,
    email varchar(254) not null unique,
    password_hash text not null,
    first_name varchar(50) not null,
	last_name varchar(50) not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

-- +goose Down
drop table if exists users;
