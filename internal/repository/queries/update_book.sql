update books
set author_id = $1,
    genre_id = $2,
    bookname = $3,
    isbn = $4
where id = $5