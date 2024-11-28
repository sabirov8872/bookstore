SELECT books.id,
       books.name,
       authors.id,
       authors.name,
       genres.id,
       genres.name,
       books.isbn,
       books.filename,
       books.description,
       books.created_at,
       books.updated_at
FROM books
JOIN authors ON books.author_id = authors.id
JOIN genres ON books.genre_id = genres.id
WHERE books.id = $1