CREATE TABLE IF NOT EXISTS roles (
  ID uuid PRIMARY KEY,
  Name varchar(255) UNIQUE
);
