# Stage 1: Build the Go binary
FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS builder

RUN apk add --no-cache git ca-certificates

ARG TARGETOS TARGETARCH

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build with full static linking for minimal containers
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-s -w -extldflags=-static" \
    -o /hue-control ./cmd/hue-control/

# Stage 2: Minimal runtime image
FROM scratch

# Copy CA certificates for HTTPS calls to Hue Bridge
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /hue-control /hue-control

ENTRYPOINT ["/hue-control"]
CMD ["mcp"]