UPDATE users
SET username = $1,
    password = $2,
    email = $3,
    phone = $4,
    session_id = $5
WHERE id = $6