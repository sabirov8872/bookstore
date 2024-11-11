select id,
       bookname,
       genre,
       author
from books
where id = $1