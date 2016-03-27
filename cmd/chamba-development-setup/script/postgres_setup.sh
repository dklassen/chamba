#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE USER chamba_user;
    CREATE DATABASE chamba;
    GRANT ALL PRIVILEGES ON DATABASE chamba TO chamba_user;
EOSQL
