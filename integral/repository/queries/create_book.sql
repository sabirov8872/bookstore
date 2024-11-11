insert into books (bookname,
                   genre,
                   author)
values ($1, $2, $3)
RETURNING id