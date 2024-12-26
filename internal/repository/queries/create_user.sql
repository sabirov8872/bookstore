INSERT INTO users (role_id,
                   username,
                   password,
                   email,
                   phone)
VALUES ($1, $2, $3, $4, $5)
RETURNING id