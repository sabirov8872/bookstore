SELECT id,
       username,
       password,
       email,
       phone,
       userrole
FROM users
WHERE id = $1