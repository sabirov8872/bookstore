UPDATE users
SET username = $1,
    password = $2,
    email = $3,
    phone = $4
WHERE id = $5