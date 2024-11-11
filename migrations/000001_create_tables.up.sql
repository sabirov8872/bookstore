create table if not exists users (
    id serial primary key,
    username varchar not null unique,
    password varchar not null,
    email varchar not null unique,
    phone varchar not null unique,
    userrole varchar not null
);

create table if not exists books (
    id serial primary key,
    bookname varchar not null,
    genre varchar not null,
    author varchar not null
);