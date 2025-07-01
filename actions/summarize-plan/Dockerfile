# Single-stage build for terraform-ops summarize-plan command
FROM golang:1.24.4-alpine

# Install build dependencies and runtime requirements
RUN apk add --no-cache git ca-certificates tzdata

# Note: Running as root user is acceptable for GitHub Actions containers
# as they are ephemeral and run in isolated environments

# Set working directory
WORKDIR /src

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o terraform-ops ./cmd/terraform-ops

# Create app directory
WORKDIR /app

# Copy the binary to app directory
RUN cp /src/terraform-ops .

# Copy entrypoint script
COPY actions/summarize-plan/entrypoint.sh .

# Make entrypoint executable
RUN chmod +x entrypoint.sh

# Change ownership to non-root user
RUN chown -R terraform-ops:terraform-ops /app

# Switch to non-root user
USER terraform-ops

# Add healthcheck for the application
HEALTHCHECK NONE

# Set the entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]
