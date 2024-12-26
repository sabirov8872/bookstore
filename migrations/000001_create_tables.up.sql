CREATE TABLE roles (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

INSERT INTO roles (name)
VALUES ('user'),
       ('admin');

CREATE TABLE users (
    id         SERIAL PRIMARY KEY,
    role_id    INT NOT NULL,
    session_id TEXT,
    username   TEXT NOT NULL UNIQUE,
    password   TEXT NOT NULL,
    email      TEXT,
    phone      TEXT,
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

CREATE TABLE authors (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE genres (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE books (
    id          SERIAL PRIMARY KEY,
    author_id   INT NOT NULL,
    genre_id    INT NOT NULL,
    title       TEXT NOT NULL,
    isbn        TEXT,
    filename    TEXT,
    description TEXT,
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES authors(id),
    FOREIGN KEY (genre_id)  REFERENCES genres(id)
);