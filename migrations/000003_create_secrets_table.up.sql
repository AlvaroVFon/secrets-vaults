CREATE TABLE IF NOT EXISTS secrets (
  id uuid PRIMARY KEY,
  key varchar(255),
  value varchar(255),
  consumer_id uuid,
  FOREIGN KEY(consumer_id) REFERENCES consumers(id),
  CONSTRAINT uq_key_consumer_id UNIQUE (key, consumer_id)
);
