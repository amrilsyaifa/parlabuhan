# ------------------------------------------------------------
# Build Stage
# ------------------------------------------------------------
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Enable auto toolchain upgrade (fix go >= 1.25 requirement)
ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git

# Copy go modules first
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o parlabuhan .

# ------------------------------------------------------------
# Runtime Stage
# ------------------------------------------------------------
FROM alpine:3.19

WORKDIR /app

# We intentionally run as ROOT (needed to access /var/run/docker.sock)
# No USER app here

# Copy binary and templates
COPY --from=builder /app/parlabuhan /app/parlabuhan
COPY --from=builder /app/templates /app/templates

ENV PORT=8080
EXPOSE 8080

CMD ["/app/parlabuhan"]
