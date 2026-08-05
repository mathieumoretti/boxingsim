#!/bin/sh
set -e

echo "Initializing PostgreSQL database for Boxing Simulator..."

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER boxing WITH PASSWORD 'boxing123';

    ALTER USER boxing CREATEDB;

    ALTER DATABASE boxing OWNER TO boxing;

    GRANT ALL PRIVILEGES ON DATABASE boxing TO boxing;

    GRANT ALL PRIVILEGES ON SCHEMA public TO boxing;
EOSQL

echo "Database initialization completed successfully."