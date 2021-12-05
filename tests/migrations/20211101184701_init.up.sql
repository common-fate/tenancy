
-- 
--  Proivision the rules and permissions for our RLS configuration and database user
-- 



CREATE OR REPLACE FUNCTION set_tenant(tenant_id text) RETURNS void AS $$
BEGIN
-- Setting this to true means that the effect of this function only lasts as long as the transaction, after which the value is set to null
    PERFORM set_config('app.current_tenant', tenant_id, false);
END;
$$ LANGUAGE plpgsql;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO tenant;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO tenant;


CREATE TABLE IF NOT EXISTS tenants (
    id uuid PRIMARY KEY
);

ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenants_isolation_policy ON tenants
USING (CAST(id AS TEXT) = current_setting('app.current_tenant'));


CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants ON DELETE CASCADE
);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY users_isolation_policy ON users
USING (CAST(tenant_id AS TEXT) = current_setting('app.current_tenant'));

