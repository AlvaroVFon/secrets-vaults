DELETE FROM consumers WHERE apikey = 'superadmin-api-key';
DELETE FROM roles WHERE name IN ('superadmin', 'admin');
