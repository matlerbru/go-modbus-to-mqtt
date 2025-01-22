# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o modbus-to-mqtt ./...

# Stage 2: Create a minimal runtime container
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /app/modbus-to-mqtt/modbus-to-mqtt /modbus-to-mqtt
COPY config.yaml /config.yaml
USER nonroot:nonroot
ENTRYPOINT ["/modbus-to-mqtt"]
