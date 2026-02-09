# Build stage
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application for multiple architectures
# CGO_ENABLED=0 is important for alpine to run the binary
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o server main.go

# Run stage
FROM --platform=$TARGETPLATFORM alpine:latest

WORKDIR /app

# Install certificates for HTTPS requests (Yahoo Finance)
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Expose port
EXPOSE 8000

# Run
CMD ["./server"]
