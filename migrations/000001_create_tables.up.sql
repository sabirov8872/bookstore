create table if not exists users (
    id serial primary key,
    username varchar unique,
    password varchar,
    email varchar unique,
    phone varchar unique,
    userrole varchar
);

create table if not exists books (
    id serial primary key,
    bookname varchar,
    genre varchar,
    author varchar
);