SELECT u.id,
       r.name,
       u.username,
       u.password,
       u.email,
       u.phone
FROM users u
JOIN roles r ON r.id = u.role_id
