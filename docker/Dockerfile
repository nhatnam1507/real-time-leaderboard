FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install air for hot reload
RUN go install github.com/cosmtrek/air@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/bin/server .

# Copy air binary for hot reload (development)
COPY --from=builder /go/bin/air /usr/local/bin/air

EXPOSE 8080

CMD ["./server"]

