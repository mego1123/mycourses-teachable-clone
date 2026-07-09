DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "pgcrypto";
