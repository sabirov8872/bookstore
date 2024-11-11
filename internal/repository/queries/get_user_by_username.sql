SELECT id,
       password,
       userrole
FROM users
WHERE username = $1