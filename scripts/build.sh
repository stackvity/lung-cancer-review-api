#!/bin/bash
# scripts/build.sh

set -e  # Exit immediately if a command exits with a non-zero status - Ensures script stops on error
set -x  # Enable tracing, print each command before executing - For debugging and verbose output during build process

# --- Project Configuration ---
# Define application name and binary output name for build process
APP_NAME="lung-cancer-review-api" # Application binary name - Consistent naming for binary executable
BINARY_NAME="${APP_NAME}"        # Output binary name - Can be customized if needed

# --- Build Stage ---
echo "Starting backend API build process..." # Informative message at the start of the build process

# Navigate to the backend directory - Ensure script is executed from the backend root directory for correct path resolution
cd backend || { echo "Error: Could not change directory to backend"; exit 1; } # Navigate to backend directory, exit if fails

echo "Current directory: $(pwd)" # Print current directory for debugging purposes - Helpful for verifying script execution location

# --- Code Formatting ---
echo "Formatting Go code with gofmt..." # Informative message before code formatting
gofmt -w . # Format Go code in the current directory and subdirectories using gofmt - Ensures consistent code style across the project
echo "gofmt completed." # Confirmation message after gofmt execution

# --- Code Linting ---
echo "Running code linters with golangci-lint..." # Informative message before linting process
golangci-lint run ./... # Run code linters on the entire project using golangci-lint - Enforces code quality standards and catches potential issues
echo "golangci-lint completed." # Confirmation message after golangci-lint execution

# --- Application Build with Version Information Embedding ---
echo "Building application binary: ${BINARY_NAME} with version information..." # Informative message indicating application build start with version info

# Retrieve version information from Git and current UTC build timestamp - Best practice for application versioning and traceability
VERSION=$(git describe --tags --always) # Retrieve version information from Git tags, fallbacks to commit hash if no tags are found
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") # Get current UTC build timestamp in ISO 8601 format

# Build the Go application binary with ldflags to embed version and build time information
# -ldflags "-X 'main.version=${VERSION}' -X 'main.buildTime=${BUILD_TIME}'":  Uses ldflags to pass version and build time as build-time variables
# -o ${BINARY_NAME}: Specifies the output binary name
# cmd/lung-cancer-review-api/main.go: Specifies the main Go file to build
go build -ldflags "-X 'main.version=${VERSION}' -X 'main.buildTime=${BUILD_TIME}'" -o ${BINARY_NAME} cmd/lung-cancer-review-api/main.go
echo "Application build completed successfully with version information: ${BINARY_NAME}" # Confirmation message upon successful build with version info

echo "Backend API build process finished." # Informative message at the end of the build process

# --- Post-Build Actions (Optional - for future enhancements and security hardening) ---
# This section is for OPTIONAL actions that can be performed AFTER the application binary is built.
# These actions are NOT REQUIRED for the core functionality of the application build process
# but are included as best practices for enhanced security, supply chain integrity, and future scalability.
# They are commented out in this version and can be enabled or extended in future sprints as needed.

# --- Post-Build Action: SBOM Generation (Optional - for future sprints - Supply Chain Security Enhancement) ---
# Generate Software Bill of Materials (SBOM) for supply chain security and vulnerability management.
# SBOM provides a comprehensive inventory of all software components used in the application, enhancing transparency and security.
# Example using syft - Requires syft binary to be available in the build environment (e.g., installed in Docker image or CI environment).
echo "Generating Software Bill of Materials (SBOM) using Syft - Optional Feature - To be implemented in future sprints" # Informative message indicating SBOM generation step
# SBOM_FILE="${BINARY_NAME}.sbom.spdx.json" # Example: Define SBOM output file name with .spdx.json extension (SPDX JSON format)
# echo "Generating SBOM to file: ${SBOM_FILE}" # Informative message indicating SBOM output file
# # Placeholder command - Replace with actual SBOM generation command (e.g., syft binary ${BINARY_NAME} -o spdx -f json > ${SBOM_FILE})
# echo "syft binary ${BINARY_NAME} -o spdx -f json > ${SBOM_FILE}" # Example placeholder command - Commented out
# echo "SBOM generation completed (placeholder): ${SBOM_FILE}" # Confirmation message after SBOM generation (placeholder)

# --- Post-Build Action: Artifact Signing (Optional - for future sprints - Enhanced Integrity Verification) ---
# Sign the application binary or Docker image for enhanced security and integrity verification.
# Artifact signing allows for cryptographic verification of the build artifact's authenticity and integrity, protecting against tampering.
# Example using placeholder command - Replace with actual artifact signing tool (e.g., cosign, gpg) in future sprints.
echo "Signing application binary - Optional Feature - To be implemented in future sprints" # Informative message indicating artifact signing step
# SIGNED_BINARY="${BINARY_NAME}.sig" # Example: Define signed binary output file name with .sig extension
# echo "Artifact signing command placeholder - Replace with actual signing tool execution" # Placeholder message
# # Placeholder command - Replace with actual artifact signing command (e.g., cosign sign --key <key> ${BINARY_NAME} -o ${SIGNED_BINARY})
# echo "cosign sign --key <key> ${BINARY_NAME} -o ${SIGNED_BINARY}" # Example placeholder command - Commented out
# echo "Artifact signing completed (placeholder): ${SIGNED_BINARY}" # Confirmation message after artifact signing (placeholder)

exit 0 # Exit script successfully - Indicates successful execution of the build script