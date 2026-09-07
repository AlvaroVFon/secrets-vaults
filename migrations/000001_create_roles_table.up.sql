CREATE TABLE IF NOT EXISTS roles (
  id uuid PRIMARY KEY,
  name varchar(255) UNIQUE
);
