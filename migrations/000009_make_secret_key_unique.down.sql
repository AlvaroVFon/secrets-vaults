ALTER TABLE secrets DROP CONSTRAINT IF EXISTS secrets_key_key;
ALTER TABLE secrets ADD CONSTRAINT uq_key_consumer_id UNIQUE (key, consumer_id);
