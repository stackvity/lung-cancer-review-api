# Dockerfile - Multi-stage Build for Patient-Centric AI Lung Cancer Review System Backend API

# -------- BUILDER STAGE: Compile the Go Application --------
# Use the official Go image (version 1.23.6 based on Debian Bullseye) as the builder stage base image.
# This stage is responsible for compiling the Go application and preparing the binary for deployment.
FROM golang:1.23.6-bullseye AS builder

# Set the working directory inside the builder container to /app - Best practice for Go projects in Docker.
WORKDIR /app

# --- OPTIMIZATION: Dependency Caching ---
# Copy go module definition files (go.mod and go.sum) first for caching - Leverages Docker layer caching for faster, incremental builds.
# This ensures that dependency download step is only re-executed when dependencies change (go.mod/go.sum), significantly speeding up subsequent builds.
COPY go.mod go.sum ./

# Download Go dependencies - leveraging Docker's build cache for efficiency.
# Downloads Go modules as defined in go.mod and go.sum. The -x flag enables verbose output for debugging.
RUN go mod download -x

# --- OPTIONAL BUILD OPTIMIZATION for large codebases ---
# [Performance Tip for Large Codebases]: Consider using COPY --link . . for potentially faster builds.
# For very large Go codebases, using COPY --link . . instead of COPY . . may offer significant build performance improvements.
# COPY --link creates copy-on-write layers, which can be faster than full copies, especially for large directories with many files.
# Uncomment the line below if build times become a bottleneck for large codebases.
# COPY --link . .

# Copy the entire backend source code into the container - Copies all source files into the /app directory after dependency management.
COPY . .

# --- SCRIPT COPYING and EXECUTION PERMISSIONS ---
# Copy necessary scripts for the build and migration processes - Ensures consistent script management and version control.
COPY scripts/build.sh /app/scripts/
COPY scripts/migrate.sh /app/scripts/
# Ensure scripts are executable - Essential for running the build and database migration scripts within the container.
RUN chmod +x /app/scripts/build.sh /app/scripts/migrate.sh

# --- BUILD STAGE ENVIRONMENT VARIABLES ---
# Set environment variables for the build process - Consistent with .env.example and Render deployment configurations.
# These environment variables are specific to the BUILD stage and will not be present in the final deploy image unless explicitly carried over.
ARG ENVIRONMENT=production
ENV ENVIRONMENT=${ENVIRONMENT}

# --- APPLICATION BUILD COMMAND ---
# Build the Go application binary - Executes the build.sh script to compile the Go backend application.
# This encapsulates the build logic within a separate, version-controlled script for consistency and maintainability.
RUN scripts/build.sh


# -------- DEPLOY STAGE: Create Minimal Runtime Image --------
# Start from a minimal Alpine Linux base image for the deploy stage - Alpine is chosen for its small size, enhancing security and reducing the final image footprint.
FROM alpine:latest AS deploy

# Set working directory for the deploy stage - Best practice for container image organization and clarity.
WORKDIR /app

# --- BINARY COPY ---
# Copy the compiled binary from the builder stage - Creates a leaner and more secure image by copying only the necessary executable.
COPY --from=builder /app/lung-cancer-review-api /app/lung-cancer-review-api

# --- ASSETS COPY ---
# Copy database migration scripts - Required for running database migrations on container startup to ensure database schema is up-to-date.
COPY migrations /app/migrations

# Copy report templates - Required for PDF report generation functionality, ensuring report generation is functional in the deployed container.
COPY internal/pdf/templates /app/templates

# Copy sqlc configuration - Required for running database migrations, as migrations might depend on sqlc configuration for database interactions.
COPY sqlc.yaml /app/

# --- NETWORK CONFIGURATION ---
# Expose the port the app runs on - Consistent with .env.example and internal/config/config.go (default port 8080).
EXPOSE 8080

