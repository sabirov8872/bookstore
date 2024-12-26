SELECT u.id,
       u.username,
       u.password,
       u.email,
       u.phone,
       r.name
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.id = $1