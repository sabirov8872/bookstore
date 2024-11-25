INSERT INTO users (username,
                   password,
                   email,
                   phone,
                   role)
VALUES ($1, $2, $3, $4, $5)
RETURNING id