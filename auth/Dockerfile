# Use the official Golang image as the base image
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Use a smaller base image for the final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Create non-root user
RUN addgroup -g 1001 appuser && \
    adduser -D -u 1001 -G appuser appuser && \
    chown -R appuser:appuser /app

# Copy the pre-built binary file from the previous stage
COPY --from=builder --chown=appuser:appuser /app/main .

# Switch to non-root user
USER appuser

# Health check - auth używa /auth/health
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/auth/health || exit 1

# Command to run the executable
CMD ["./main"]

EXPOSE 8080