UPDATE users
SET username = $1,
    password = $2,
    email = $3,
    phone = $4,
    userrole = $5
WHERE id = $6