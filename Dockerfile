# Stage 1: Build the plugin
FROM golang:1.22.7-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the Go source code to the container
COPY . .

# Install musl-dev and gcc for cgo compilation
RUN apk add --no-cache musl-dev gcc 

# Enable CGO for building the plugin
ENV CGO_ENABLED=1

# Build the plugin using the same Go version as KrakenD, in plugin mode
RUN go build -buildmode=plugin -o krakend-validator-bypass.so

# Define the output as a stage to copy artifacts out
FROM scratch AS output

# Copy the built .so file to the output stage
COPY --from=builder /app/krakend-validator-bypass.so /krakend-validator-bypass.so


