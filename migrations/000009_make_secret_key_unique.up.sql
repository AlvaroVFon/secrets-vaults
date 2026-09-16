ALTER TABLE secrets DROP CONSTRAINT IF EXISTS uq_key_consumer_id;
ALTER TABLE secrets ADD CONSTRAINT secrets_key_key UNIQUE (key);
