CREATE TABLE users (
    id        SERIAL PRIMARY KEY,
    username  TEXT NOT NULL UNIQUE,
    password  TEXT NOT NULL,
    email     TEXT UNIQUE,
    phone     TEXT UNIQUE,
    role      TEXT NOT NULL
);

CREATE TABLE authors (
    id      SERIAL PRIMARY KEY,
    name  TEXT NOT NULL UNIQUE
);

CREATE TABLE genres (
    id     SERIAL PRIMARY KEY,
    name  TEXT NOT NULL UNIQUE
);

CREATE TABLE books (
    id           SERIAL PRIMARY KEY,
    author_id    INT NOT NULL,
    genre_id     INT NOT NULL,
    name         TEXT NOT NULL,
    isbn         TEXT NOT NULL,
    filename     TEXT NOT NULL,
    FOREIGN KEY (author_id) REFERENCES authors(id),
    FOREIGN KEY (genre_id) REFERENCES genres(id)
);