# --- RUNTIME ENVIRONMENT VARIABLES - SECURITY WARNING - CRITICAL ---
# ==================================================================================================================================================
# --- !!!  CRITICAL SECURITY WARNING: SECURE ENVIRONMENT VARIABLE INJECTION IS MANDATORY FOR PRODUCTION !!! ---
# ==================================================================================================================================================
# Sensitive environment variables like DB credentials, API keys, and encryption keys
# MUST NOT be hardcoded directly in this Dockerfile. Doing so creates a MAJOR SECURITY VULNERABILITY.
# Instead, SECURELY INJECT these variables at RUNTIME via your deployment platform's secrets management features
# (e.g., Render Environment Variables, Kubernetes Secrets, AWS Secrets Manager, Google Cloud Secret Manager, Azure Key Vault).
# The following ENV declarations are provided as EXAMPLES and PLACEHOLDERS for LOCAL DEVELOPMENT and TESTING PURPOSES ONLY.
# In PRODUCTION DEPLOYMENTS, REMOVE or OVERRIDE these EXAMPLE ENV declarations and rely EXCLUSIVELY on SECURE SECRETS INJECTION.
# ==================================================================================================================================================
ENV ENVIRONMENT=production
ENV DB_DRIVER=postgres
# DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSL_MODE: SECURELY INJECT THESE AT RUNTIME (e.g., Render Environment Variables) - Database connection details - SECURITY SENSITIVE - *MANDATORY EXTERNAL INJECTION*
# Example value - Database connection pool setting - Can be overridden via secure environment variables if needed
ENV DB_MAX_OPEN_CONNS=25
# Example value - Database connection pool setting - Can be overridden via secure environment variables if needed
ENV DB_MAX_IDLE_CONNS=5
# Example value - Database connection pool setting - Can be overridden via secure environment variables if needed
ENV DB_CONN_MAX_LIFETIME=30m
# Example value - Database connection pool setting - Can be overridden via secure environment variables if needed
ENV DB_CONN_MAX_IDLE_TIME=5m
# Server listening address and port - Can be overridden via secure environment variables if needed
ENV HTTP_SERVER_ADDRESS=:8080
# SENTRY_DSN, GEMINI_API_KEY, CLOUD_STORAGE_BUCKET, FILE_ENCRYPTION_KEY: SECURELY INJECT THESE AT RUNTIME via SECRETS MANAGEMENT - API Keys and Secrets - SECURITY SENSITIVE - *MANDATORY EXTERNAL INJECTION*
# Default logging level - Can be overridden via secure environment variables if needed
ENV LOG_LEVEL=info
ENV LOG_FORMAT=json
# Default storage type: cloud - Can be overridden via secure environment variables if needed
ENV STORAGE_TYPE=cloud
# CLOUD_STORAGE_BUCKET, AWS_REGION, FILE_ENCRYPTION_KEY are expected to be set in the runtime environment (secrets management) - Cloud Storage Configuration - SECURITY SENSITIVE - *MANDATORY EXTERNAL INJECTION*
# Default link expiration duration - Can be overridden via secure environment variables if needed
ENV LINK_EXPIRATION=24h
# Default data retention duration - Can be overridden via secure environment variables if needed
ENV DATA_RETENTION=90d
# Path to report templates in deploy stage - Adjust path if templates are moved
ENV REPORT_TEMPLATE_PATH=/app/templates

# --- Security Hardening ---
# Set user to non-root - Security best practice for containerized applications - Minimize container privileges - Reduces risk of container breakout vulnerabilities
USER nonroot:nonroot

# --- Health Check Configuration ---
# Healthcheck to ensure container is running and API is responsive - Essential for container orchestration and monitoring in production environments
HEALTHCHECK --interval=30s --timeout=5s --retries=3 CMD wget -q -O - http://localhost:8080/api/v1/health || exit 1

# --- Startup Command ---
# Command to run migrations and then start the API server - Combined for simplified Dockerfile CMD - Streamlines container startup process
# Combines database migrations and API server start in a single entrypoint for simplicity and operational efficiency in Dockerfile deployments
CMD /app/scripts/migrate.sh up && /app/lung-cancer-review-api