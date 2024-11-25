CREATE TABLE users (
    id        SERIAL PRIMARY KEY,
    username  VARCHAR(255) NOT NULL UNIQUE,
    password  VARCHAR(255) NOT NULL,
    email     VARCHAR(255) UNIQUE,
    phone     VARCHAR(255) UNIQUE,
    role      VARCHAR(255) NOT NULL
);

CREATE TABLE authors (
    id      SERIAL PRIMARY KEY,
    author  VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE genres (
    id     SERIAL PRIMARY KEY,
    genre  VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE books (
    id           SERIAL PRIMARY KEY,
    author_id    INT NOT NULL,
    genre_id     INT NOT NULL,
    bookname     VARCHAR(255) NOT NULL,
    isbn         varchar(255) not null,
    filename     VARCHAR(255),
    FOREIGN KEY (author_id)   REFERENCES authors(id),
    FOREIGN KEY (genre_id)    REFERENCES genres(id)
);