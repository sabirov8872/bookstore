create table if not exists users (
    id serial primary key,
    username varchar(255) not null unique,
    password varchar(255) not null,
    userrole varchar(255) not null
);