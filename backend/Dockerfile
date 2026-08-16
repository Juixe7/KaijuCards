# Build Stage
FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Run Stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/server .

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8080
ENTRYPOINT ["/app/server"]
