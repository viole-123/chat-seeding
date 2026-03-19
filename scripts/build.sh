#!/bin/bash
set -e

echo "🔧 Initializing database..."

# Wait for PostgreSQL to be ready
until pg_isready -U postgres; do
  echo "⏳ Waiting for PostgreSQL to be ready..."
  sleep 2
done

# Create database if not exists
psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'uniscore_seeding'" | grep -q 1 || psql -U postgres -c "CREATE DATABASE uniscore_seeding"

echo "✅ Database initialized successfully"