INSERT INTO users (username,
                   password,
                   email,
                   phone,
                   userrole)
VALUES ($1, $2, $3, $4, $5)
RETURNING id