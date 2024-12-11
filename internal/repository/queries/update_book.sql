update books
set author_id = $1,
    genre_id = $2,
    title = $3,
    isbn = $4,
    description = $5,
    updated_at = $6
WHERE id = $7

