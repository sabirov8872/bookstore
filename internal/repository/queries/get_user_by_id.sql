SELECT id,
       username,
       password,
       email,
       phone,
       role
FROM users
WHERE id = $1