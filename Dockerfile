# Start from a Go base image
FROM golang:1.24-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o zeiterfassung_app .  

# --- Final stage for a small, efficient image ---
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built application from the build stage
COPY --from=build /app/zeiterfassung_app .

# Copy the public directory containing static files
COPY --from=build /app/public ./public

# Expose the application port
EXPOSE 8080

# Run the application when the container starts
CMD ["./zeiterfassung_app"]
