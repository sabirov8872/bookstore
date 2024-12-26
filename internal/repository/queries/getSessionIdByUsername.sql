SELECT u.id,
       u.password,
       r.name
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.username = $1