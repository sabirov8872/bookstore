insert into books (author_id,
                   genre_id,
                   name,
                   isbn,
                   filename,
                   description,
                   created_at,
                   updated_at)
values($1, $2, $3, $4, $5, $6, $7, $8)
returning id
