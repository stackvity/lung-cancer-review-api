# Makefile for Patient-Centric AI Lung Cancer Review System - Backend API
#
# Defines common development tasks for building, running, testing, and managing the backend API.
# This Makefile streamlines the development workflow, promotes best practices, and ensures consistency across different operations.
# It is designed to be highly informative and user-friendly, guiding developers through common tasks with clear instructions and feedback.
#
# Tasks Include:
#   - build: Builds the Go application binary, incorporating code linting for quality assurance.
#   - run: Builds and runs the application locally using Docker Compose, setting up a complete development environment.
#   - test-unit: Runs Go unit tests specifically, allowing for focused unit-level testing.
#   - test-integration: Runs Go integration tests specifically, enabling targeted integration testing.
#   - test: Runs all tests (unit and integration sequentially) for comprehensive testing coverage.
#   - lint: Runs code linters to enforce code quality and coding standards, ensuring code maintainability.
#   - migrate-up: Applies database migrations to the latest version, updating the database schema to the latest definitions.
#   - migrate-down: Rolls back database migrations by one step, reverting the last schema change (useful for debugging and development).
#   - migrate-force: Forces database migration to a specific version (for development/debugging), bypassing normal migration flow for advanced schema management.
#   - clean: Cleans up build artifacts and dependencies, removing compiled binaries and tidying Go modules to maintain a clean workspace.
#   - help: Displays help information for Makefile targets, providing detailed usage instructions and documentation within the Makefile itself.
#
# Prerequisites:
#   - Go toolchain (version specified in go.mod) installed and correctly configured in your development environment.
#   - Docker and Docker Compose installed and running (required for the 'run' target and database migrations).
#   - golangci-lint installed and accessible in your system's PATH (required for the 'lint' target).
#   - migrate CLI tool installed and accessible in your system's PATH (required for the 'migrate-*' targets).


# --- Project Configuration ---
# Define project-specific variables for customization and consistency.
# These variables are used throughout the Makefile to configure build, run, and test commands.
APP_NAME = lung-cancer-review-api         # Name of the application binary executable.
BINARY_NAME = $(APP_NAME)                # Output binary name after compilation (defaults to APP_NAME, customizable if needed).
DOCKER_COMPOSE_FILE = docker-compose.yml     # Path to the Docker Compose file, defining services for local development.
GOLANGCI_LINT_VERSION = v1.54.2              # Specify golangci-lint version for consistent linting across development environments.
MIGRATE_VERSION_TARGET ?= latest           # Default target migration version for 'migrate-force', defaults to 'latest' if not explicitly set during command invocation.


# --- Build Targets ---

## Build the application binary, incorporating code linting for quality assurance
build: lint # Build target depends on 'lint' target, ensuring code quality checks are performed before building the binary
	@echo "--> Building backend API binary..." # Informative echo message to the console, indicating the start of the build process
	### Best Practice: Robust Error Handling ###
	@if ! bash scripts/build.sh; then # Execute the build script (scripts/build.sh) and check its exit status using 'if ! ...; then ...; fi' for robust error handling
		echo "❌ Backend API build failed. See scripts/build.sh output for details." # Error message displayed if the build script fails, directing users to check script output
		exit 1 # Exit immediately with error code 1 to signal build failure and stop further Makefile execution
	fi
	@echo "✅ Backend API built successfully: $(BINARY_NAME)" # Success message displayed upon successful build completion, including the name of the compiled binary

## Run the application using Docker Compose, ensuring application is built first
run: build # 'run' target depends on 'build' to ensure the application is built before running
	@echo "--> Running backend API with Docker Compose..." # Informative echo message indicating Docker Compose startup
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up # Execute docker-compose up command to start all services defined in docker-compose.yml (API and database)
	@echo "✅ Backend API running - access at http://localhost:8080" # Success message providing the URL to access the running API, improving user experience

## Run unit tests specifically, focusing on individual components and functions
test-unit:
	@echo "--> Running unit tests..." # Informative echo message for unit tests execution
	@go test -v ./... -run=TestUnit # Execute unit tests using 'go test -v ./...' command, with '-run=TestUnit' flag to target unit tests by name
	@echo "✅ Unit tests finished" # Confirmation message displayed after unit test completion

## Run integration tests specifically, focusing on component interactions and external dependencies
test-integration:
	@echo "--> Running integration tests..." # Informative echo message for integration tests execution
	@go test -v ./... -tags=integration -run=TestIntegration # Execute integration tests using 'go test -v ./...' command, with '-tags=integration' and '-run=TestIntegration' flags
	@echo "✅ Integration tests finished" # Confirmation message after integration test completion

## Run all tests (unit and integration tests sequentially) for comprehensive testing
test: test-unit test-integration # 'test' target depends on both 'test-unit' and 'test-integration', ensuring both test suites are run
	@echo "--> Running all tests: Unit and Integration Tests" # Informative echo message for combined test execution
	@echo "✅ All tests finished" # Confirmation message after combined test execution

## Run code linters to check for code quality and style adherence
lint:
	@echo "--> Running code linters..." # Informative echo message before linting process
	@golangci-lint run ./... # Execute golangci-lint command to analyze codebase for linting errors and code quality issues
	@echo "✅ Code linting finished" # Confirmation message after linting process is complete

