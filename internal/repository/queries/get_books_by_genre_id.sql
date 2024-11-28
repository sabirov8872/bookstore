SELECT b.id,
       b.name,
       a.id,
       a.name,
       g.id,
       g.name,
       b.isbn,
       b.filename,
       b.description,
       b.created_at,
       b.updated_at
FROM books b
JOIN authors a ON b.author_id = a.id
JOIN genres g ON b.genre_id = g.id
where genre_id = $1