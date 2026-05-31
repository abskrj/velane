-- Create the nango database if it doesn't already exist.
-- This script runs once on first Postgres initialization (empty volume).
-- If you already have a postgres_data volume without the nango DB, run
-- `make down` first to wipe it, or connect and run: CREATE DATABASE nango;
SELECT 'CREATE DATABASE nango'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'nango'
)\gexec