## Apply database migrations to the latest version, updating the database schema
migrate-up:
	@echo "--> Applying database migrations..." # Informative echo message before database migration
	@docker-compose -f $(DOCKER_COMPOSE_FILE) run --rm api scripts/migrate.sh up # Execute database migrations using 'migrate' CLI tool, within the 'api' Docker container for correct environment
	@echo "✅ Database migrations applied to latest version" # Confirmation message after successful database migration

## Rollback database migrations by one step, reverting the last schema change
migrate-down:
	@echo "--> Rolling back database migrations by one step..." # Informative echo message before database migration rollback
	@docker-compose -f $(DOCKER_COMPOSE_FILE) run --rm api scripts/migrate.sh down 1 # Execute database migration rollback using 'migrate' CLI tool, rolling back a single step
	@echo "✅ Database migrations rolled back by one step" # Confirmation message after successful database migration rollback

## Force database migration to a specific version (for development/debugging), bypassing normal migration flow
migrate-force:
	@echo "--> Force database migration to version $(MIGRATE_VERSION_TARGET)..." # Informative message indicating forced migration and target version
	@docker-compose -f $(DOCKER_COMPOSE_FILE) run --rm api scripts/migrate.sh force $(MIGRATE_VERSION_TARGET) # Execute forced database migration to a specific version using 'migrate' CLI tool
	@echo "✅ Database migrations forced to version $(MIGRATE_VERSION_TARGET)" # Confirmation message after forced database migration

## Clean build artifacts and dependencies, removing compiled binary and tidying Go modules
clean:
	@echo "--> Cleaning build artifacts and dependencies..." # Informative echo message before cleanup
	@go clean # Execute 'go clean' command to remove compiled binaries and temporary build files
	@rm -f $(BINARY_NAME) # Remove the compiled binary executable file specifically
	@go mod tidy # Execute 'go mod tidy' to tidy up Go module dependencies, removing unused modules and ensuring go.mod and go.sum are consistent
	@echo "✅ Cleaned build artifacts and dependencies" # Confirmation message after cleanup process is finished

# --- Helper Targets (Optional) ---
# These targets are for convenience and do not represent core functionalities.

## Display help information for Makefile targets, providing usage instructions
help: # Help target - Provides usage instructions for common Makefile commands
	@echo "Usage: make <target> [VARIABLE=value]" # General usage instruction, showing how to override variables
	@echo "" # Blank line for better readability
	@echo "Targets:" # Section header for targets
	@echo "  build         - Build the backend API binary. Depends on 'lint'." # Help text for 'build' target, indicating dependency
	@echo "  run           - Build and run the application using Docker Compose. Depends on 'build'." # Help text for 'run' target, indicating dependency and execution method
	@echo "  test          - Run all tests (unit and integration) sequentially." # Help text for 'test' target, clarifying execution order
	@echo "  test-unit     - Run unit tests only. Faster for focused testing." # Help text for 'test-unit' target, explaining use case
	@echo "  test-integration - Run integration tests only. Useful for testing external dependencies." # Help text for 'test-integration' target, explaining use case
	@echo "  lint          - Run code linters. Enforces code quality and style." # Help text for 'lint' target, highlighting its purpose
	@echo "  migrate-up    - Apply database migrations to the latest version. Updates database schema." # Help text for 'migrate-up' target, clarifying its action
	@echo "  migrate-down  - Rollback database migrations by one step. Reverts last schema change." # Help text for 'migrate-down' target, clarifying its action and use case
	@echo "  migrate-force VERSION=N - Force database migration to version N. For development/debugging, bypasses normal flow." # Help text for 'migrate-force' target, including variable usage and use case
	@echo "                Example: make migrate-force VERSION=3" # Example usage for 'migrate-force' target, showing how to set VERSION variable
	@echo "  clean         - Clean build artifacts and dependencies. Resets build environment." # Help text for 'clean' target, explaining its effect
	@echo "" # Blank line for better readability
	@echo "Variables:" # Section header for variables
	@echo "  MIGRATE_VERSION_TARGET - Target version for migrate-force (default: latest). Example: VERSION=3" # Help text for 'MIGRATE_VERSION_TARGET' variable, showing example of overriding
	@echo "" # Blank line for better readability
	@echo "Example:" # Section header for example usage
	@echo "  make build" # Example command - Build the application
	@echo "  make run" # Example command - Run the application with Docker Compose
	@echo "  make test" # Example command - Run all tests
	@echo "  make lint" # Example command - Run code linters
	@echo "  make migrate-up" # Example command - Apply database migrations
	@echo "  make migrate-down" # Example command - Rollback database migrations
	@echo "  make migrate-force VERSION=0003" # Example command - Force database migration to version 0003
	@echo "  make clean" # Example command - Clean build artifacts


# --- Variables with Default Values (Optional) ---
# You can define variables with default values here, if needed.
# Example:
# DB_USER ?= lung_cancer_review_user   # Default database user (can be overridden)


# --- Dependency Management (Optional) ---
# You can add targets for dependency management if needed.
# Example:
# dependencies:
#  @echo "--> Checking and updating dependencies..."
#  @go mod tidy
#  @go mod vendor
#  @echo "✅ Dependencies updated"

# --- Phony Targets ---
# Declare phony targets to prevent naming conflicts with files.
.PHONY: build run test test-unit test-integration lint migrate-up migrate-down migrate-force clean help