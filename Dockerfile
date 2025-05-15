# Build stage
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/golang:1.24.2-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./


RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/

# Download dependencies
RUN go mod download


# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

apk --no-cache add curl

# Expose port if needed
# EXPOSE 8080

# Set Gin to release mode
ENV GIN_MODE=release

# Run the application
CMD ["./main"] 