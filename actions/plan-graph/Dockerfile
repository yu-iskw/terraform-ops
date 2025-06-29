# Single-stage build for terraform-ops plan-graph command
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
COPY actions/plan-graph/entrypoint.sh .

# Make entrypoint executable
RUN chmod +x entrypoint.sh

# Note: No healthcheck needed for GitHub Actions as containers are short-lived
# and the action completes when the entrypoint script finishes

# Set the entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]
