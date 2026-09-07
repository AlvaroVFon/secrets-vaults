CREATE TABLE IF NOT EXISTS consumers (
  id uuid PRIMARY KEY,
  name varchar(255),
  apikey varchar(255) UNIQUE,
  role_id uuid,
  active boolean DEFAULT FALSE,
  FOREIGN KEY (role_id) REFERENCES roles(id)
);
