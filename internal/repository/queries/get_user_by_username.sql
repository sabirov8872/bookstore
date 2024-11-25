SELECT id,
       password,
       role
FROM users
WHERE username = $1