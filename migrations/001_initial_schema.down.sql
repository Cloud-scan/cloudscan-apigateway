-- CloudScan API Gateway - Rollback Initial Schema

DROP TRIGGER IF EXISTS update_api_keys_updated_at ON api_keys CASCADE;
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects CASCADE;
DROP TRIGGER IF EXISTS update_users_updated_at ON users CASCADE;
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations CASCADE;

DROP TABLE IF EXISTS api_keys CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;

DROP FUNCTION IF EXISTS update_updated_at_column CASCADE;

DROP EXTENSION IF EXISTS "uuid-ossp";