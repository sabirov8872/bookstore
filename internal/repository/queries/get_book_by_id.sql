SELECT b.id,
       b.bookname,
       a.author,
       g.genre,
       b.isbn,
       b.filename
FROM books b
JOIN authors a ON b.author_id = a.id
JOIN genres g ON b.genre_id = g.id
WHERE b.id = $1