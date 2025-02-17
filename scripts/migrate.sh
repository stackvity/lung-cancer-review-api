#!/bin/bash
# scripts/migrate.sh

set -e  # Exit immediately if a command exits with a non-zero status
set -x  # Enable tracing, print each command before executing

# --- Script Description ---
# This script applies database migrations for the Patient-Centric AI Lung Cancer Review System Backend API.
# It uses golang-migrate/migrate command-line tool to manage database schema changes.
# Intended for development, deployment, rollback, and automated CI/CD scenarios.

# --- Configuration Variables ---
# MIGRATIONS_PATH: Path to migrations directory (consistent with backend folder structure).
MIGRATIONS_PATH="./migrations"
# DATABASE_URL_FILE: Path to Docker Secret containing database URL - **ENHANCED SECURITY** - Recommendation 1
DATABASE_URL_FILE="/run/secrets/database_url"

# --- Check for Migrate CLI Tool ---
# Verify if 'migrate' command is installed and in PATH; exit if not found.
if ! command -v migrate &> /dev/null
then
    echo "Error: migrate command not found. Ensure golang-migrate/migrate CLI is installed and in PATH."
    exit 1
fi
echo "migrate command found: $(migrate --version)"

# --- Database Migration Command Execution ---
# Execute database migrations using 'migrate' CLI tool.
# -path: Path to migrations directory.
# -database: Database connection URL (read from Docker Secret for enhanced security). - Recommendation 1
# up: Apply all pending 'up' migrations.

echo "Applying database migrations..."

# --- ENHANCEMENT: Securely Read DATABASE_URL from Docker Secret File - Recommendation 1 ---
# Read DATABASE_URL from Docker Secret file for enhanced security in production.
DATABASE_URL=$(cat "${DATABASE_URL_FILE}")

# --- ENHANCEMENT: Capture and Check Migrate Command Output for Detailed Error Handling - Recommendation 2 ---
# Execute 'migrate up' command and capture output for detailed error reporting.
MIGRATE_OUTPUT=$(migrate -path "${MIGRATIONS_PATH}" -database "${DATABASE_URL}" up)

# Check if the migration command was successful (exit code 0).
if [[ $? -ne 0 ]]; then
    echo "Error: Database migrations failed." # General error message for migration failure
    echo "Detailed Migration Output:"
    echo "${MIGRATE_OUTPUT}"

    # --- ENHANCEMENT: Rollback Migration on Failure - Recommendation 3 ---
    # Implement automated rollback to revert database changes in case of migration failure.
    echo "Attempting to rollback database migration..." # Informative message before rollback attempt
    MIGRATE_ROLLBACK_OUTPUT=$(migrate -path "${MIGRATIONS_PATH}" -database "${DATABASE_URL}" down 1) # Execute 'migrate down 1' to rollback last migration
    if [[ $? -ne 0 ]]; then # Check exit status of rollback command
        echo "Error: Database migration rollback failed." # Error message if rollback also fails
        echo "Rollback Output: ${MIGRATE_ROLLBACK_OUTPUT}" # Print rollback output for debugging
        echo "Please check database state and consider manual rollback if necessary." # Guidance for manual intervention
    else
        echo "Database migration rolled back successfully." # Confirmation message if rollback is successful
    fi
    exit 1 # Exit with error code 1 after migration failure (and attempted rollback)
fi

echo "Database migrations applied successfully." # Confirmation message for successful migration

exit 0 # Exit script successfully