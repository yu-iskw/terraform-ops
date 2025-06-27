# Multi-stage build for terraform-ops plan-graph command
FROM golang:1.24.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

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

# Final stage - create minimal runtime image
FROM alpine:3.19

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create app directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /src/terraform-ops .

# Copy entrypoint script
COPY actions/plan-graph/entrypoint.sh .

# Make entrypoint executable
RUN chmod +x entrypoint.sh

# Set the entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]
