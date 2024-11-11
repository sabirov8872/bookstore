update books
set bookname = $1,
    genre = $2,
    author = $3
where id = $4
